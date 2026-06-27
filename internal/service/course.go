package service

import (
	"context"

	"github.com/andruho/courses/internal/domain"
	"github.com/andruho/courses/internal/repository"
)

type CourseService interface {
	List(ctx context.Context, f domain.CourseFilter) ([]domain.CourseListItem, int, error)
	GetByID(ctx context.Context, id string) (*domain.CourseDetail, error)
	GetFeatured(ctx context.Context) ([]domain.CourseListItem, error)
	GetRecommended(ctx context.Context, userID string) ([]domain.CourseListItem, error)
}

type courseService struct {
	courses repository.CourseRepository
}

func NewCourseService(courses repository.CourseRepository) CourseService {
	return &courseService{courses: courses}
}

func (s *courseService) List(ctx context.Context, f domain.CourseFilter) ([]domain.CourseListItem, int, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PerPage < 1 || f.PerPage > 50 {
		f.PerPage = 12
	}
	return s.courses.List(ctx, f)
}

func (s *courseService) GetByID(ctx context.Context, id string) (*domain.CourseDetail, error) {
	return s.courses.GetByID(ctx, id)
}

func (s *courseService) GetFeatured(ctx context.Context) ([]domain.CourseListItem, error) {
	return s.courses.GetFeatured(ctx, 8)
}

func (s *courseService) GetRecommended(ctx context.Context, userID string) ([]domain.CourseListItem, error) {
	return s.courses.GetRecommended(ctx, userID, 4)
}
