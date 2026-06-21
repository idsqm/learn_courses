package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/andruho/courses/internal/domain"
)

type ReviewRepository interface {
	Create(ctx context.Context, userID, courseID, name, initials, text string, rating int) error
}

type reviewRepo struct {
	pool *pgxpool.Pool
}

func NewReviewRepository(pool *pgxpool.Pool) ReviewRepository {
	return &reviewRepo{pool: pool}
}

func (r *reviewRepo) Create(ctx context.Context, userID, courseID, name, initials, text string, rating int) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO reviews (user_id, course_id, name, initials, text, rating)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, userID, courseID, name, initials, text, rating)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrReviewAlreadyExists
		}
		return err
	}
	return nil
}
