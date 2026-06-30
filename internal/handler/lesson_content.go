package handler

import (
	"encoding/json"
	"net/http"

	"github.com/andruho/courses/internal/domain"
	"github.com/andruho/courses/internal/service"
)

type LessonContentHandler struct {
	content service.LessonContentService
}

func NewLessonContentHandler(content service.LessonContentService) *LessonContentHandler {
	return &LessonContentHandler{content: content}
}

func (h *LessonContentHandler) GetLessonDetail(w http.ResponseWriter, r *http.Request) {
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

	detail, err := h.content.GetLessonDetail(r.Context(), authorID, courseID, lessonID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, detail)
}

func (h *LessonContentHandler) SaveContent(w http.ResponseWriter, r *http.Request) {
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

	var req domain.SaveLessonContentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeDecodeError(w, err)
		return
	}

	if err := h.content.SaveContent(r.Context(), authorID, courseID, lessonID, req); err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]string{"message": "Content saved"})
}

func (h *LessonContentHandler) CreateQuestion(w http.ResponseWriter, r *http.Request) {
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

	var req domain.SaveQuizQuestionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeDecodeError(w, err)
		return
	}

	ve := domain.NewValidationErrors()
	if req.Text == "" {
		ve.Add("text", "required")
	}
	if ve.HasErrors() {
		writeError(w, ve)
		return
	}

	id, err := h.content.CreateQuestion(r.Context(), authorID, courseID, lessonID, req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"id": id})
}

func (h *LessonContentHandler) UpdateQuestion(w http.ResponseWriter, r *http.Request) {
	authorID := AuthorIDFromContext(r.Context())
	courseID, err := intURLParam(r, "id")
	if err != nil {
		writeError(w, domain.ErrCourseNotFound)
		return
	}
	questionID, err := intURLParam(r, "questionId")
	if err != nil {
		writeError(w, domain.ErrQuestionNotFound)
		return
	}

	var req domain.SaveQuizQuestionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeDecodeError(w, err)
		return
	}

	ve := domain.NewValidationErrors()
	if req.Text == "" {
		ve.Add("text", "required")
	}
	if ve.HasErrors() {
		writeError(w, ve)
		return
	}

	if err := h.content.UpdateQuestion(r.Context(), authorID, courseID, questionID, req); err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]string{"message": "Question updated"})
}

func (h *LessonContentHandler) DeleteQuestion(w http.ResponseWriter, r *http.Request) {
	authorID := AuthorIDFromContext(r.Context())
	courseID, err := intURLParam(r, "id")
	if err != nil {
		writeError(w, domain.ErrCourseNotFound)
		return
	}
	questionID, err := intURLParam(r, "questionId")
	if err != nil {
		writeError(w, domain.ErrQuestionNotFound)
		return
	}

	if err := h.content.DeleteQuestion(r.Context(), authorID, courseID, questionID); err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]string{"message": "Question deleted"})
}
