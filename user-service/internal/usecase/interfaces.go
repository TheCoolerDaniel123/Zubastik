package usecase

import (
	"context"

	"github.com/google/uuid"
	"shared/entity"
	"user-service/api/dto"
)

type AIProvider interface {
	Complete(ctx context.Context, systemPrompt string, history []entity.ChatMessage, userMessage string) (string, error)
}

type UserUseCaseInterface interface {
	CreateUserInitially(ctx context.Context, user entity.User) error
	GetUserByUUID(ctx context.Context) (*entity.User, error)
	UpdateProfile(ctx context.Context, body dto.UpdateProfileBody) error
	GetTeeth(ctx context.Context) ([]entity.Tooth, error)
	UpdateTeeth(ctx context.Context, body dto.UpdateTeethBody) error
	GetBrushSchedule(ctx context.Context) ([]entity.BrushTime, error)
	UpdateBrushSchedule(ctx context.Context, body dto.UpdateBrushScheduleBody) error
	GetCalendarEvents(ctx context.Context) ([]entity.CalendarEvent, error)
	CreateCalendarEvent(ctx context.Context, body dto.CreateCalendarEventBody) (*entity.CalendarEvent, error)
	DeleteCalendarEvent(ctx context.Context, eventID uuid.UUID) error
}

type MessagingUseCaseInterface interface {
	GetOrCreateConversation(ctx context.Context, patientUUID uuid.UUID) (*entity.Conversation, error)
	GetDoctorConversations(ctx context.Context) ([]entity.Conversation, error)
	GetPatientConversations(ctx context.Context) ([]dto.ConversationResponse, error)
	SendMessage(ctx context.Context, conversationID uuid.UUID, body dto.SendDirectMessageBody) (*entity.DirectMessage, error)
	GetMessages(ctx context.Context, conversationID uuid.UUID) ([]dto.MessageResponse, error)
}

type AiChatUseCaseInterface interface {
	SendMessage(ctx context.Context, body dto.SendMessageBody) (*entity.ChatMessage, error)
	GetHistory(ctx context.Context) ([]entity.ChatMessage, error)
}

type DoctorUseCaseInterface interface {
	GetPatientByUUID(ctx context.Context, patientUUID uuid.UUID) (*entity.User, error)
	GetPatientTeeth(ctx context.Context, patientUUID uuid.UUID) ([]entity.Tooth, error)
	CreateToothNote(ctx context.Context, note entity.ToothNote) (*entity.ToothNote, error)
	GetToothNotes(ctx context.Context, patientUUID uuid.UUID) ([]entity.ToothNote, error)
	UpdateToothNote(ctx context.Context, noteID uuid.UUID, newNote string) error
	DeleteToothNote(ctx context.Context, noteID uuid.UUID) error
	CreateRecommendation(ctx context.Context, rec entity.Recommendation) (*entity.Recommendation, error)
	GetRecommendations(ctx context.Context, patientUUID uuid.UUID) ([]entity.Recommendation, error)
}
