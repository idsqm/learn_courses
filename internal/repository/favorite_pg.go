package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/andruho/courses/internal/domain"
)

type FavoriteRepository interface {
	Add(ctx context.Context, userID, courseID string) error
	Remove(ctx context.Context, userID, courseID string) error
	ListByUser(ctx context.Context, userID string) ([]domain.CourseListItem, error)
}

type favoriteRepo struct {
	pool *pgxpool.Pool
}

func NewFavoriteRepository(pool *pgxpool.Pool) FavoriteRepository {
	return &favoriteRepo{pool: pool}
}

func (r *favoriteRepo) Add(ctx context.Context, userID, courseID string) error {
	_, err := r.pool.Exec(ctx,
		"INSERT INTO favorites (user_id, course_id) VALUES ($1, $2)", userID, courseID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrAlreadyFavorited
		}
		return err
	}
	return nil
}

func (r *favoriteRepo) Remove(ctx context.Context, userID, courseID string) error {
	tag, err := r.pool.Exec(ctx,
		"DELETE FROM favorites WHERE user_id = $1 AND course_id = $2", userID, courseID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFavorited
	}
	return nil
}

func (r *favoriteRepo) ListByUser(ctx context.Context, userID string) ([]domain.CourseListItem, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			c.id, c.title, a.name, a.initials,
			cat.name, c.level, c.price, c.old_price,
			COALESCE(AVG(rv.rating), 0)::float,
			COUNT(DISTINCT rv.id)::int,
			COUNT(DISTINCT e.id)::int,
			(COALESCE(SUM(DISTINCT l.duration_minutes), 0) / 60)::int,
			COUNT(DISTINCT l.id)::int,
			c.color_1, c.color_2, c.tag
		FROM favorites f
		JOIN courses c ON f.course_id = c.id
		JOIN authors a ON c.author_id = a.id
		JOIN categories cat ON c.category_id = cat.id
		LEFT JOIN reviews rv ON rv.course_id = c.id
		LEFT JOIN enrollments e ON e.course_id = c.id
		LEFT JOIN lessons l ON l.course_id = c.id
		WHERE f.user_id = $1
		GROUP BY c.id, a.name, a.initials, cat.name, f.created_at
		ORDER BY f.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	return scanCourseList(rows)
}
