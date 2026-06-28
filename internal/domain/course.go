package domain

type CourseFilter struct {
	Categories []string
	Level      string
	PriceMin   *float64
	PriceMax   *float64
	RatingMin  *float64
	Sort       string
	Search     string
	Page       int
	PerPage    int
}

type CourseListItem struct {
	ID            int      `json:"id"`
	Title         string   `json:"title"`
	Author        string   `json:"author"`
	Initials      string   `json:"initials"`
	Category      string   `json:"category"`
	Level         string   `json:"level"`
	Price         float64  `json:"price"`
	OldPrice      *float64 `json:"old_price,omitempty"`
	Rating        float64  `json:"rating"`
	ReviewsCount  int      `json:"reviews_count"`
	StudentsCount int      `json:"students_count"`
	Hours         int      `json:"hours"`
	LessonsCount  int      `json:"lessons_count"`
	Color1        string   `json:"color_1"`
	Color2        string   `json:"color_2"`
	Tag           *string  `json:"tag,omitempty"`
}

type CourseDetail struct {
	ID            int        `json:"id"`
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	Category      string     `json:"category"`
	Level         string     `json:"level"`
	Price         float64    `json:"price"`
	OldPrice      *float64   `json:"old_price,omitempty"`
	Rating        float64    `json:"rating"`
	ReviewsCount  int        `json:"reviews_count"`
	StudentsCount int        `json:"students_count"`
	Hours         int        `json:"hours"`
	LessonsCount  int        `json:"lessons_count"`
	Color1        string     `json:"color_1"`
	Color2        string     `json:"color_2"`
	Tag           *string    `json:"tag,omitempty"`
	PreviewURL    *string    `json:"preview_url,omitempty"`
	LearnItems    []string   `json:"learn_items"`
	Curriculum    []Module   `json:"curriculum"`
	Includes      []string   `json:"includes"`
	Reviews       []Review   `json:"reviews"`
	Author        AuthorInfo `json:"author"`
}

type Module struct {
	Title        string   `json:"title"`
	Duration     string   `json:"duration"`
	LessonsCount int      `json:"lessons_count"`
	Lessons      []Lesson `json:"lessons"`
}

type Lesson struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Duration string `json:"duration"`
	IsFree   bool   `json:"is_free"`
}
