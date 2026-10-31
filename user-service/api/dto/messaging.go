package dto

import (
	"time"

	"github.com/google/uuid"
)

type SendDirectMessageBody struct {
	Content string `json:"Content" validate:"required,min=1,max=2000"`
}

type StartConversationBody struct {
	DoctorUUID uuid.UUID `json:"DoctorUUID" validate:"required"`
}

type MessageResponse struct {
	ID             uuid.UUID `json:"ID"`
	ConversationID uuid.UUID `json:"ConversationID"`
	SenderUUID     uuid.UUID `json:"SenderUUID"`
	Content        string    `json:"Content"`
	IsMine         bool      `json:"IsMine"`
	CreatedAt      time.Time `json:"CreatedAt"`
}

type ConversationResponse struct {
	ID              uuid.UUID `json:"ID"`
	DoctorUUID      uuid.UUID `json:"DoctorUUID"`
	DoctorName      string    `json:"DoctorName"`
	LastMessage     string    `json:"LastMessage"`
	LastMessageTime string    `json:"LastMessageTime"`
}
