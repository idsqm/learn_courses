package domain

type AuthorInfo struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	Initials        string `json:"initials"`
	Subtitle        string `json:"subtitle"`
	Bio             string `json:"bio"`
	CoursesCount    int    `json:"courses_count"`
	YearsExperience int    `json:"years_experience"`
}

type ApplyAuthorRequest struct {
	Name            string `json:"name"`
	Subtitle        string `json:"subtitle"`
	Bio             string `json:"bio"`
	YearsExperience int    `json:"years_experience"`
}
