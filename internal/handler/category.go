package handler

import (
	"net/http"

	"github.com/andruho/courses/internal/service"
)

type CategoryHandler struct {
	categories service.CategoryService
}

func NewCategoryHandler(categories service.CategoryService) *CategoryHandler {
	return &CategoryHandler{categories: categories}
}

func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	categories, err := h.categories.List(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]any{"data": categories})
}
