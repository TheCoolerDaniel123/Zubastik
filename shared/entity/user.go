package entity

import "github.com/google/uuid"

type (
	UserRole int
	Gender   string
)

const (
	RoleUser   UserRole = 0
	RoleDoctor UserRole = 1
	RoleAdmin  UserRole = 2

	Male   Gender = "Male"
	Female Gender = "Female"
)

type User struct {
	UUID  uuid.UUID
	Phone string
	Role  UserRole // 0 = обычный пользователь

	Name   string
	Age    *int
	Gender *Gender

	Teeth      []Tooth
	BrushTimes []BrushTime
}
