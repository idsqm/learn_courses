package domain

type Review struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Initials string `json:"initials"`
	Text     string `json:"text"`
	Rating   int    `json:"rating"`
}

type CreateReviewRequest struct {
	Name     string `json:"name"`
	Initials string `json:"initials"`
	Text     string `json:"text"`
	Rating   int    `json:"rating"`
}
