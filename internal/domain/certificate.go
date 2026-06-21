package domain

import "time"

type Certificate struct {
	ID         string    `json:"id"`
	CourseID   string    `json:"course_id"`
	CourseName string    `json:"course_name"`
	IssuedAt   time.Time `json:"issued_at"`
}
