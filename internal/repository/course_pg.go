package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/andruho/courses/internal/domain"
)

type CourseRepository interface {
	List(ctx context.Context, f domain.CourseFilter) ([]domain.CourseListItem, int, error)
	GetByID(ctx context.Context, id string) (*domain.CourseDetail, error)
	GetFeatured(ctx context.Context, limit int) ([]domain.CourseListItem, error)
	GetRecommended(ctx context.Context, userID string, limit int) ([]domain.CourseListItem, error)
	Exists(ctx context.Context, id string) (bool, error)
}

type courseRepo struct {
	pool *pgxpool.Pool
}

func NewCourseRepository(pool *pgxpool.Pool) CourseRepository {
	return &courseRepo{pool: pool}
}

func (r *courseRepo) List(ctx context.Context, f domain.CourseFilter) ([]domain.CourseListItem, int, error) {
	where := []string{"c.published = true"}
	args := []any{}
	argIdx := 1

	if len(f.Categories) > 0 {
		where = append(where, fmt.Sprintf("cat.name = ANY($%d)", argIdx))
		args = append(args, f.Categories)
		argIdx++
	}
	if f.Level != "" && f.Level != "Любой" {
		where = append(where, fmt.Sprintf("c.level = $%d", argIdx))
		args = append(args, f.Level)
		argIdx++
	}
	if f.PriceMin != nil {
		where = append(where, fmt.Sprintf("c.price >= $%d", argIdx))
		args = append(args, *f.PriceMin)
		argIdx++
	}
	if f.PriceMax != nil {
		where = append(where, fmt.Sprintf("c.price <= $%d", argIdx))
		args = append(args, *f.PriceMax)
		argIdx++
	}
	if f.Search != "" {
		where = append(where, fmt.Sprintf("(c.title ILIKE $%d OR a.name ILIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+f.Search+"%")
		argIdx++
	}

	whereClause := strings.Join(where, " AND ")

	havingClause := ""
	if f.RatingMin != nil {
		havingClause = fmt.Sprintf("HAVING COALESCE(AVG(rv.rating), 0) >= $%d", argIdx)
		args = append(args, *f.RatingMin)
		argIdx++
	}

	orderBy := "c.created_at DESC"
	switch f.Sort {
	case "popular", "популярные":
		orderBy = "COUNT(DISTINCT e.id) DESC"
	case "new", "новые":
		orderBy = "c.created_at DESC"
	case "price_asc", "по цене":
		orderBy = "c.price ASC"
	case "price_desc":
		orderBy = "c.price DESC"
	case "rating", "по рейтингу":
		orderBy = "COALESCE(AVG(rv.rating), 0) DESC"
	}

	baseQuery := fmt.Sprintf(`
		SELECT
			c.id, c.title, a.name AS author, a.initials,
			cat.name AS category, c.level, c.price, c.old_price,
			COALESCE(AVG(rv.rating), 0)::float AS rating,
			COUNT(DISTINCT rv.id)::int AS reviews_count,
			COUNT(DISTINCT e.id)::int AS students_count,
			(COALESCE(SUM(DISTINCT l.duration_minutes), 0) / 60)::int AS hours,
			COUNT(DISTINCT l.id)::int AS lessons_count,
			c.color_1, c.color_2, c.tag
		FROM courses c
		JOIN authors a ON c.author_id = a.id
		JOIN categories cat ON c.category_id = cat.id
		LEFT JOIN reviews rv ON rv.course_id = c.id
		LEFT JOIN enrollments e ON e.course_id = c.id
		LEFT JOIN lessons l ON l.course_id = c.id
		WHERE %s
		GROUP BY c.id, a.name, a.initials, cat.name
		%s
	`, whereClause, havingClause)

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM (%s) sub", baseQuery)
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count courses: %w", err)
	}

	offset := (f.Page - 1) * f.PerPage
	dataQuery := fmt.Sprintf("%s ORDER BY %s LIMIT $%d OFFSET $%d", baseQuery, orderBy, argIdx, argIdx+1)
	args = append(args, f.PerPage, offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query courses: %w", err)
	}

	courses, err := scanCourseList(rows)
	if err != nil {
		return nil, 0, err
	}

	return courses, total, nil
}

func (r *courseRepo) GetByID(ctx context.Context, id string) (*domain.CourseDetail, error) {
	var d domain.CourseDetail
	var authorID string
	err := r.pool.QueryRow(ctx, `
		SELECT
			c.id, c.title, c.description, cat.name AS category,
			c.level, c.price, c.old_price,
			COALESCE(sub_rv.avg_rating, 0)::float,
			COALESCE(sub_rv.cnt, 0)::int,
			COALESCE(sub_e.cnt, 0)::int,
			(COALESCE(sub_l.total_dur, 0) / 60)::int,
			COALESCE(sub_l.cnt, 0)::int,
			c.color_1, c.color_2, c.tag, c.preview_url,
			c.author_id
		FROM courses c
		JOIN categories cat ON c.category_id = cat.id
		LEFT JOIN LATERAL (SELECT AVG(rating) AS avg_rating, COUNT(*) AS cnt FROM reviews WHERE course_id = c.id) sub_rv ON true
		LEFT JOIN LATERAL (SELECT COUNT(*) AS cnt FROM enrollments WHERE course_id = c.id) sub_e ON true
		LEFT JOIN LATERAL (SELECT SUM(duration_minutes) AS total_dur, COUNT(*) AS cnt FROM lessons WHERE course_id = c.id) sub_l ON true
		WHERE c.id = $1
	`, id).Scan(
		&d.ID, &d.Title, &d.Description, &d.Category,
		&d.Level, &d.Price, &d.OldPrice,
		&d.Rating, &d.ReviewsCount, &d.StudentsCount,
		&d.Hours, &d.LessonsCount,
		&d.Color1, &d.Color2, &d.Tag, &d.PreviewURL,
		&authorID,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get course: %w", err)
	}

	var errAgg error

	d.LearnItems, errAgg = r.getStringList(ctx, "SELECT text FROM course_learn_items WHERE course_id = $1 ORDER BY sort_order", id)
	if errAgg != nil {
		return nil, errAgg
	}

	d.Includes, errAgg = r.getStringList(ctx, "SELECT text FROM course_includes WHERE course_id = $1 ORDER BY sort_order", id)
	if errAgg != nil {
		return nil, errAgg
	}

	d.Curriculum, errAgg = r.getCurriculum(ctx, id)
	if errAgg != nil {
		return nil, errAgg
	}

	d.Reviews, errAgg = r.getReviews(ctx, id)
	if errAgg != nil {
		return nil, errAgg
	}

	author, errAgg := r.getAuthor(ctx, authorID)
	if errAgg != nil {
		return nil, errAgg
	}
	d.Author = *author

	return &d, nil
}

func (r *courseRepo) GetFeatured(ctx context.Context, limit int) ([]domain.CourseListItem, error) {
	return r.listByOrder(ctx, "COUNT(DISTINCT e.id) DESC, COALESCE(AVG(rv.rating), 0) DESC", limit)
}

func (r *courseRepo) GetRecommended(ctx context.Context, userID string, limit int) ([]domain.CourseListItem, error) {
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
		FROM courses c
		JOIN authors a ON c.author_id = a.id
		JOIN categories cat ON c.category_id = cat.id
		LEFT JOIN reviews rv ON rv.course_id = c.id
		LEFT JOIN enrollments e ON e.course_id = c.id
		LEFT JOIN lessons l ON l.course_id = c.id
		WHERE c.published = true
			AND c.id NOT IN (SELECT course_id FROM enrollments WHERE user_id = $1)
		GROUP BY c.id, a.name, a.initials, cat.name
		ORDER BY COALESCE(AVG(rv.rating), 0) DESC, COUNT(DISTINCT e.id) DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	return scanCourseList(rows)
}

func (r *courseRepo) Exists(ctx context.Context, id string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM courses WHERE id = $1)", id).Scan(&exists)
	return exists, err
}

func (r *courseRepo) listByOrder(ctx context.Context, order string, limit int) ([]domain.CourseListItem, error) {
	query := fmt.Sprintf(`
		SELECT
			c.id, c.title, a.name, a.initials,
			cat.name, c.level, c.price, c.old_price,
			COALESCE(AVG(rv.rating), 0)::float,
			COUNT(DISTINCT rv.id)::int,
			COUNT(DISTINCT e.id)::int,
			(COALESCE(SUM(DISTINCT l.duration_minutes), 0) / 60)::int,
			COUNT(DISTINCT l.id)::int,
			c.color_1, c.color_2, c.tag
		FROM courses c
		JOIN authors a ON c.author_id = a.id
		JOIN categories cat ON c.category_id = cat.id
		LEFT JOIN reviews rv ON rv.course_id = c.id
		LEFT JOIN enrollments e ON e.course_id = c.id
		LEFT JOIN lessons l ON l.course_id = c.id
		WHERE c.published = true
		GROUP BY c.id, a.name, a.initials, cat.name
		ORDER BY %s
		LIMIT $1
	`, order)

	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	return scanCourseList(rows)
}

func (r *courseRepo) getStringList(ctx context.Context, query, courseID string) ([]string, error) {
	rows, err := r.pool.Query(ctx, query, courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		items = append(items, s)
	}
	if items == nil {
		items = []string{}
	}
	return items, nil
}

func (r *courseRepo) getCurriculum(ctx context.Context, courseID string) ([]domain.Module, error) {
	moduleRows, err := r.pool.Query(ctx,
		"SELECT id, title FROM course_modules WHERE course_id = $1 ORDER BY sort_order", courseID)
	if err != nil {
		return nil, err
	}
	defer moduleRows.Close()

	type moduleRow struct {
		id, title string
	}
	var mods []moduleRow
	for moduleRows.Next() {
		var m moduleRow
		if err := moduleRows.Scan(&m.id, &m.title); err != nil {
			return nil, err
		}
		mods = append(mods, m)
	}

	var curriculum []domain.Module
	for _, m := range mods {
		lessonRows, err := r.pool.Query(ctx,
			"SELECT id, name, duration_minutes, is_free FROM lessons WHERE module_id = $1 ORDER BY sort_order", m.id)
		if err != nil {
			return nil, err
		}

		var lessons []domain.Lesson
		var totalMinutes int
		for lessonRows.Next() {
			var l domain.Lesson
			var dur int
			if err := lessonRows.Scan(&l.ID, &l.Name, &dur, &l.IsFree); err != nil {
				lessonRows.Close()
				return nil, err
			}
			l.Duration = formatDuration(dur)
			totalMinutes += dur
			lessons = append(lessons, l)
		}
		lessonRows.Close()

		if lessons == nil {
			lessons = []domain.Lesson{}
		}

		curriculum = append(curriculum, domain.Module{
			Title:        m.title,
			Duration:     formatDuration(totalMinutes),
			LessonsCount: len(lessons),
			Lessons:      lessons,
		})
	}

	if curriculum == nil {
		curriculum = []domain.Module{}
	}
	return curriculum, nil
}

func (r *courseRepo) getReviews(ctx context.Context, courseID string) ([]domain.Review, error) {
	rows, err := r.pool.Query(ctx,
		"SELECT id, name, initials, text, rating FROM reviews WHERE course_id = $1 ORDER BY created_at DESC", courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []domain.Review
	for rows.Next() {
		var rv domain.Review
		if err := rows.Scan(&rv.ID, &rv.Name, &rv.Initials, &rv.Text, &rv.Rating); err != nil {
			return nil, err
		}
		reviews = append(reviews, rv)
	}
	if reviews == nil {
		reviews = []domain.Review{}
	}
	return reviews, nil
}

func (r *courseRepo) getAuthor(ctx context.Context, authorID string) (*domain.AuthorInfo, error) {
	var a domain.AuthorInfo
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, initials, subtitle, bio,
			(SELECT COUNT(*) FROM courses WHERE author_id = authors.id AND published = true)::int,
			years_experience
		FROM authors WHERE id = $1
	`, authorID).Scan(&a.ID, &a.Name, &a.Initials, &a.Subtitle, &a.Bio, &a.CoursesCount, &a.YearsExperience)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

func scanCourseList(rows pgx.Rows) ([]domain.CourseListItem, error) {
	defer rows.Close()
	var courses []domain.CourseListItem
	for rows.Next() {
		var c domain.CourseListItem
		if err := rows.Scan(
			&c.ID, &c.Title, &c.Author, &c.Initials,
			&c.Category, &c.Level, &c.Price, &c.OldPrice,
			&c.Rating, &c.ReviewsCount, &c.StudentsCount,
			&c.Hours, &c.LessonsCount,
			&c.Color1, &c.Color2, &c.Tag,
		); err != nil {
			return nil, err
		}
		courses = append(courses, c)
	}
	if courses == nil {
		courses = []domain.CourseListItem{}
	}
	return courses, nil
}

func formatDuration(minutes int) string {
	if minutes < 60 {
		return fmt.Sprintf("%d мин", minutes)
	}
	h := minutes / 60
	m := minutes % 60
	if m == 0 {
		return fmt.Sprintf("%d ч", h)
	}
	return fmt.Sprintf("%d ч %d мин", h, m)
}
