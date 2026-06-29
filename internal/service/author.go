package service

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/andruho/courses/internal/domain"
	"github.com/andruho/courses/internal/repository"
)

type AuthorService interface {
	GetByID(ctx context.Context, id int) (*domain.AuthorInfo, error)
	Apply(ctx context.Context, userID string, req domain.ApplyAuthorRequest) error
}

type authorService struct {
	authors repository.AuthorRepository
	authURL string
}

func NewAuthorService(authors repository.AuthorRepository, authURL string) AuthorService {
	return &authorService{authors: authors, authURL: authURL}
}

func (s *authorService) GetByID(ctx context.Context, id int) (*domain.AuthorInfo, error) {
	return s.authors.GetByID(ctx, id)
}

func (s *authorService) Apply(ctx context.Context, userID string, req domain.ApplyAuthorRequest) error {
	if err := s.updateRole(ctx, userID); err != nil {
		return err
	}

	if err := s.authors.Apply(ctx, userID, req); err != nil {
		return err
	}

	return nil
}

func (s *authorService) updateRole(ctx context.Context, userID string) error {
	url := fmt.Sprintf("%s/api/v1/auth/users/%s/role", s.authURL, userID)
	body := []byte(`{"role":"author"}`)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create role request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("auth service unavailable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("auth service returned status %d", resp.StatusCode)
	}

	return nil
}
