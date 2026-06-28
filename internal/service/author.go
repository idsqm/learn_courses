package service

import (
	"context"

	"github.com/andruho/courses/internal/domain"
	"github.com/andruho/courses/internal/repository"
)

type AuthorService interface {
	GetByID(ctx context.Context, id int) (*domain.AuthorInfo, error)
	Apply(ctx context.Context, userID string) error
}

type authorService struct {
	authors repository.AuthorRepository
}

func NewAuthorService(authors repository.AuthorRepository) AuthorService {
	return &authorService{authors: authors}
}

func (s *authorService) GetByID(ctx context.Context, id int) (*domain.AuthorInfo, error) {
	return s.authors.GetByID(ctx, id)
}

func (s *authorService) Apply(ctx context.Context, userID string) error {
	return s.authors.Apply(ctx, userID)
}
