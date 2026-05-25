package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/thumbtrend/backend/internal/store"
)

type VideoHandler struct {
	store *store.VideoStore
}

func NewVideoHandler(s *store.VideoStore) *VideoHandler {
	return &VideoHandler{store: s}
}

func (h *VideoHandler) GetTrending(w http.ResponseWriter, r *http.Request) {
	since := parseTimeRange(r)
	cat := parseCategoryParam(r)
	limit := parseIntParam(r, "limit", 50)

	videos, err := h.store.GetTrending(r.Context(), since, cat, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch videos")
		return
	}

	writeJSON(w, http.StatusOK, videos)
}

func (h *VideoHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	video, err := h.store.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "video not found")
		return
	}
	writeJSON(w, http.StatusOK, video)
}
