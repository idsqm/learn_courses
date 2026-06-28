package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/andruho/courses/internal/domain"
)

type AuthorRepository interface {
	GetByID(ctx context.Context, id int) (*domain.AuthorInfo, error)
	Apply(ctx context.Context, userID string) error
}

type authorRepo struct {
	pool *pgxpool.Pool
}

func NewAuthorRepository(pool *pgxpool.Pool) AuthorRepository {
	return &authorRepo{pool: pool}
}

func (r *authorRepo) GetByID(ctx context.Context, id int) (*domain.AuthorInfo, error) {
	var a domain.AuthorInfo
	err := r.pool.QueryRow(ctx, `
		SELECT a.id, a.name, a.initials, a.subtitle, a.bio,
			(SELECT COUNT(*) FROM courses WHERE author_id = a.id AND published = true)::int,
			a.years_experience
		FROM authors a
		WHERE a.id = $1
	`, id).Scan(&a.ID, &a.Name, &a.Initials, &a.Subtitle, &a.Bio, &a.CoursesCount, &a.YearsExperience)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

func (r *authorRepo) Apply(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx,
		"INSERT INTO author_applications (user_id) VALUES ($1)", userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrAlreadyApplied
		}
		return err
	}
	return nil
}
