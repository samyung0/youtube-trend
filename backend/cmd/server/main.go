package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/thumbtrend/backend/internal/auth"
	"github.com/thumbtrend/backend/internal/config"
	"github.com/thumbtrend/backend/internal/handler"
	"github.com/thumbtrend/backend/internal/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("config error:", err)
	}

	pool, err := store.NewPool(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("db error:", err)
	}
	defer pool.Close()

	userStore := store.NewUserStore(pool)
	videoStore := store.NewVideoStore(pool)
	topicStore := store.NewTopicStore(pool)
	channelStore := store.NewChannelStore(pool)
	analysisStore := store.NewAnalysisStore(pool)

	oauthCfg := auth.NewGoogleOAuthConfig(cfg.GoogleClientID, cfg.GoogleSecret, cfg.GoogleRedirect)

	authHandler := handler.NewAuthHandler(oauthCfg, userStore, cfg.JWTSecret, cfg.FrontendURL)
	videoHandler := handler.NewVideoHandler(videoStore)
	topicHandler := handler.NewTopicHandler(topicStore)
	channelHandler := handler.NewChannelHandler(channelStore)
	analysisHandler := handler.NewAnalysisHandler(analysisStore)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.FrontendURL},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Route("/api", func(r chi.Router) {
		// Auth - public
		r.Get("/auth/google", authHandler.GoogleLogin)
		r.Get("/auth/callback", authHandler.GoogleCallback)
		r.With(auth.OptionalAuth(cfg.JWTSecret)).Get("/auth/me", authHandler.Me)

		// Videos
		r.Group(func(r chi.Router) {
			r.Use(auth.RangeGated(cfg.JWTSecret, userStore))
			r.Get("/videos/trending", videoHandler.GetTrending)
		})
		r.Get("/videos/{id}", videoHandler.GetByID)

		// Topics
		r.Group(func(r chi.Router) {
			r.Use(auth.RangeGated(cfg.JWTSecret, userStore))
			r.Get("/topics", topicHandler.List)
			r.Get("/topics/bubble", topicHandler.BubbleData)
		})
		r.Get("/topics/{slug}", topicHandler.GetBySlug)

		// Channels
		r.Group(func(r chi.Router) {
			r.Use(auth.RangeGated(cfg.JWTSecret, userStore))
			r.Get("/channels/trending", channelHandler.GetTrending)
			r.Get("/channels/bubble", channelHandler.BubbleData)
		})
		r.With(auth.RequireAuth(cfg.JWTSecret), auth.RequireProMiddleware(userStore)).
			Get("/channels/{id}/history", channelHandler.GetHistory)

		// Analysis
		r.Group(func(r chi.Router) {
			r.Use(auth.RangeGated(cfg.JWTSecret, userStore))
			r.Get("/analysis/stats", analysisHandler.GetStats)
		})
	})

	srv := &http.Server{Addr: ":" + cfg.Port, Handler: r}

	go func() {
		log.Printf("Server starting on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	log.Println("Server stopped")
}
