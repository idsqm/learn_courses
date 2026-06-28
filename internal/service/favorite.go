package service

import (
	"context"

	"github.com/andruho/courses/internal/domain"
	"github.com/andruho/courses/internal/repository"
)

type FavoriteService interface {
	Add(ctx context.Context, userID string, courseID int) error
	Remove(ctx context.Context, userID string, courseID int) error
	ListByUser(ctx context.Context, userID string) ([]domain.CourseListItem, error)
}

type favoriteService struct {
	favorites repository.FavoriteRepository
	courses   repository.CourseRepository
}

func NewFavoriteService(favorites repository.FavoriteRepository, courses repository.CourseRepository) FavoriteService {
	return &favoriteService{favorites: favorites, courses: courses}
}

func (s *favoriteService) Add(ctx context.Context, userID string, courseID int) error {
	exists, err := s.courses.Exists(ctx, courseID)
	if err != nil {
		return err
	}
	if !exists {
		return domain.ErrCourseNotFound
	}
	return s.favorites.Add(ctx, userID, courseID)
}

func (s *favoriteService) Remove(ctx context.Context, userID string, courseID int) error {
	return s.favorites.Remove(ctx, userID, courseID)
}

func (s *favoriteService) ListByUser(ctx context.Context, userID string) ([]domain.CourseListItem, error) {
	return s.favorites.ListByUser(ctx, userID)
}
