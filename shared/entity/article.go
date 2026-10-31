package entity

import "github.com/google/uuid"

type ArticleStatus string

const (
	ArticleStatusPending  ArticleStatus = "pending"
	ArticleStatusApproved ArticleStatus = "approved"
	ArticleStatusRejected ArticleStatus = "rejected"
)

type Article struct {
	ID         uuid.UUID
	AuthorUUID uuid.UUID
	AuthorName string
	Title      string
	Content    string
	Status     ArticleStatus
	ImageURL   string
	CreatedAt  string
}
