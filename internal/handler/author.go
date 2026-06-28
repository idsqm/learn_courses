package handler

import (
	"net/http"

	"github.com/andruho/courses/internal/domain"
	"github.com/andruho/courses/internal/service"
)

type AuthorHandler struct {
	authors service.AuthorService
}

func NewAuthorHandler(authors service.AuthorService) *AuthorHandler {
	return &AuthorHandler{authors: authors}
}

func (h *AuthorHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := intURLParam(r, "id")
	if err != nil {
		writeError(w, domain.ErrAuthorNotFound)
		return
	}

	author, err := h.authors.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	if author == nil {
		writeError(w, domain.ErrAuthorNotFound)
		return
	}
	writeOK(w, author)
}

func (h *AuthorHandler) Apply(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())

	if err := h.authors.Apply(r.Context(), userID); err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"message": "Application submitted"})
}
