package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/andruho/courses/internal/domain"
)

type CategoryRepository interface {
	List(ctx context.Context) ([]domain.Category, error)
}

type categoryRepo struct {
	pool *pgxpool.Pool
}

func NewCategoryRepository(pool *pgxpool.Pool) CategoryRepository {
	return &categoryRepo{pool: pool}
}

func (r *categoryRepo) List(ctx context.Context) ([]domain.Category, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			cat.id, cat.name, cat.abbreviation, cat.color,
			COUNT(c.id)::int AS courses_count
		FROM categories cat
		LEFT JOIN courses c ON c.category_id = cat.id AND c.published = true
		GROUP BY cat.id
		ORDER BY cat.name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []domain.Category
	for rows.Next() {
		var c domain.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Abbreviation, &c.Color, &c.CoursesCount); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	if categories == nil {
		categories = []domain.Category{}
	}
	return categories, nil
}
