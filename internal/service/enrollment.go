package service

import (
	"context"

	"github.com/andruho/courses/internal/domain"
	"github.com/andruho/courses/internal/repository"
)

type EnrollmentService interface {
	Enroll(ctx context.Context, userID string, courseID int) error
	ListByUser(ctx context.Context, userID string) ([]domain.EnrolledCourse, error)
}

type enrollmentService struct {
	enrollments repository.EnrollmentRepository
	courses     repository.CourseRepository
}

func NewEnrollmentService(enrollments repository.EnrollmentRepository, courses repository.CourseRepository) EnrollmentService {
	return &enrollmentService{enrollments: enrollments, courses: courses}
}

func (s *enrollmentService) Enroll(ctx context.Context, userID string, courseID int) error {
	exists, err := s.courses.Exists(ctx, courseID)
	if err != nil {
		return err
	}
	if !exists {
		return domain.ErrCourseNotFound
	}
	return s.enrollments.Enroll(ctx, userID, courseID)
}

func (s *enrollmentService) ListByUser(ctx context.Context, userID string) ([]domain.EnrolledCourse, error) {
	return s.enrollments.ListByUser(ctx, userID)
}
