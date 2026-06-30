package domain

type LessonContent struct {
	ID       int     `json:"id"`
	LessonID int     `json:"lesson_id"`
	VideoURL *string `json:"video_url,omitempty"`
	Body     *string `json:"body,omitempty"`
}

type QuizQuestion struct {
	ID           int          `json:"id"`
	LessonID     int          `json:"lesson_id"`
	Text         string       `json:"text"`
	QuestionType string       `json:"question_type"`
	Points       int          `json:"points"`
	SortOrder    int          `json:"sort_order"`
	Options      []QuizOption `json:"options"`
}

type QuizOption struct {
	ID        int    `json:"id"`
	Text      string `json:"text"`
	IsCorrect bool   `json:"is_correct"`
	SortOrder int    `json:"sort_order"`
}

type LessonDetail struct {
	ID              int            `json:"id"`
	Name            string         `json:"name"`
	Type            string         `json:"type"`
	DurationMinutes int            `json:"duration_minutes"`
	IsFree          bool           `json:"is_free"`
	SortOrder       int            `json:"sort_order"`
	Content         *LessonContent `json:"content,omitempty"`
	Questions       []QuizQuestion `json:"questions,omitempty"`
}

type SaveLessonContentRequest struct {
	VideoURL *string `json:"video_url,omitempty"`
	Body     *string `json:"body,omitempty"`
}

type SaveQuizQuestionRequest struct {
	Text         string                  `json:"text"`
	QuestionType string                  `json:"question_type"`
	Points       int                     `json:"points"`
	Options      []SaveQuizOptionRequest `json:"options"`
}

type SaveQuizOptionRequest struct {
	Text      string `json:"text"`
	IsCorrect bool   `json:"is_correct"`
}
