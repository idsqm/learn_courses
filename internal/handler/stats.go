package handler

import (
	"net/http"

	"github.com/andruho/courses/internal/service"
)

type StatsHandler struct {
	progress service.ProgressService
}

func NewStatsHandler(progress service.ProgressService) *StatsHandler {
	return &StatsHandler{progress: progress}
}

func (h *StatsHandler) GetUserStats(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())

	stats, err := h.progress.GetUserStats(r.Context(), userID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeOK(w, stats)
}
