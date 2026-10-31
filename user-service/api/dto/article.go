package dto

type CreateArticleBody struct {
	Title    string `json:"title" validate:"required"`
	Content  string `json:"content" validate:"required"`
	ImageURL string `json:"image_url"`
}
