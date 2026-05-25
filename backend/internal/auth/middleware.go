package auth

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const UserIDKey contextKey = "user_id"

type ProChecker interface {
	IsPro(ctx context.Context, userID string) (bool, error)
}

func OptionalAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if token := extractBearer(r); token != "" {
				if claims, err := ValidateToken(token, secret); err == nil {
					ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
					r = r.WithContext(ctx)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RequireAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractBearer(r)
			if token == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			claims, err := ValidateToken(token, secret)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireProMiddleware(store ProChecker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := GetUserID(r.Context())
			if userID == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			isPro, err := store.IsPro(r.Context(), userID)
			if err != nil || !isPro {
				http.Error(w, "pro subscription required", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RangeGated(secret string, store ProChecker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rang := r.URL.Query().Get("range")
			if rang != "" && rang != "24h" {
				token := extractBearer(r)
				if token == "" {
					http.Error(w, "pro subscription required", http.StatusForbidden)
					return
				}

				claims, err := ValidateToken(token, secret)
				if err != nil {
					http.Error(w, "unauthorized", http.StatusUnauthorized)
					return
				}

				isPro, err := store.IsPro(r.Context(), claims.UserID)
				if err != nil || !isPro {
					http.Error(w, "pro subscription required", http.StatusForbidden)
					return
				}

				ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
				r = r.WithContext(ctx)
			}

			next.ServeHTTP(w, r)
		})
	}
}

func GetUserID(ctx context.Context) string {
	if v, ok := ctx.Value(UserIDKey).(string); ok {
		return v
	}
	return ""
}

func extractBearer(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if after, found := strings.CutPrefix(h, "Bearer "); found {
		return after
	}
	return ""
}
