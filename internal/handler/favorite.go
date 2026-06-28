package handler

import (
	"net/http"

	"github.com/andruho/courses/internal/domain"
	"github.com/andruho/courses/internal/service"
)

type FavoriteHandler struct {
	favorites service.FavoriteService
}

func NewFavoriteHandler(favorites service.FavoriteService) *FavoriteHandler {
	return &FavoriteHandler{favorites: favorites}
}

func (h *FavoriteHandler) Add(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	courseID, err := intURLParam(r, "id")
	if err != nil {
		writeError(w, domain.ErrCourseNotFound)
		return
	}

	if err := h.favorites.Add(r.Context(), userID, courseID); err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"message": "Added to favorites"})
}

func (h *FavoriteHandler) Remove(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	courseID, err := intURLParam(r, "id")
	if err != nil {
		writeError(w, domain.ErrCourseNotFound)
		return
	}

	if err := h.favorites.Remove(r.Context(), userID, courseID); err != nil {
		writeError(w, err)
		return
	}

	writeOK(w, map[string]string{"message": "Removed from favorites"})
}

func (h *FavoriteHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())

	favorites, err := h.favorites.ListByUser(r.Context(), userID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]any{"data": favorites})
}
