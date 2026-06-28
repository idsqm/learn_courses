package service

import (
	"context"

	"github.com/andruho/courses/internal/domain"
	"github.com/andruho/courses/internal/repository"
)

type ProgressService interface {
	CompleteLesson(ctx context.Context, userID string, courseID, lessonID int) error
	GetCourseProgress(ctx context.Context, userID string, courseID int) (*domain.CourseProgress, error)
	GetUserStats(ctx context.Context, userID string) (*domain.UserStats, error)
}

type progressService struct {
	progress     repository.ProgressRepository
	certificates repository.CertificateRepository
	enrollments  repository.EnrollmentRepository
}

func NewProgressService(
	progress repository.ProgressRepository,
	certificates repository.CertificateRepository,
	enrollments repository.EnrollmentRepository,
) ProgressService {
	return &progressService{
		progress:     progress,
		certificates: certificates,
		enrollments:  enrollments,
	}
}

func (s *progressService) CompleteLesson(ctx context.Context, userID string, courseID, lessonID int) error {
	enrolled, err := s.enrollments.IsEnrolled(ctx, userID, courseID)
	if err != nil {
		return err
	}
	if !enrolled {
		return domain.ErrNotEnrolled
	}

	exists, err := s.progress.LessonExists(ctx, courseID, lessonID)
	if err != nil {
		return err
	}
	if !exists {
		return domain.ErrLessonNotFound
	}

	if err := s.progress.CompleteLesson(ctx, userID, courseID, lessonID); err != nil {
		return err
	}

	completed, err := s.progress.IsFullyCompleted(ctx, userID, courseID)
	if err != nil {
		return err
	}
	if completed {
		_ = s.certificates.Issue(ctx, userID, courseID)
	}

	return nil
}

func (s *progressService) GetCourseProgress(ctx context.Context, userID string, courseID int) (*domain.CourseProgress, error) {
	return s.progress.GetCourseProgress(ctx, userID, courseID)
}

func (s *progressService) GetUserStats(ctx context.Context, userID string) (*domain.UserStats, error) {
	return s.progress.GetUserStats(ctx, userID)
}
