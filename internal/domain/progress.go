package domain

type CourseProgress struct {
	CourseID     string  `json:"course_id"`
	Progress     float64 `json:"progress"`
	DoneLessons  int     `json:"done_lessons"`
	TotalLessons int     `json:"total_lessons"`
	Completed    bool    `json:"completed"`
}

type UserStats struct {
	CoursesInProgress int `json:"courses_in_progress"`
	TotalStudyTime    int `json:"total_study_time"`
	CertificatesCount int `json:"certificates_count"`
	StudyStreakDays   int `json:"study_streak_days"`
}
