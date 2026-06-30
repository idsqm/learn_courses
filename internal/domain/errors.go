package domain

import "errors"

type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
}

func (e *AppError) Error() string {
	return e.Message
}

func NewAppError(code, message string, httpStatus int) *AppError {
	return &AppError{Code: code, Message: message, HTTPStatus: httpStatus}
}

func IsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}

// Course
var (
	ErrCourseNotFound = NewAppError("COURSE_NOT_FOUND", "Course not found", 404)
)

// Enrollment
var (
	ErrAlreadyEnrolled = NewAppError("ALREADY_ENROLLED", "You are already enrolled in this course", 409)
	ErrNotEnrolled     = NewAppError("NOT_ENROLLED", "You are not enrolled in this course", 403)
)

// Favorites
var (
	ErrAlreadyFavorited = NewAppError("ALREADY_FAVORITED", "Course is already in favorites", 409)
	ErrNotFavorited     = NewAppError("NOT_FAVORITED", "Course is not in favorites", 404)
)

// Progress
var (
	ErrLessonNotFound   = NewAppError("LESSON_NOT_FOUND", "Lesson not found", 404)
	ErrAlreadyCompleted = NewAppError("ALREADY_COMPLETED", "Lesson is already completed", 409)
)

// Reviews
var (
	ErrReviewAlreadyExists = NewAppError("REVIEW_ALREADY_EXISTS", "You have already reviewed this course", 409)
)

// Certificates
var (
	ErrCertificateNotFound = NewAppError("CERTIFICATE_NOT_FOUND", "Certificate not found", 404)
)

// Authors
var (
	ErrAuthorNotFound = NewAppError("AUTHOR_NOT_FOUND", "Author not found", 404)
	ErrAlreadyApplied = NewAppError("ALREADY_APPLIED", "You have already applied to become an author", 409)
	ErrNotAuthor      = NewAppError("NOT_AUTHOR", "You are not an approved author", 403)
	ErrCourseNotOwned = NewAppError("COURSE_NOT_OWNED", "You don't own this course", 403)
)

// Studio
var (
	ErrModuleNotFound   = NewAppError("MODULE_NOT_FOUND", "Module not found", 404)
	ErrReviewNotFound   = NewAppError("REVIEW_NOT_FOUND", "Review not found", 404)
	ErrQuestionNotFound = NewAppError("QUESTION_NOT_FOUND", "Question not found", 404)
)

// Auth
var (
	ErrAccessTokenExpired = NewAppError("ACCESS_TOKEN_EXPIRED", "Access token has expired", 401)
	ErrAccessTokenInvalid = NewAppError("ACCESS_TOKEN_INVALID", "Access token is invalid", 401)
)

// Generic
var (
	ErrInternal   = NewAppError("INTERNAL_ERROR", "Internal server error", 500)
	ErrValidation = NewAppError("VALIDATION_ERROR", "Validation error", 422)
)

// Validation errors
type ValidationErrors struct {
	Fields map[string][]string
}

func NewValidationErrors() *ValidationErrors {
	return &ValidationErrors{Fields: make(map[string][]string)}
}

func (e *ValidationErrors) Add(field, code string) {
	e.Fields[field] = append(e.Fields[field], code)
}

func (e *ValidationErrors) HasErrors() bool {
	return len(e.Fields) > 0
}

func (e *ValidationErrors) Error() string {
	return "validation error"
}
