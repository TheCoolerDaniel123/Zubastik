package entity

import (
	"time"

	"github.com/google/uuid"
)

type DirectMessage struct {
	ID             uuid.UUID
	ConversationID uuid.UUID
	SenderUUID     uuid.UUID
	Content        string
	CreatedAt      time.Time
}
