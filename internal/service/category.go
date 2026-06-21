package service

import (
	"context"

	"github.com/andruho/courses/internal/domain"
	"github.com/andruho/courses/internal/repository"
)

type CategoryService interface {
	List(ctx context.Context) ([]domain.Category, error)
}

type categoryService struct {
	categories repository.CategoryRepository
}

func NewCategoryService(categories repository.CategoryRepository) CategoryService {
	return &categoryService{categories: categories}
}

func (s *categoryService) List(ctx context.Context) ([]domain.Category, error) {
	return s.categories.List(ctx)
}
