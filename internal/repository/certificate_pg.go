package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/andruho/courses/internal/domain"
)

type CertificateRepository interface {
	ListByUser(ctx context.Context, userID string) ([]domain.Certificate, error)
	GetByID(ctx context.Context, userID, certID string) (*domain.Certificate, error)
	Issue(ctx context.Context, userID, courseID string) error
}

type certificateRepo struct {
	pool *pgxpool.Pool
}

func NewCertificateRepository(pool *pgxpool.Pool) CertificateRepository {
	return &certificateRepo{pool: pool}
}

func (r *certificateRepo) ListByUser(ctx context.Context, userID string) ([]domain.Certificate, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT cert.id, cert.course_id, c.title, cert.issued_at
		FROM certificates cert
		JOIN courses c ON cert.course_id = c.id
		WHERE cert.user_id = $1
		ORDER BY cert.issued_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var certs []domain.Certificate
	for rows.Next() {
		var c domain.Certificate
		if err := rows.Scan(&c.ID, &c.CourseID, &c.CourseName, &c.IssuedAt); err != nil {
			return nil, err
		}
		certs = append(certs, c)
	}
	if certs == nil {
		certs = []domain.Certificate{}
	}
	return certs, nil
}

func (r *certificateRepo) GetByID(ctx context.Context, userID, certID string) (*domain.Certificate, error) {
	var c domain.Certificate
	err := r.pool.QueryRow(ctx, `
		SELECT cert.id, cert.course_id, co.title, cert.issued_at
		FROM certificates cert
		JOIN courses co ON cert.course_id = co.id
		WHERE cert.id = $1 AND cert.user_id = $2
	`, certID, userID).Scan(&c.ID, &c.CourseID, &c.CourseName, &c.IssuedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (r *certificateRepo) Issue(ctx context.Context, userID, courseID string) error {
	_, err := r.pool.Exec(ctx,
		"INSERT INTO certificates (user_id, course_id) VALUES ($1, $2) ON CONFLICT DO NOTHING",
		userID, courseID)
	return err
}
