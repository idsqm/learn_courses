package handler

import (
	"net/http"

	"github.com/andruho/courses/internal/domain"
	"github.com/andruho/courses/internal/service"
)

type EnrollmentHandler struct {
	enrollments service.EnrollmentService
}

func NewEnrollmentHandler(enrollments service.EnrollmentService) *EnrollmentHandler {
	return &EnrollmentHandler{enrollments: enrollments}
}

func (h *EnrollmentHandler) Enroll(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	courseID, err := intURLParam(r, "id")
	if err != nil {
		writeError(w, domain.ErrCourseNotFound)
		return
	}

	if err := h.enrollments.Enroll(r.Context(), userID, courseID); err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"message": "Successfully enrolled"})
}

func (h *EnrollmentHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())

	enrollments, err := h.enrollments.ListByUser(r.Context(), userID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]any{"data": enrollments})
}
