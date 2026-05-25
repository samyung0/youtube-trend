package handler

import (
	"net/http"

	"github.com/thumbtrend/backend/internal/store"
)

type AnalysisHandler struct {
	store *store.AnalysisStore
}

func NewAnalysisHandler(s *store.AnalysisStore) *AnalysisHandler {
	return &AnalysisHandler{store: s}
}

func (h *AnalysisHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	since := parseTimeRange(r)
	stats, err := h.store.GetStats(r.Context(), since)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch analysis stats")
		return
	}
	writeJSON(w, http.StatusOK, stats)
}
