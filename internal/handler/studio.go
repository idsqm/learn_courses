package handler

import (
	"encoding/json"
	"net/http"

	"github.com/andruho/courses/internal/domain"
	"github.com/andruho/courses/internal/service"
)

type StudioHandler struct {
	studio service.StudioService
}

func NewStudioHandler(studio service.StudioService) *StudioHandler {
	return &StudioHandler{studio: studio}
}

// --- Courses ---

func (h *StudioHandler) ListCourses(w http.ResponseWriter, r *http.Request) {
	authorID := AuthorIDFromContext(r.Context())
	courses, err := h.studio.ListCourses(r.Context(), authorID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]any{"data": courses})
}

func (h *StudioHandler) GetCourse(w http.ResponseWriter, r *http.Request) {
	authorID := AuthorIDFromContext(r.Context())
	courseID, err := intURLParam(r, "id")
	if err != nil {
		writeError(w, domain.ErrCourseNotFound)
		return
	}
	course, err := h.studio.GetCourse(r.Context(), authorID, courseID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, course)
}

func (h *StudioHandler) CreateCourse(w http.ResponseWriter, r *http.Request) {
	authorID := AuthorIDFromContext(r.Context())

	var req domain.CreateCourseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, domain.ErrValidation)
		return
	}

	ve := domain.NewValidationErrors()
	if req.Title == "" {
		ve.Add("title", "required")
	}
	if req.CategoryID == 0 {
		ve.Add("category_id", "required")
	}
	if ve.HasErrors() {
		writeError(w, ve)
		return
	}

	if req.Level == "" {
		req.Level = "Любой"
	}

	id, err := h.studio.CreateCourse(r.Context(), authorID, req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"id": id})
}

func (h *StudioHandler) UpdateCourse(w http.ResponseWriter, r *http.Request) {
	authorID := AuthorIDFromContext(r.Context())
	courseID, err := intURLParam(r, "id")
	if err != nil {
		writeError(w, domain.ErrCourseNotFound)
		return
	}

	var req domain.UpdateCourseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, domain.ErrValidation)
		return
	}

	if err := h.studio.UpdateCourse(r.Context(), authorID, courseID, req); err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]string{"message": "Course updated"})
}

func (h *StudioHandler) DeleteCourse(w http.ResponseWriter, r *http.Request) {
	authorID := AuthorIDFromContext(r.Context())
	courseID, err := intURLParam(r, "id")
	if err != nil {
		writeError(w, domain.ErrCourseNotFound)
		return
	}

	if err := h.studio.DeleteCourse(r.Context(), authorID, courseID); err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]string{"message": "Course deleted"})
}

func (h *StudioHandler) PublishCourse(w http.ResponseWriter, r *http.Request) {
	authorID := AuthorIDFromContext(r.Context())
	courseID, err := intURLParam(r, "id")
	if err != nil {
		writeError(w, domain.ErrCourseNotFound)
		return
	}

	if err := h.studio.PublishCourse(r.Context(), authorID, courseID); err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]string{"message": "Course published"})
}

func (h *StudioHandler) UnpublishCourse(w http.ResponseWriter, r *http.Request) {
	authorID := AuthorIDFromContext(r.Context())
	courseID, err := intURLParam(r, "id")
	if err != nil {
		writeError(w, domain.ErrCourseNotFound)
		return
	}

	if err := h.studio.UnpublishCourse(r.Context(), authorID, courseID); err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]string{"message": "Course unpublished"})
}

// --- Modules ---

func (h *StudioHandler) CreateModule(w http.ResponseWriter, r *http.Request) {
	authorID := AuthorIDFromContext(r.Context())
	courseID, err := intURLParam(r, "id")
	if err != nil {
		writeError(w, domain.ErrCourseNotFound)
		return
	}

	var req domain.CreateModuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, domain.ErrValidation)
		return
	}
	if req.Title == "" {
		ve := domain.NewValidationErrors()
		ve.Add("title", "required")
		writeError(w, ve)
		return
	}

	id, err := h.studio.CreateModule(r.Context(), authorID, courseID, req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"id": id})
}

func (h *StudioHandler) UpdateModule(w http.ResponseWriter, r *http.Request) {
	authorID := AuthorIDFromContext(r.Context())
	courseID, err := intURLParam(r, "id")
	if err != nil {
		writeError(w, domain.ErrCourseNotFound)
		return
	}
	moduleID, err := intURLParam(r, "moduleId")
	if err != nil {
		writeError(w, domain.ErrModuleNotFound)
		return
	}

	var req domain.UpdateModuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, domain.ErrValidation)
		return
	}

	if err := h.studio.UpdateModule(r.Context(), authorID, courseID, moduleID, req); err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]string{"message": "Module updated"})
}

func (h *StudioHandler) DeleteModule(w http.ResponseWriter, r *http.Request) {
	authorID := AuthorIDFromContext(r.Context())
	courseID, err := intURLParam(r, "id")
	if err != nil {
		writeError(w, domain.ErrCourseNotFound)
		return
	}
	moduleID, err := intURLParam(r, "moduleId")
	if err != nil {
		writeError(w, domain.ErrModuleNotFound)
		return
	}

	if err := h.studio.DeleteModule(r.Context(), authorID, courseID, moduleID); err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]string{"message": "Module deleted"})
}

// --- Lessons ---

func (h *StudioHandler) CreateLesson(w http.ResponseWriter, r *http.Request) {
	authorID := AuthorIDFromContext(r.Context())
	courseID, err := intURLParam(r, "id")
	if err != nil {
		writeError(w, domain.ErrCourseNotFound)
		return
	}
	moduleID, err := intURLParam(r, "moduleId")
	if err != nil {
		writeError(w, domain.ErrModuleNotFound)
		return
	}

	var req domain.CreateLessonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, domain.ErrValidation)
		return
	}
	if req.Name == "" {
		ve := domain.NewValidationErrors()
		ve.Add("name", "required")
		writeError(w, ve)
		return
	}

	id, err := h.studio.CreateLesson(r.Context(), authorID, courseID, moduleID, req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"id": id})
}

func (h *StudioHandler) UpdateLesson(w http.ResponseWriter, r *http.Request) {
	authorID := AuthorIDFromContext(r.Context())
	courseID, err := intURLParam(r, "id")
	if err != nil {
		writeError(w, domain.ErrCourseNotFound)
		return
	}
	lessonID, err := intURLParam(r, "lessonId")
	if err != nil {
		writeError(w, domain.ErrLessonNotFound)
		return
	}

	var req domain.UpdateLessonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, domain.ErrValidation)
		return
	}

	if err := h.studio.UpdateLesson(r.Context(), authorID, courseID, lessonID, req); err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]string{"message": "Lesson updated"})
}

func (h *StudioHandler) DeleteLesson(w http.ResponseWriter, r *http.Request) {
	authorID := AuthorIDFromContext(r.Context())
	courseID, err := intURLParam(r, "id")
	if err != nil {
		writeError(w, domain.ErrCourseNotFound)
		return
	}
	lessonID, err := intURLParam(r, "lessonId")
	if err != nil {
		writeError(w, domain.ErrLessonNotFound)
		return
	}

	if err := h.studio.DeleteLesson(r.Context(), authorID, courseID, lessonID); err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]string{"message": "Lesson deleted"})
}

// --- Stats / Students / Income ---

func (h *StudioHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	authorID := AuthorIDFromContext(r.Context())
	stats, err := h.studio.GetStats(r.Context(), authorID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, stats)
}

func (h *StudioHandler) ListStudents(w http.ResponseWriter, r *http.Request) {
	authorID := AuthorIDFromContext(r.Context())
	students, err := h.studio.ListStudents(r.Context(), authorID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]any{"data": students})
}

func (h *StudioHandler) GetIncome(w http.ResponseWriter, r *http.Request) {
	authorID := AuthorIDFromContext(r.Context())
	income, err := h.studio.GetIncome(r.Context(), authorID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, income)
}

func (h *StudioHandler) ListPayouts(w http.ResponseWriter, r *http.Request) {
	authorID := AuthorIDFromContext(r.Context())
	payouts, err := h.studio.ListPayouts(r.Context(), authorID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]any{"data": payouts})
}

// --- Reviews ---

func (h *StudioHandler) ListReviews(w http.ResponseWriter, r *http.Request) {
	authorID := AuthorIDFromContext(r.Context())
	reviews, err := h.studio.ListReviews(r.Context(), authorID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]any{"data": reviews})
}

func (h *StudioHandler) ReplyToReview(w http.ResponseWriter, r *http.Request) {
	authorID := AuthorIDFromContext(r.Context())
	reviewID, err := intURLParam(r, "id")
	if err != nil {
		writeError(w, domain.ErrReviewNotFound)
		return
	}

	var req struct {
		Reply string `json:"reply"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, domain.ErrValidation)
		return
	}
	if req.Reply == "" {
		ve := domain.NewValidationErrors()
		ve.Add("reply", "required")
		writeError(w, ve)
		return
	}

	if err := h.studio.ReplyToReview(r.Context(), authorID, reviewID, req.Reply); err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]string{"message": "Reply added"})
}
