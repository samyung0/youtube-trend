package handler

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/thumbtrend/backend/internal/store"
)

type TopicHandler struct {
	store *store.TopicStore
}

func NewTopicHandler(s *store.TopicStore) *TopicHandler {
	return &TopicHandler{store: s}
}

func (h *TopicHandler) List(w http.ResponseWriter, r *http.Request) {
	since := parseTimeRange(r)
	topics, err := h.store.GetTopics(r.Context(), since)
	if err != nil {
		log.Printf("topics list: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to fetch topics")
		return
	}
	writeJSON(w, http.StatusOK, topics)
}

func (h *TopicHandler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	topic, err := h.store.GetBySlug(r.Context(), slug)
	if err != nil {
		writeError(w, http.StatusNotFound, "topic not found")
		return
	}

	limit := parseIntParam(r, "limit", 50)
	videos, err := h.store.GetVideosByTopic(r.Context(), topic.ID, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch videos")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"topic":  topic,
		"videos": videos,
	})
}

func (h *TopicHandler) BubbleData(w http.ResponseWriter, r *http.Request) {
	since := parseTimeRange(r)
	data, err := h.store.GetBubbleData(r.Context(), since)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch bubble data")
		return
	}
	writeJSON(w, http.StatusOK, data)
}
