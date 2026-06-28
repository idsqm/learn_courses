package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/andruho/courses/internal/domain"
)

type FavoriteRepository interface {
	Add(ctx context.Context, userID string, courseID int) error
	Remove(ctx context.Context, userID string, courseID int) error
	ListByUser(ctx context.Context, userID string) ([]domain.CourseListItem, error)
}

type favoriteRepo struct {
	pool *pgxpool.Pool
}

func NewFavoriteRepository(pool *pgxpool.Pool) FavoriteRepository {
	return &favoriteRepo{pool: pool}
}

func (r *favoriteRepo) Add(ctx context.Context, userID string, courseID int) error {
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

func (r *favoriteRepo) Remove(ctx context.Context, userID string, courseID int) error {
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
			COALESCE(sub_rv.avg_rating, 0)::float,
			COALESCE(sub_rv.cnt, 0)::int,
			COALESCE(sub_e.cnt, 0)::int,
			(COALESCE(sub_l.total_dur, 0) / 60)::int,
			COALESCE(sub_l.cnt, 0)::int,
			c.color_1, c.color_2, c.tag
		FROM favorites f
		JOIN courses c ON f.course_id = c.id
		JOIN authors a ON c.author_id = a.id
		JOIN categories cat ON c.category_id = cat.id
		LEFT JOIN LATERAL (SELECT AVG(rating) AS avg_rating, COUNT(*) AS cnt FROM reviews WHERE course_id = c.id) sub_rv ON true
		LEFT JOIN LATERAL (SELECT COUNT(*) AS cnt FROM enrollments WHERE course_id = c.id) sub_e ON true
		LEFT JOIN LATERAL (SELECT SUM(duration_minutes) AS total_dur, COUNT(*) AS cnt FROM lessons WHERE course_id = c.id) sub_l ON true
		WHERE f.user_id = $1
		ORDER BY f.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	return scanCourseList(rows)
}
