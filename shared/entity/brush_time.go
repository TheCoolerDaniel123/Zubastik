package entity

import (
	"time"

	"github.com/google/uuid"
)

type BrushTime struct {
	UserUUID uuid.UUID
	Time     time.Time
}
