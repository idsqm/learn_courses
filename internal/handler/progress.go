package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/andruho/courses/internal/service"
)

type ProgressHandler struct {
	progress service.ProgressService
}

func NewProgressHandler(progress service.ProgressService) *ProgressHandler {
	return &ProgressHandler{progress: progress}
}

func (h *ProgressHandler) CompleteLesson(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	courseID := chi.URLParam(r, "id")
	lessonID := chi.URLParam(r, "lessonID")

	if err := h.progress.CompleteLesson(r.Context(), userID, courseID, lessonID); err != nil {
		writeError(w, err)
		return
	}

	writeOK(w, map[string]string{"message": "Lesson completed"})
}

func (h *ProgressHandler) GetCourseProgress(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	courseID := chi.URLParam(r, "id")

	progress, err := h.progress.GetCourseProgress(r.Context(), userID, courseID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeOK(w, progress)
}
