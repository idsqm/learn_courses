package handler

import (
	"encoding/json"
	"net/http"

	"github.com/andruho/courses/internal/domain"
	"github.com/andruho/courses/internal/service"
)

type ReviewHandler struct {
	reviews service.ReviewService
}

func NewReviewHandler(reviews service.ReviewService) *ReviewHandler {
	return &ReviewHandler{reviews: reviews}
}

func (h *ReviewHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	courseID, err := intURLParam(r, "id")
	if err != nil {
		writeError(w, domain.ErrCourseNotFound)
		return
	}

	var req domain.CreateReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeDecodeError(w, err)
		return
	}

	ve := domain.NewValidationErrors()
	if req.Name == "" {
		ve.Add("name", "required")
	}
	if req.Initials == "" {
		ve.Add("initials", "required")
	}
	if req.Text == "" {
		ve.Add("text", "required")
	}
	if req.Rating < 1 || req.Rating > 5 {
		ve.Add("rating", "must be between 1 and 5")
	}
	if ve.HasErrors() {
		writeError(w, ve)
		return
	}

	if err := h.reviews.Create(r.Context(), userID, courseID, req.Name, req.Initials, req.Text, req.Rating); err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"message": "Review created"})
}
