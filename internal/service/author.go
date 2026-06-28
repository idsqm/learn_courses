package service

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
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
	if err := s.authors.Apply(ctx, userID, req); err != nil {
		return err
	}

	url := fmt.Sprintf("%s/api/v1/auth/users/%s/role", s.authURL, userID)
	body := []byte(`{"role":"author"}`)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		slog.Error("failed to create auth service request", "error", err, "userID", userID)
		return nil
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		slog.Error("failed to call auth service to set author role", "error", err, "userID", userID)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("auth service returned non-OK status for role update", "status", resp.StatusCode, "userID", userID)
	}

	return nil
}
