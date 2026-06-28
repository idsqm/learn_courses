package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/andruho/courses/internal/domain"
	"github.com/andruho/courses/internal/service"
)

type CourseHandler struct {
	courses service.CourseService
}

func NewCourseHandler(courses service.CourseService) *CourseHandler {
	return &CourseHandler{courses: courses}
}

func (h *CourseHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	f := domain.CourseFilter{
		Level:   q.Get("level"),
		Sort:    q.Get("sort"),
		Search:  q.Get("search"),
		Page:    intParam(q.Get("page"), 1),
		PerPage: intParam(q.Get("per_page"), 12),
	}

	if cats := q.Get("category"); cats != "" {
		f.Categories = strings.Split(cats, ",")
	}
	if v := q.Get("price_min"); v != "" {
		val, _ := strconv.ParseFloat(v, 64)
		f.PriceMin = &val
	}
	if v := q.Get("price_max"); v != "" {
		val, _ := strconv.ParseFloat(v, 64)
		f.PriceMax = &val
	}
	if v := q.Get("rating_min"); v != "" {
		val, _ := strconv.ParseFloat(v, 64)
		f.RatingMin = &val
	}

	courses, total, err := h.courses.List(r.Context(), f)
	if err != nil {
		writeError(w, err)
		return
	}

	totalPages := (total + f.PerPage - 1) / f.PerPage
	writeOK(w, map[string]any{
		"data": courses,
		"pagination": domain.Pagination{
			Page:       f.Page,
			PerPage:    f.PerPage,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

func (h *CourseHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := intURLParam(r, "id")
	if err != nil {
		writeError(w, domain.ErrCourseNotFound)
		return
	}

	course, err := h.courses.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	if course == nil {
		writeError(w, domain.ErrCourseNotFound)
		return
	}

	writeOK(w, course)
}

func (h *CourseHandler) Featured(w http.ResponseWriter, r *http.Request) {
	courses, err := h.courses.GetFeatured(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]any{"data": courses})
}

func (h *CourseHandler) Recommended(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())

	courses, err := h.courses.GetRecommended(r.Context(), userID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]any{"data": courses})
}

func intParam(s string, def int) int {
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil || v < 1 {
		return def
	}
	return v
}
