package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/thumbtrend/backend/internal/store"
)

type ChannelHandler struct {
	store *store.ChannelStore
}

func NewChannelHandler(s *store.ChannelStore) *ChannelHandler {
	return &ChannelHandler{store: s}
}

func (h *ChannelHandler) GetTrending(w http.ResponseWriter, r *http.Request) {
	since := parseTimeRange(r)
	limit := parseIntParam(r, "limit", 50)
	channels, err := h.store.GetTrendingChannels(r.Context(), since, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch channels")
		return
	}
	writeJSON(w, http.StatusOK, channels)
}

func (h *ChannelHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	snapshots, err := h.store.GetHistory(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch history")
		return
	}
	writeJSON(w, http.StatusOK, snapshots)
}

func (h *ChannelHandler) BubbleData(w http.ResponseWriter, r *http.Request) {
	since := parseTimeRange(r)
	data, err := h.store.GetBubbleData(r.Context(), since)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch bubble data")
		return
	}
	writeJSON(w, http.StatusOK, data)
}
