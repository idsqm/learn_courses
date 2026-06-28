package domain

import "time"

type Certificate struct {
	ID         int       `json:"id"`
	CourseID   int       `json:"course_id"`
	CourseName string    `json:"course_name"`
	IssuedAt   time.Time `json:"issued_at"`
}
