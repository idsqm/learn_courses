package service

import (
	"context"

	"github.com/andruho/courses/internal/domain"
	"github.com/andruho/courses/internal/repository"
)

type StudioService interface {
	ResolveAuthor(ctx context.Context, userID string) (int, error)

	ListCourses(ctx context.Context, authorID int) ([]domain.StudioCourse, error)
	CreateCourse(ctx context.Context, authorID int, req domain.CreateCourseRequest) (int, error)
	UpdateCourse(ctx context.Context, authorID, courseID int, req domain.UpdateCourseRequest) error
	DeleteCourse(ctx context.Context, authorID, courseID int) error
	PublishCourse(ctx context.Context, authorID, courseID int) error
	UnpublishCourse(ctx context.Context, authorID, courseID int) error

	CreateModule(ctx context.Context, authorID, courseID int, req domain.CreateModuleRequest) (int, error)
	UpdateModule(ctx context.Context, authorID, courseID, moduleID int, req domain.UpdateModuleRequest) error
	DeleteModule(ctx context.Context, authorID, courseID, moduleID int) error

	CreateLesson(ctx context.Context, authorID, courseID, moduleID int, req domain.CreateLessonRequest) (int, error)
	UpdateLesson(ctx context.Context, authorID, courseID, lessonID int, req domain.UpdateLessonRequest) error
	DeleteLesson(ctx context.Context, authorID, courseID, lessonID int) error

	GetStats(ctx context.Context, authorID int) (*domain.StudioStats, error)
	ListStudents(ctx context.Context, authorID int) ([]domain.StudioStudent, error)
	GetIncome(ctx context.Context, authorID int) (*domain.StudioIncome, error)
	ListPayouts(ctx context.Context, authorID int) ([]domain.Payout, error)
	ListReviews(ctx context.Context, authorID int) ([]domain.StudioReview, error)
	ReplyToReview(ctx context.Context, authorID, reviewID int, reply string) error
}

type studioService struct {
	studio repository.StudioRepository
}

func NewStudioService(studio repository.StudioRepository) StudioService {
	return &studioService{studio: studio}
}

func (s *studioService) ResolveAuthor(ctx context.Context, userID string) (int, error) {
	return s.studio.GetAuthorByUserID(ctx, userID)
}

func (s *studioService) ListCourses(ctx context.Context, authorID int) ([]domain.StudioCourse, error) {
	return s.studio.ListCourses(ctx, authorID)
}

func (s *studioService) CreateCourse(ctx context.Context, authorID int, req domain.CreateCourseRequest) (int, error) {
	return s.studio.CreateCourse(ctx, authorID, req)
}

func (s *studioService) UpdateCourse(ctx context.Context, authorID, courseID int, req domain.UpdateCourseRequest) error {
	return s.studio.UpdateCourse(ctx, authorID, courseID, req)
}

func (s *studioService) DeleteCourse(ctx context.Context, authorID, courseID int) error {
	return s.studio.DeleteCourse(ctx, authorID, courseID)
}

func (s *studioService) PublishCourse(ctx context.Context, authorID, courseID int) error {
	return s.studio.PublishCourse(ctx, authorID, courseID)
}

func (s *studioService) UnpublishCourse(ctx context.Context, authorID, courseID int) error {
	return s.studio.UnpublishCourse(ctx, authorID, courseID)
}

func (s *studioService) CreateModule(ctx context.Context, authorID, courseID int, req domain.CreateModuleRequest) (int, error) {
	return s.studio.CreateModule(ctx, authorID, courseID, req)
}

func (s *studioService) UpdateModule(ctx context.Context, authorID, courseID, moduleID int, req domain.UpdateModuleRequest) error {
	return s.studio.UpdateModule(ctx, authorID, courseID, moduleID, req)
}

func (s *studioService) DeleteModule(ctx context.Context, authorID, courseID, moduleID int) error {
	return s.studio.DeleteModule(ctx, authorID, courseID, moduleID)
}

func (s *studioService) CreateLesson(ctx context.Context, authorID, courseID, moduleID int, req domain.CreateLessonRequest) (int, error) {
	return s.studio.CreateLesson(ctx, authorID, courseID, moduleID, req)
}

func (s *studioService) UpdateLesson(ctx context.Context, authorID, courseID, lessonID int, req domain.UpdateLessonRequest) error {
	return s.studio.UpdateLesson(ctx, authorID, courseID, lessonID, req)
}

func (s *studioService) DeleteLesson(ctx context.Context, authorID, courseID, lessonID int) error {
	return s.studio.DeleteLesson(ctx, authorID, courseID, lessonID)
}

func (s *studioService) GetStats(ctx context.Context, authorID int) (*domain.StudioStats, error) {
	return s.studio.GetStats(ctx, authorID)
}

func (s *studioService) ListStudents(ctx context.Context, authorID int) ([]domain.StudioStudent, error) {
	return s.studio.ListStudents(ctx, authorID)
}

func (s *studioService) GetIncome(ctx context.Context, authorID int) (*domain.StudioIncome, error) {
	return s.studio.GetIncome(ctx, authorID)
}

func (s *studioService) ListPayouts(ctx context.Context, authorID int) ([]domain.Payout, error) {
	return s.studio.ListPayouts(ctx, authorID)
}

func (s *studioService) ListReviews(ctx context.Context, authorID int) ([]domain.StudioReview, error) {
	return s.studio.ListReviews(ctx, authorID)
}

func (s *studioService) ReplyToReview(ctx context.Context, authorID, reviewID int, reply string) error {
	return s.studio.ReplyToReview(ctx, authorID, reviewID, reply)
}
