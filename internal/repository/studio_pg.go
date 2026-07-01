package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/andruho/courses/internal/domain"
)

type StudioRepository interface {
	GetAuthorByUserID(ctx context.Context, userID string) (int, error)

	ListCourses(ctx context.Context, authorID int) ([]domain.StudioCourse, error)
	GetCourse(ctx context.Context, authorID, courseID int) (*domain.StudioCourseDetail, error)
	CreateCourse(ctx context.Context, authorID int, req domain.CreateCourseRequest) (int, error)
	UpdateCourse(ctx context.Context, authorID, courseID int, req domain.UpdateCourseRequest) error
	DeleteCourse(ctx context.Context, authorID, courseID int) error
	PublishCourse(ctx context.Context, authorID, courseID int) error
	UnpublishCourse(ctx context.Context, authorID, courseID int) error

	CreateModule(ctx context.Context, authorID, courseID int, req domain.CreateModuleRequest) (int, error)
	UpdateModule(ctx context.Context, authorID, courseID, moduleID int, req domain.UpdateModuleRequest) error
	DeleteModule(ctx context.Context, authorID, courseID, moduleID int) error

	CreateLesson(ctx context.Context, authorID, courseID, moduleID int, req domain.CreateLessonRequest) (int, error)
	UpdateLesson(ctx context.Context, authorID, courseID, lessonID int, req domain.UpdateLessonRequest) error
	DeleteLesson(ctx context.Context, authorID, courseID, lessonID int) error

	GetStats(ctx context.Context, authorID int) (*domain.StudioStats, error)
	ListStudents(ctx context.Context, authorID int) ([]domain.StudioStudent, error)
	GetIncome(ctx context.Context, authorID int) (*domain.StudioIncome, error)
	ListPayouts(ctx context.Context, authorID int) ([]domain.Payout, error)
	ListReviews(ctx context.Context, authorID int) ([]domain.StudioReview, error)
	ReplyToReview(ctx context.Context, authorID, reviewID int, reply string) error
}

type studioRepo struct {
	pool *pgxpool.Pool
}

func NewStudioRepository(pool *pgxpool.Pool) StudioRepository {
	return &studioRepo{pool: pool}
}

func (r *studioRepo) GetAuthorByUserID(ctx context.Context, userID string) (int, error) {
	var authorID int
	err := r.pool.QueryRow(ctx,
		"SELECT id FROM authors WHERE user_id = $1 AND approved = true", userID).Scan(&authorID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, domain.ErrNotAuthor
		}
		return 0, err
	}
	return authorID, nil
}

// --- Courses ---

func (r *studioRepo) ListCourses(ctx context.Context, authorID int) ([]domain.StudioCourse, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			c.id, c.title, cat.name,
			COALESCE(sub_l.cnt, 0)::int,
			COALESCE(sub_e.cnt, 0)::int,
			(COALESCE(sub_e30.cnt, 0) * c.price)::float,
			CASE WHEN c.published THEN 'published' ELSE 'draft' END,
			c.preview_url
		FROM courses c
		JOIN categories cat ON c.category_id = cat.id
		LEFT JOIN LATERAL (SELECT COUNT(*) AS cnt FROM lessons WHERE course_id = c.id) sub_l ON true
		LEFT JOIN LATERAL (SELECT COUNT(*) AS cnt FROM enrollments WHERE course_id = c.id) sub_e ON true
		LEFT JOIN LATERAL (
			SELECT COUNT(*) AS cnt
			FROM enrollments WHERE course_id = c.id AND created_at >= NOW() - INTERVAL '30 days'
		) sub_e30 ON true
		WHERE c.author_id = $1
		ORDER BY c.created_at DESC
	`, authorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var courses []domain.StudioCourse
	for rows.Next() {
		var sc domain.StudioCourse
		if err := rows.Scan(&sc.ID, &sc.Title, &sc.Category, &sc.LessonsCount, &sc.StudentsCount, &sc.Revenue30d, &sc.Status, &sc.PreviewURL); err != nil {
			return nil, err
		}
		courses = append(courses, sc)
	}
	if courses == nil {
		courses = []domain.StudioCourse{}
	}
	return courses, nil
}

func (r *studioRepo) GetCourse(ctx context.Context, authorID, courseID int) (*domain.StudioCourseDetail, error) {
	var d domain.StudioCourseDetail
	err := r.pool.QueryRow(ctx, `
		SELECT c.id, c.title, c.subtitle, c.description,
			c.category_id, cat.name, c.level, c.price, c.old_price,
			c.is_free, c.color_1, c.color_2, c.tag, c.preview_url, c.published
		FROM courses c
		JOIN categories cat ON c.category_id = cat.id
		WHERE c.id = $1 AND c.author_id = $2
	`, courseID, authorID).Scan(
		&d.ID, &d.Title, &d.Subtitle, &d.Description,
		&d.CategoryID, &d.Category, &d.Level, &d.Price, &d.OldPrice,
		&d.IsFree, &d.Color1, &d.Color2, &d.Tag, &d.PreviewURL, &d.Published,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrCourseNotFound
		}
		return nil, err
	}

	moduleRows, err := r.pool.Query(ctx,
		"SELECT id, title FROM course_modules WHERE course_id = $1 ORDER BY sort_order", courseID)
	if err != nil {
		return nil, err
	}
	defer moduleRows.Close()

	type modRow struct{ id int; title string }
	var mods []modRow
	for moduleRows.Next() {
		var m modRow
		if err := moduleRows.Scan(&m.id, &m.title); err != nil {
			return nil, err
		}
		mods = append(mods, m)
	}

	d.Curriculum = make([]domain.Module, 0, len(mods))
	for _, m := range mods {
		lessonRows, err := r.pool.Query(ctx,
			"SELECT id, name, type, duration_minutes, is_free FROM lessons WHERE module_id = $1 ORDER BY sort_order", m.id)
		if err != nil {
			return nil, err
		}
		var lessons []domain.Lesson
		var totalMin int
		for lessonRows.Next() {
			var l domain.Lesson
			var dur int
			if err := lessonRows.Scan(&l.ID, &l.Name, &l.Type, &dur, &l.IsFree); err != nil {
				lessonRows.Close()
				return nil, err
			}
			l.Duration = formatDuration(dur)
			totalMin += dur
			lessons = append(lessons, l)
		}
		lessonRows.Close()
		if lessons == nil {
			lessons = []domain.Lesson{}
		}
		d.Curriculum = append(d.Curriculum, domain.Module{
			ID:           m.id,
			Title:        m.title,
			Duration:     formatDuration(totalMin),
			LessonsCount: len(lessons),
			Lessons:      lessons,
		})
	}

	d.LearnItems, err = r.getStringList(ctx, "SELECT text FROM course_learn_items WHERE course_id = $1 ORDER BY sort_order", courseID)
	if err != nil {
		return nil, err
	}
	d.Includes, err = r.getStringList(ctx, "SELECT text FROM course_includes WHERE course_id = $1 ORDER BY sort_order", courseID)
	if err != nil {
		return nil, err
	}

	return &d, nil
}

func (r *studioRepo) getStringList(ctx context.Context, query string, courseID int) ([]string, error) {
	rows, err := r.pool.Query(ctx, query, courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []string{}
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		items = append(items, s)
	}
	return items, nil
}

func (r *studioRepo) CreateCourse(ctx context.Context, authorID int, req domain.CreateCourseRequest) (int, error) {
	color1 := req.Color1
	if color1 == "" {
		color1 = "#6366f1"
	}
	color2 := req.Color2
	if color2 == "" {
		color2 = "#8b5cf6"
	}

	var id int
	err := r.pool.QueryRow(ctx, `
		INSERT INTO courses (title, subtitle, description, author_id, category_id, level, price, old_price, is_free, color_1, color_2)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`, req.Title, req.Subtitle, req.Description, authorID, req.CategoryID,
		req.Level, req.Price, req.OldPrice, req.IsFree, color1, color2).Scan(&id)
	if err != nil {
		return 0, err
	}

	if len(req.LearnItems) > 0 {
		if err := r.replaceStringList(ctx, "course_learn_items", id, req.LearnItems); err != nil {
			return 0, err
		}
	}
	if len(req.Includes) > 0 {
		if err := r.replaceStringList(ctx, "course_includes", id, req.Includes); err != nil {
			return 0, err
		}
	}

	return id, nil
}

func (r *studioRepo) UpdateCourse(ctx context.Context, authorID, courseID int, req domain.UpdateCourseRequest) error {
	if err := r.verifyCourseOwner(ctx, authorID, courseID); err != nil {
		return err
	}

	sets := []string{}
	args := []any{}
	idx := 1

	add := func(col string, val any) {
		sets = append(sets, fmt.Sprintf("%s = $%d", col, idx))
		args = append(args, val)
		idx++
	}

	if req.Title != nil {
		add("title", *req.Title)
	}
	if req.Subtitle != nil {
		add("subtitle", *req.Subtitle)
	}
	if req.CategoryID != nil {
		add("category_id", *req.CategoryID)
	}
	if req.Level != nil {
		add("level", *req.Level)
	}
	if req.Description != nil {
		add("description", *req.Description)
	}
	if req.Price != nil {
		add("price", *req.Price)
	}
	if req.OldPrice != nil {
		add("old_price", *req.OldPrice)
	}
	if req.IsFree != nil {
		add("is_free", *req.IsFree)
	}
	if req.Color1 != nil {
		add("color_1", *req.Color1)
	}
	if req.Color2 != nil {
		add("color_2", *req.Color2)
	}
	if req.Tag != nil {
		add("tag", *req.Tag)
	}
	if req.PreviewURL != nil {
		add("preview_url", *req.PreviewURL)
	}

	if len(sets) > 0 {
		sets = append(sets, "updated_at = NOW()")
		query := fmt.Sprintf("UPDATE courses SET %s WHERE id = $%d AND author_id = $%d",
			strings.Join(sets, ", "), idx, idx+1)
		args = append(args, courseID, authorID)

		if _, err := r.pool.Exec(ctx, query, args...); err != nil {
			return err
		}
	}

	if req.LearnItems != nil {
		if err := r.replaceStringList(ctx, "course_learn_items", courseID, *req.LearnItems); err != nil {
			return err
		}
	}
	if req.Includes != nil {
		if err := r.replaceStringList(ctx, "course_includes", courseID, *req.Includes); err != nil {
			return err
		}
	}

	return nil
}

// replaceStringList overwrites a course's child rows (learn items / includes) with the given list.
// table must be a fixed internal constant, never user input.
func (r *studioRepo) replaceStringList(ctx context.Context, table string, courseID int, items []string) error {
	if _, err := r.pool.Exec(ctx, fmt.Sprintf("DELETE FROM %s WHERE course_id = $1", table), courseID); err != nil {
		return err
	}
	for i, text := range items {
		if _, err := r.pool.Exec(ctx,
			fmt.Sprintf("INSERT INTO %s (course_id, text, sort_order) VALUES ($1, $2, $3)", table),
			courseID, text, i); err != nil {
			return err
		}
	}
	return nil
}

func (r *studioRepo) DeleteCourse(ctx context.Context, authorID, courseID int) error {
	tag, err := r.pool.Exec(ctx, "DELETE FROM courses WHERE id = $1 AND author_id = $2", courseID, authorID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrCourseNotFound
	}
	return nil
}

func (r *studioRepo) PublishCourse(ctx context.Context, authorID, courseID int) error {
	return r.setCoursePublished(ctx, authorID, courseID, true)
}

func (r *studioRepo) UnpublishCourse(ctx context.Context, authorID, courseID int) error {
	return r.setCoursePublished(ctx, authorID, courseID, false)
}

func (r *studioRepo) setCoursePublished(ctx context.Context, authorID, courseID int, published bool) error {
	tag, err := r.pool.Exec(ctx,
		"UPDATE courses SET published = $1, updated_at = NOW() WHERE id = $2 AND author_id = $3",
		published, courseID, authorID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrCourseNotFound
	}
	return nil
}

// --- Modules ---

func (r *studioRepo) CreateModule(ctx context.Context, authorID, courseID int, req domain.CreateModuleRequest) (int, error) {
	if err := r.verifyCourseOwner(ctx, authorID, courseID); err != nil {
		return 0, err
	}
	var id int
	err := r.pool.QueryRow(ctx,
		"INSERT INTO course_modules (course_id, title, sort_order) VALUES ($1, $2, $3) RETURNING id",
		courseID, req.Title, req.SortOrder).Scan(&id)
	return id, err
}

func (r *studioRepo) UpdateModule(ctx context.Context, authorID, courseID, moduleID int, req domain.UpdateModuleRequest) error {
	if err := r.verifyCourseOwner(ctx, authorID, courseID); err != nil {
		return err
	}

	sets := []string{}
	args := []any{}
	idx := 1

	if req.Title != nil {
		sets = append(sets, fmt.Sprintf("title = $%d", idx))
		args = append(args, *req.Title)
		idx++
	}
	if req.SortOrder != nil {
		sets = append(sets, fmt.Sprintf("sort_order = $%d", idx))
		args = append(args, *req.SortOrder)
		idx++
	}

	if len(sets) == 0 {
		return nil
	}

	query := fmt.Sprintf("UPDATE course_modules SET %s WHERE id = $%d AND course_id = $%d",
		strings.Join(sets, ", "), idx, idx+1)
	args = append(args, moduleID, courseID)

	tag, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrModuleNotFound
	}
	return nil
}

func (r *studioRepo) DeleteModule(ctx context.Context, authorID, courseID, moduleID int) error {
	if err := r.verifyCourseOwner(ctx, authorID, courseID); err != nil {
		return err
	}
	tag, err := r.pool.Exec(ctx, "DELETE FROM course_modules WHERE id = $1 AND course_id = $2", moduleID, courseID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrModuleNotFound
	}
	return nil
}

// --- Lessons ---

func (r *studioRepo) CreateLesson(ctx context.Context, authorID, courseID, moduleID int, req domain.CreateLessonRequest) (int, error) {
	if err := r.verifyCourseOwner(ctx, authorID, courseID); err != nil {
		return 0, err
	}
	lessonType := req.Type
	if lessonType == "" {
		lessonType = "video"
	}
	var id int
	err := r.pool.QueryRow(ctx,
		"INSERT INTO lessons (module_id, course_id, name, type, is_free) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		moduleID, courseID, req.Name, lessonType, req.IsFree).Scan(&id)
	return id, err
}

func (r *studioRepo) UpdateLesson(ctx context.Context, authorID, courseID, lessonID int, req domain.UpdateLessonRequest) error {
	if err := r.verifyCourseOwner(ctx, authorID, courseID); err != nil {
		return err
	}

	sets := []string{}
	args := []any{}
	idx := 1

	if req.Name != nil {
		sets = append(sets, fmt.Sprintf("name = $%d", idx))
		args = append(args, *req.Name)
		idx++
	}
	if req.Type != nil {
		sets = append(sets, fmt.Sprintf("type = $%d", idx))
		args = append(args, *req.Type)
		idx++
	}
	if req.IsFree != nil {
		sets = append(sets, fmt.Sprintf("is_free = $%d", idx))
		args = append(args, *req.IsFree)
		idx++
	}
	if req.DurationMinutes != nil {
		sets = append(sets, fmt.Sprintf("duration_minutes = $%d", idx))
		args = append(args, *req.DurationMinutes)
		idx++
	}
	if req.SortOrder != nil {
		sets = append(sets, fmt.Sprintf("sort_order = $%d", idx))
		args = append(args, *req.SortOrder)
		idx++
	}

	if len(sets) == 0 {
		return nil
	}

	query := fmt.Sprintf("UPDATE lessons SET %s WHERE id = $%d AND course_id = $%d",
		strings.Join(sets, ", "), idx, idx+1)
	args = append(args, lessonID, courseID)

	tag, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrLessonNotFound
	}
	return nil
}

func (r *studioRepo) DeleteLesson(ctx context.Context, authorID, courseID, lessonID int) error {
	if err := r.verifyCourseOwner(ctx, authorID, courseID); err != nil {
		return err
	}
	tag, err := r.pool.Exec(ctx, "DELETE FROM lessons WHERE id = $1 AND course_id = $2", lessonID, courseID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrLessonNotFound
	}
	return nil
}

// --- Stats ---

func (r *studioRepo) GetStats(ctx context.Context, authorID int) (*domain.StudioStats, error) {
	var s domain.StudioStats
	err := r.pool.QueryRow(ctx, `
		SELECT
			COALESCE((SELECT SUM(c.price) FROM enrollments e JOIN courses c ON e.course_id = c.id
				WHERE c.author_id = $1 AND e.created_at >= NOW() - INTERVAL '30 days'), 0)::float,
			COALESCE((SELECT COUNT(*) FROM enrollments e JOIN courses c ON e.course_id = c.id
				WHERE c.author_id = $1 AND e.created_at >= NOW() - INTERVAL '30 days'), 0)::int,
			COALESCE((SELECT COUNT(DISTINCT e.user_id) FROM enrollments e JOIN courses c ON e.course_id = c.id
				WHERE c.author_id = $1), 0)::int,
			COALESCE((SELECT AVG(rv.rating) FROM reviews rv JOIN courses c ON rv.course_id = c.id
				WHERE c.author_id = $1), 0)::float
	`, authorID).Scan(&s.Revenue30d, &s.NewStudents30d, &s.TotalStudents, &s.AvgRating)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// --- Students ---

func (r *studioRepo) ListStudents(ctx context.Context, authorID int) ([]domain.StudioStudent, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			e.user_id, c.title,
			CASE WHEN COALESCE(sub_l.cnt, 0) = 0 THEN 0
				ELSE (COALESCE(sub_p.done, 0)::float / sub_l.cnt * 100) END,
			COALESCE(sub_p.last_active, e.created_at)
		FROM enrollments e
		JOIN courses c ON e.course_id = c.id
		LEFT JOIN LATERAL (SELECT COUNT(*) AS cnt FROM lessons WHERE course_id = c.id) sub_l ON true
		LEFT JOIN LATERAL (
			SELECT COUNT(*) AS done, MAX(completed_at) AS last_active
			FROM lesson_progress WHERE user_id = e.user_id AND course_id = c.id
		) sub_p ON true
		WHERE c.author_id = $1
		ORDER BY e.created_at DESC
	`, authorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []domain.StudioStudent
	for rows.Next() {
		var s domain.StudioStudent
		if err := rows.Scan(&s.UserID, &s.CourseName, &s.Progress, &s.LastActive); err != nil {
			return nil, err
		}
		students = append(students, s)
	}
	if students == nil {
		students = []domain.StudioStudent{}
	}
	return students, nil
}

// --- Income ---

func (r *studioRepo) GetIncome(ctx context.Context, authorID int) (*domain.StudioIncome, error) {
	var inc domain.StudioIncome
	err := r.pool.QueryRow(ctx, `
		SELECT
			COALESCE((SELECT SUM(c.price) FROM enrollments e JOIN courses c ON e.course_id = c.id WHERE c.author_id = $1), 0)
			- COALESCE((SELECT SUM(amount) FROM payouts WHERE author_id = $1 AND status = 'completed'), 0),
			COALESCE((SELECT SUM(c.price) FROM enrollments e JOIN courses c ON e.course_id = c.id
				WHERE c.author_id = $1 AND e.created_at >= date_trunc('month', NOW())), 0),
			COALESCE((SELECT COUNT(*) FROM enrollments e JOIN courses c ON e.course_id = c.id WHERE c.author_id = $1), 0)
	`, authorID).Scan(&inc.Available, &inc.MonthlyRevenue, &inc.SalesCount)
	if err != nil {
		return nil, err
	}
	return &inc, nil
}

func (r *studioRepo) ListPayouts(ctx context.Context, authorID int) ([]domain.Payout, error) {
	rows, err := r.pool.Query(ctx,
		"SELECT id, amount, status, created_at, completed_at FROM payouts WHERE author_id = $1 ORDER BY created_at DESC",
		authorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payouts []domain.Payout
	for rows.Next() {
		var p domain.Payout
		if err := rows.Scan(&p.ID, &p.Amount, &p.Status, &p.CreatedAt, &p.CompletedAt); err != nil {
			return nil, err
		}
		payouts = append(payouts, p)
	}
	if payouts == nil {
		payouts = []domain.Payout{}
	}
	return payouts, nil
}

// --- Reviews ---

func (r *studioRepo) ListReviews(ctx context.Context, authorID int) ([]domain.StudioReview, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT rv.id, c.title, rv.name, rv.initials, rv.text, rv.rating, rv.reply, rv.created_at
		FROM reviews rv
		JOIN courses c ON rv.course_id = c.id
		WHERE c.author_id = $1
		ORDER BY rv.created_at DESC
	`, authorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []domain.StudioReview
	for rows.Next() {
		var rv domain.StudioReview
		if err := rows.Scan(&rv.ID, &rv.CourseName, &rv.Name, &rv.Initials, &rv.Text, &rv.Rating, &rv.Reply, &rv.CreatedAt); err != nil {
			return nil, err
		}
		reviews = append(reviews, rv)
	}
	if reviews == nil {
		reviews = []domain.StudioReview{}
	}
	return reviews, nil
}

func (r *studioRepo) ReplyToReview(ctx context.Context, authorID, reviewID int, reply string) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE reviews SET reply = $1, replied_at = NOW()
		WHERE id = $2 AND course_id IN (SELECT id FROM courses WHERE author_id = $3)
	`, reply, reviewID, authorID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrReviewNotFound
	}
	return nil
}

// --- Helpers ---

func (r *studioRepo) verifyCourseOwner(ctx context.Context, authorID, courseID int) error {
	var exists bool
	err := r.pool.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM courses WHERE id = $1 AND author_id = $2)",
		courseID, authorID).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return domain.ErrCourseNotOwned
	}
	return nil
}
