package service

import (
	"context"

	"github.com/andruho/courses/internal/domain"
	"github.com/andruho/courses/internal/repository"
)

type ReviewService interface {
	Create(ctx context.Context, userID string, courseID int, name, initials, text string, rating int) error
}

type reviewService struct {
	reviews repository.ReviewRepository
	courses repository.CourseRepository
}

func NewReviewService(reviews repository.ReviewRepository, courses repository.CourseRepository) ReviewService {
	return &reviewService{reviews: reviews, courses: courses}
}

func (s *reviewService) Create(ctx context.Context, userID string, courseID int, name, initials, text string, rating int) error {
	exists, err := s.courses.Exists(ctx, courseID)
	if err != nil {
		return err
	}
	if !exists {
		return domain.ErrCourseNotFound
	}
	return s.reviews.Create(ctx, userID, courseID, name, initials, text, rating)
}
