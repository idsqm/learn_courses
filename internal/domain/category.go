package domain

type Category struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Abbreviation string `json:"abbreviation"`
	CoursesCount int    `json:"courses_count"`
	Color        string `json:"color"`
}
