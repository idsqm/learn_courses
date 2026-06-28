package handler

import (
	"encoding/json"
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

	var req domain.ApplyAuthorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeDecodeError(w, err)
		return
	}

	ve := domain.NewValidationErrors()
	if req.Name == "" {
		ve.Add("name", "required")
	}
	if ve.HasErrors() {
		writeError(w, ve)
		return
	}

	if err := h.authors.Apply(r.Context(), userID, req); err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"message": "Author profile created"})
}
