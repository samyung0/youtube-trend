package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func parseTimeRange(r *http.Request) time.Time {
	rng := r.URL.Query().Get("range")
	switch rng {
	case "7d":
		return time.Now().Add(-7 * 24 * time.Hour)
	case "30d":
		return time.Now().Add(-30 * 24 * time.Hour)
	default:
		return time.Now().Add(-24 * time.Hour)
	}
}

func parseRange(r *http.Request) string {
	rng := r.URL.Query().Get("range")
	if rng == "7d" || rng == "30d" {
		return rng
	}
	return "24h"
}

func parseIntParam(r *http.Request, key string, fallback int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func parseCategoryParam(r *http.Request) *int {
	v := r.URL.Query().Get("category")
	if v == "" {
		return nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return nil
	}
	return &n
}
