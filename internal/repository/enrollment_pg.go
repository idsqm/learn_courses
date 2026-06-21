package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/andruho/courses/internal/domain"
)

type EnrollmentRepository interface {
	Enroll(ctx context.Context, userID, courseID string) error
	ListByUser(ctx context.Context, userID string) ([]domain.EnrolledCourse, error)
	IsEnrolled(ctx context.Context, userID, courseID string) (bool, error)
}

type enrollmentRepo struct {
	pool *pgxpool.Pool
}

func NewEnrollmentRepository(pool *pgxpool.Pool) EnrollmentRepository {
	return &enrollmentRepo{pool: pool}
}

func (r *enrollmentRepo) Enroll(ctx context.Context, userID, courseID string) error {
	_, err := r.pool.Exec(ctx,
		"INSERT INTO enrollments (user_id, course_id) VALUES ($1, $2)", userID, courseID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrAlreadyEnrolled
		}
		return err
	}
	return nil
}

func (r *enrollmentRepo) ListByUser(ctx context.Context, userID string) ([]domain.EnrolledCourse, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			c.id, c.title, a.name, a.initials,
			cat.name, c.level, c.price, c.old_price,
			COALESCE(sub_rv.avg_rating, 0)::float,
			COALESCE(sub_rv.cnt, 0)::int,
			COALESCE(sub_e.cnt, 0)::int,
			(COALESCE(sub_l.total_dur, 0) / 60)::int,
			COALESCE(sub_l.cnt, 0)::int,
			c.color_1, c.color_2, c.tag,
			COALESCE(sub_l.cnt, 0)::int AS total_lessons,
			COALESCE(sub_p.done, 0)::int AS done_lessons,
			sub_p.last_lesson
		FROM enrollments en
		JOIN courses c ON en.course_id = c.id
		JOIN authors a ON c.author_id = a.id
		JOIN categories cat ON c.category_id = cat.id
		LEFT JOIN LATERAL (SELECT AVG(rating) AS avg_rating, COUNT(*) AS cnt FROM reviews WHERE course_id = c.id) sub_rv ON true
		LEFT JOIN LATERAL (SELECT COUNT(*) AS cnt FROM enrollments WHERE course_id = c.id) sub_e ON true
		LEFT JOIN LATERAL (SELECT SUM(duration_minutes) AS total_dur, COUNT(*) AS cnt FROM lessons WHERE course_id = c.id) sub_l ON true
		LEFT JOIN LATERAL (
			SELECT COUNT(*) AS done,
				(SELECT l.name FROM lesson_progress lp2 JOIN lessons l ON l.id = lp2.lesson_id
				 WHERE lp2.user_id = en.user_id AND lp2.course_id = c.id
				 ORDER BY lp2.completed_at DESC LIMIT 1) AS last_lesson
			FROM lesson_progress WHERE user_id = en.user_id AND course_id = c.id
		) sub_p ON true
		WHERE en.user_id = $1
		ORDER BY en.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.EnrolledCourse
	for rows.Next() {
		var ec domain.EnrolledCourse
		var c domain.CourseListItem
		if err := rows.Scan(
			&c.ID, &c.Title, &c.Author, &c.Initials,
			&c.Category, &c.Level, &c.Price, &c.OldPrice,
			&c.Rating, &c.ReviewsCount, &c.StudentsCount,
			&c.Hours, &c.LessonsCount,
			&c.Color1, &c.Color2, &c.Tag,
			&ec.TotalLessons, &ec.DoneLessons, &ec.LastLesson,
		); err != nil {
			return nil, fmt.Errorf("scan enrolled course: %w", err)
		}
		ec.Course = c
		if ec.TotalLessons > 0 {
			ec.Progress = float64(ec.DoneLessons) / float64(ec.TotalLessons) * 100
		}
		ec.Completed = ec.TotalLessons > 0 && ec.DoneLessons >= ec.TotalLessons
		result = append(result, ec)
	}
	if result == nil {
		result = []domain.EnrolledCourse{}
	}
	return result, nil
}

func (r *enrollmentRepo) IsEnrolled(ctx context.Context, userID, courseID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM enrollments WHERE user_id = $1 AND course_id = $2)",
		userID, courseID).Scan(&exists)
	if err != nil && err != pgx.ErrNoRows {
		return false, err
	}
	return exists, nil
}
