package domain

import "time"

type StudioCourse struct {
	ID            int     `json:"id"`
	Title         string  `json:"title"`
	Category      string  `json:"category"`
	LessonsCount  int     `json:"lessons_count"`
	StudentsCount int     `json:"students_count"`
	Revenue30d    float64 `json:"revenue_30d"`
	Status        string  `json:"status"`
}

type StudioCourseDetail struct {
	ID          int      `json:"id"`
	Title       string   `json:"title"`
	Subtitle    string   `json:"subtitle"`
	Description string   `json:"description"`
	CategoryID  int      `json:"category_id"`
	Category    string   `json:"category"`
	Level       string   `json:"level"`
	Price       float64  `json:"price"`
	OldPrice    *float64 `json:"old_price,omitempty"`
	IsFree      bool     `json:"is_free"`
	Color1      string   `json:"color_1"`
	Color2      string   `json:"color_2"`
	Tag         *string  `json:"tag,omitempty"`
	Published   bool     `json:"published"`
	Curriculum  []Module `json:"curriculum"`
}

type CreateCourseRequest struct {
	Title       string   `json:"title"`
	Subtitle    string   `json:"subtitle"`
	CategoryID  int      `json:"category_id"`
	Level       string   `json:"level"`
	Description string   `json:"description"`
	Price       float64  `json:"price"`
	OldPrice    *float64 `json:"old_price"`
	IsFree      bool     `json:"is_free"`
	Color1      string   `json:"color_1"`
	Color2      string   `json:"color_2"`
}

type UpdateCourseRequest struct {
	Title       *string  `json:"title"`
	Subtitle    *string  `json:"subtitle"`
	CategoryID  *int     `json:"category_id"`
	Level       *string  `json:"level"`
	Description *string  `json:"description"`
	Price       *float64 `json:"price"`
	OldPrice    *float64 `json:"old_price"`
	IsFree      *bool    `json:"is_free"`
	Color1      *string  `json:"color_1"`
	Color2      *string  `json:"color_2"`
	Tag         *string  `json:"tag"`
}

type StudioStats struct {
	Revenue30d     float64 `json:"revenue_30d"`
	NewStudents30d int     `json:"new_students_30d"`
	TotalStudents  int     `json:"total_students"`
	AvgRating      float64 `json:"avg_rating"`
}

type StudioStudent struct {
	UserID     string    `json:"user_id"`
	CourseName string    `json:"course"`
	Progress   float64   `json:"progress"`
	LastActive time.Time `json:"last_active"`
}

type StudioIncome struct {
	Available      float64 `json:"available"`
	MonthlyRevenue float64 `json:"monthly_revenue"`
	SalesCount     int     `json:"sales_count"`
}

type Payout struct {
	ID          int        `json:"id"`
	Amount      float64    `json:"amount"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type CreateModuleRequest struct {
	Title     string `json:"title"`
	SortOrder int    `json:"sort_order"`
}

type UpdateModuleRequest struct {
	Title     *string `json:"title"`
	SortOrder *int    `json:"sort_order"`
}

type CreateLessonRequest struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	IsFree bool   `json:"is_free"`
}

type UpdateLessonRequest struct {
	Name            *string `json:"name"`
	Type            *string `json:"type"`
	IsFree          *bool   `json:"is_free"`
	DurationMinutes *int    `json:"duration_minutes"`
	SortOrder       *int    `json:"sort_order"`
}

type StudioReview struct {
	ID         int       `json:"id"`
	CourseName string    `json:"course_name"`
	Name       string    `json:"name"`
	Initials   string    `json:"initials"`
	Text       string    `json:"text"`
	Rating     int       `json:"rating"`
	Reply      *string   `json:"reply,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}
