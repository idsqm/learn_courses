package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/andruho/courses/internal/domain"
)

type ProgressRepository interface {
	CompleteLesson(ctx context.Context, userID, courseID, lessonID string) error
	LessonExists(ctx context.Context, courseID, lessonID string) (bool, error)
	GetCourseProgress(ctx context.Context, userID, courseID string) (*domain.CourseProgress, error)
	IsFullyCompleted(ctx context.Context, userID, courseID string) (bool, error)
	GetUserStats(ctx context.Context, userID string) (*domain.UserStats, error)
}

type progressRepo struct {
	pool *pgxpool.Pool
}

func NewProgressRepository(pool *pgxpool.Pool) ProgressRepository {
	return &progressRepo{pool: pool}
}

func (r *progressRepo) CompleteLesson(ctx context.Context, userID, courseID, lessonID string) error {
	_, err := r.pool.Exec(ctx,
		"INSERT INTO lesson_progress (user_id, course_id, lesson_id) VALUES ($1, $2, $3)",
		userID, courseID, lessonID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrAlreadyCompleted
		}
		return err
	}
	return nil
}

func (r *progressRepo) LessonExists(ctx context.Context, courseID, lessonID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM lessons WHERE id = $1 AND course_id = $2)",
		lessonID, courseID).Scan(&exists)
	return exists, err
}

func (r *progressRepo) GetCourseProgress(ctx context.Context, userID, courseID string) (*domain.CourseProgress, error) {
	var p domain.CourseProgress
	p.CourseID = courseID

	err := r.pool.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM lessons WHERE course_id = $2)::int,
			(SELECT COUNT(*) FROM lesson_progress WHERE user_id = $1 AND course_id = $2)::int
	`, userID, courseID).Scan(&p.TotalLessons, &p.DoneLessons)
	if err != nil {
		return nil, err
	}

	if p.TotalLessons > 0 {
		p.Progress = float64(p.DoneLessons) / float64(p.TotalLessons) * 100
	}
	p.Completed = p.TotalLessons > 0 && p.DoneLessons >= p.TotalLessons

	return &p, nil
}

func (r *progressRepo) IsFullyCompleted(ctx context.Context, userID, courseID string) (bool, error) {
	p, err := r.GetCourseProgress(ctx, userID, courseID)
	if err != nil {
		return false, err
	}
	return p.Completed, nil
}

func (r *progressRepo) GetUserStats(ctx context.Context, userID string) (*domain.UserStats, error) {
	var stats domain.UserStats

	err := r.pool.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(DISTINCT en.course_id) FROM enrollments en
			 WHERE en.user_id = $1
			 AND (SELECT COUNT(*) FROM lessons WHERE course_id = en.course_id) >
			     (SELECT COUNT(*) FROM lesson_progress WHERE user_id = $1 AND course_id = en.course_id)
			)::int,
			(SELECT COALESCE(SUM(l.duration_minutes), 0) FROM enrollments en2
			 JOIN lessons l ON l.course_id = en2.course_id
			 WHERE en2.user_id = $1
			)::int,
			(SELECT COUNT(*) FROM certificates WHERE user_id = $1)::int
	`, userID).Scan(&stats.CoursesInProgress, &stats.TotalStudyTime, &stats.CertificatesCount)
	if err != nil {
		return nil, err
	}

	streak, err := r.calculateStreak(ctx, userID)
	if err != nil {
		return nil, err
	}
	stats.StudyStreakDays = streak

	return &stats, nil
}

func (r *progressRepo) calculateStreak(ctx context.Context, userID string) (int, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT DISTINCT completed_at::date AS d
		FROM lesson_progress
		WHERE user_id = $1
		ORDER BY d DESC
	`, userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	defer rows.Close()

	streak := 0
	expected := time.Now().Truncate(24 * time.Hour)

	for rows.Next() {
		var d time.Time
		if err := rows.Scan(&d); err != nil {
			return 0, err
		}
		day := d.Truncate(24 * time.Hour)

		if day.Equal(expected) {
			streak++
			expected = expected.AddDate(0, 0, -1)
		} else if day.Equal(expected.AddDate(0, 0, -1)) && streak == 0 {
			expected = day
			streak++
			expected = expected.AddDate(0, 0, -1)
		} else {
			break
		}
	}

	return streak, nil
}
