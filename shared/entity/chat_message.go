package entity

import (
	"time"

	"github.com/google/uuid"
)

const (
	ChatRoleUser      = "user"
	ChatRoleAssistant = "assistant"
)

type ChatMessage struct {
	ID        uuid.UUID
	UserUUID  uuid.UUID
	Role      string
	Content   string
	CreatedAt time.Time
}
