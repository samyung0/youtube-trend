package main

import (
	"context"
	"log"
	"os"

	"github.com/thumbtrend/backend/internal/config"
	"github.com/thumbtrend/backend/internal/scraper"
	"github.com/thumbtrend/backend/internal/store"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: scraper <fetch|topics|analyze|channels|all>")
	}
	cmd := os.Args[1]

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("config:", err)
	}

	pool, err := store.NewPool(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("db:", err)
	}
	defer pool.Close()

	ctx := context.Background()
	vs := store.NewVideoStore(pool)
	ts := store.NewTopicStore(pool)
	cs := store.NewChannelStore(pool)
	// as := store.NewAnalysisStore(pool)
	yt := scraper.NewYouTubeClient(cfg.YouTubeAPIKey)

	switch cmd {
	case "fetch":
		if err := scraper.RunFetchTrending(ctx, yt, vs, cs); err != nil {
			log.Fatal(err)
		}
	case "topics":
		if err := scraper.RunClusterTopics(ctx, cfg.DeepSeekAPIKey, vs, ts); err != nil {
			log.Fatal(err)
		}
	case "analyze":
		// if err := scraper.RunAnalyzeThumbnails(ctx, vs, as); err != nil {
		// 	log.Fatal(err)
		// }
	case "channels":
		if err := scraper.RunTrackChannels(ctx, yt, cs); err != nil {
			log.Fatal(err)
		}
	case "all":
		scraper.RunFetchTrending(ctx, yt, vs, cs)
		scraper.RunClusterTopics(ctx, cfg.DeepSeekAPIKey, vs, ts)
		// scraper.RunAnalyzeThumbnails(ctx, vs, as)
		scraper.RunTrackChannels(ctx, yt, cs)
	default:
		log.Fatalf("Unknown command: %s. Use: fetch, topics, channels, all", cmd)
	}
}
