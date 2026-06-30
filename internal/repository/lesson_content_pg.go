package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/andruho/courses/internal/domain"
)

type LessonContentRepository interface {
	GetLessonDetail(ctx context.Context, lessonID int) (*domain.LessonDetail, error)
	SaveContent(ctx context.Context, lessonID int, req domain.SaveLessonContentRequest) error
	ListQuestions(ctx context.Context, lessonID int) ([]domain.QuizQuestion, error)
	CreateQuestion(ctx context.Context, lessonID int, req domain.SaveQuizQuestionRequest) (int, error)
	UpdateQuestion(ctx context.Context, questionID int, req domain.SaveQuizQuestionRequest) error
	DeleteQuestion(ctx context.Context, questionID int) error
	VerifyLessonOwner(ctx context.Context, authorID, lessonID int) error
	VerifyCourseOwner(ctx context.Context, authorID, courseID int) error
	VerifyQuestionOwner(ctx context.Context, authorID, questionID int) error
}

type lessonContentRepo struct {
	pool *pgxpool.Pool
}

func NewLessonContentRepository(pool *pgxpool.Pool) LessonContentRepository {
	return &lessonContentRepo{pool: pool}
}

func (r *lessonContentRepo) GetLessonDetail(ctx context.Context, lessonID int) (*domain.LessonDetail, error) {
	var d domain.LessonDetail
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, type, duration_minutes, is_free, sort_order
		FROM lessons WHERE id = $1
	`, lessonID).Scan(&d.ID, &d.Name, &d.Type, &d.DurationMinutes, &d.IsFree, &d.SortOrder)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrLessonNotFound
		}
		return nil, err
	}

	var lc domain.LessonContent
	err = r.pool.QueryRow(ctx, `
		SELECT id, lesson_id, video_url, body
		FROM lesson_content WHERE lesson_id = $1
	`, lessonID).Scan(&lc.ID, &lc.LessonID, &lc.VideoURL, &lc.Body)
	if err != nil && err != pgx.ErrNoRows {
		return nil, err
	}
	if err == nil {
		d.Content = &lc
	}

	if d.Type == "quiz" {
		questions, err := r.ListQuestions(ctx, lessonID)
		if err != nil {
			return nil, err
		}
		d.Questions = questions
	}

	return &d, nil
}

func (r *lessonContentRepo) SaveContent(ctx context.Context, lessonID int, req domain.SaveLessonContentRequest) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO lesson_content (lesson_id, video_url, body)
		VALUES ($1, $2, $3)
		ON CONFLICT (lesson_id)
		DO UPDATE SET video_url = EXCLUDED.video_url, body = EXCLUDED.body
	`, lessonID, req.VideoURL, req.Body)
	return err
}

func (r *lessonContentRepo) ListQuestions(ctx context.Context, lessonID int) ([]domain.QuizQuestion, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, lesson_id, text, question_type, points, sort_order
		FROM quiz_questions WHERE lesson_id = $1 ORDER BY sort_order
	`, lessonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questions []domain.QuizQuestion
	for rows.Next() {
		var q domain.QuizQuestion
		if err := rows.Scan(&q.ID, &q.LessonID, &q.Text, &q.QuestionType, &q.Points, &q.SortOrder); err != nil {
			return nil, err
		}
		questions = append(questions, q)
	}

	for i := range questions {
		optRows, err := r.pool.Query(ctx, `
			SELECT id, text, is_correct, sort_order
			FROM quiz_options WHERE question_id = $1 ORDER BY sort_order
		`, questions[i].ID)
		if err != nil {
			return nil, err
		}

		var options []domain.QuizOption
		for optRows.Next() {
			var o domain.QuizOption
			if err := optRows.Scan(&o.ID, &o.Text, &o.IsCorrect, &o.SortOrder); err != nil {
				optRows.Close()
				return nil, err
			}
			options = append(options, o)
		}
		optRows.Close()
		if options == nil {
			options = []domain.QuizOption{}
		}
		questions[i].Options = options
	}

	if questions == nil {
		questions = []domain.QuizQuestion{}
	}
	return questions, nil
}

func (r *lessonContentRepo) CreateQuestion(ctx context.Context, lessonID int, req domain.SaveQuizQuestionRequest) (int, error) {
	var id int
	err := r.pool.QueryRow(ctx, `
		INSERT INTO quiz_questions (lesson_id, text, question_type, points, sort_order)
		VALUES ($1, $2, $3, $4, (SELECT COALESCE(MAX(sort_order), 0) + 1 FROM quiz_questions WHERE lesson_id = $1))
		RETURNING id
	`, lessonID, req.Text, req.QuestionType, req.Points).Scan(&id)
	if err != nil {
		return 0, err
	}

	for i, opt := range req.Options {
		_, err := r.pool.Exec(ctx, `
			INSERT INTO quiz_options (question_id, text, is_correct, sort_order)
			VALUES ($1, $2, $3, $4)
		`, id, opt.Text, opt.IsCorrect, i)
		if err != nil {
			return 0, err
		}
	}

	return id, nil
}

func (r *lessonContentRepo) UpdateQuestion(ctx context.Context, questionID int, req domain.SaveQuizQuestionRequest) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE quiz_questions SET text = $1, question_type = $2, points = $3
		WHERE id = $4
	`, req.Text, req.QuestionType, req.Points, questionID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrQuestionNotFound
	}

	_, err = r.pool.Exec(ctx, "DELETE FROM quiz_options WHERE question_id = $1", questionID)
	if err != nil {
		return err
	}

	for i, opt := range req.Options {
		_, err := r.pool.Exec(ctx, `
			INSERT INTO quiz_options (question_id, text, is_correct, sort_order)
			VALUES ($1, $2, $3, $4)
		`, questionID, opt.Text, opt.IsCorrect, i)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *lessonContentRepo) DeleteQuestion(ctx context.Context, questionID int) error {
	tag, err := r.pool.Exec(ctx, "DELETE FROM quiz_questions WHERE id = $1", questionID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrQuestionNotFound
	}
	return nil
}

func (r *lessonContentRepo) VerifyLessonOwner(ctx context.Context, authorID, lessonID int) error {
	var exists bool
	err := r.pool.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM lessons l JOIN courses c ON l.course_id = c.id WHERE l.id = $1 AND c.author_id = $2)",
		lessonID, authorID).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return domain.ErrLessonNotFound
	}
	return nil
}

func (r *lessonContentRepo) VerifyCourseOwner(ctx context.Context, authorID, courseID int) error {
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

func (r *lessonContentRepo) VerifyQuestionOwner(ctx context.Context, authorID, questionID int) error {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM quiz_questions q
			JOIN lessons l ON q.lesson_id = l.id
			JOIN courses c ON l.course_id = c.id
			WHERE q.id = $1 AND c.author_id = $2
		)
	`, questionID, authorID).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return domain.ErrQuestionNotFound
	}
	return nil
}
