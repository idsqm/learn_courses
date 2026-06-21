package domain

type EnrolledCourse struct {
	Course       CourseListItem `json:"course"`
	Progress     float64       `json:"progress"`
	DoneLessons  int           `json:"done_lessons"`
	TotalLessons int           `json:"total_lessons"`
	LastLesson   *string       `json:"last_lesson,omitempty"`
	Completed    bool          `json:"completed"`
}
