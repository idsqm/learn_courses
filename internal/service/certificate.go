package service

import (
	"context"

	"github.com/andruho/courses/internal/domain"
	"github.com/andruho/courses/internal/repository"
)

type CertificateService interface {
	ListByUser(ctx context.Context, userID string) ([]domain.Certificate, error)
	GetByID(ctx context.Context, userID string, certID int) (*domain.Certificate, error)
}

type certificateService struct {
	certificates repository.CertificateRepository
}

func NewCertificateService(certificates repository.CertificateRepository) CertificateService {
	return &certificateService{certificates: certificates}
}

func (s *certificateService) ListByUser(ctx context.Context, userID string) ([]domain.Certificate, error) {
	return s.certificates.ListByUser(ctx, userID)
}

func (s *certificateService) GetByID(ctx context.Context, userID string, certID int) (*domain.Certificate, error) {
	return s.certificates.GetByID(ctx, userID, certID)
}
