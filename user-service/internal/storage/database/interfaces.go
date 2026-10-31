package database

import (
	"context"

	"github.com/google/uuid"
	"shared/entity"
	"user-service/api/dto"
)

type UserDatabase interface {
	CreateUser(ctx context.Context, user entity.User) error
	GetUserByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	UpdateUser(ctx context.Context, id uuid.UUID, updateInfo dto.UpdateProfileBody) error
}

type ToothDatabase interface {
	CreateInitialTeeth(ctx context.Context, userUUID uuid.UUID) ([]entity.Tooth, error)
	GetTeethByUserUUID(ctx context.Context, userUUID uuid.UUID) ([]entity.Tooth, error)
	UpdateTooth(ctx context.Context, tooth entity.Tooth) error
}

type BrushingTimeDatabase interface {
	CreateBrushingTime(ctx context.Context, brushingTime entity.BrushTime) error
	GetBrushingTimesByUserUUID(ctx context.Context, userUUID uuid.UUID) ([]entity.BrushTime, error)
	UpdateBrushingTime(ctx context.Context, brushingTime entity.BrushTime) error
	ReplaceAllBrushingTimes(ctx context.Context, userUUID uuid.UUID, times []entity.BrushTime) error
}

type ToothNoteDatabase interface {
	CreateToothNote(ctx context.Context, note entity.ToothNote) (*entity.ToothNote, error)
	GetToothNotesByPatient(ctx context.Context, doctorUUID, patientUUID uuid.UUID) ([]entity.ToothNote, error)
	UpdateToothNote(ctx context.Context, noteID, doctorUUID uuid.UUID, newNote string) error
	DeleteToothNote(ctx context.Context, noteID, doctorUUID uuid.UUID) error
}

type RecommendationDatabase interface {
	CreateRecommendation(ctx context.Context, rec entity.Recommendation) (*entity.Recommendation, error)
	GetRecommendationsByPatient(ctx context.Context, doctorUUID, patientUUID uuid.UUID) ([]entity.Recommendation, error)
}

type ChatDatabase interface {
	SaveMessage(ctx context.Context, msg entity.ChatMessage) (*entity.ChatMessage, error)
	GetHistory(ctx context.Context, userUUID uuid.UUID) ([]entity.ChatMessage, error)
}

type ConversationDatabase interface {
	GetOrCreate(ctx context.Context, doctorUUID, patientUUID uuid.UUID) (*entity.Conversation, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Conversation, error)
	GetByDoctor(ctx context.Context, doctorUUID uuid.UUID) ([]entity.Conversation, error)
	GetByPatientWithDetails(ctx context.Context, patientUUID uuid.UUID) ([]dto.ConversationResponse, error)
}

type DoctorProfileDatabase interface {
	GetAll(ctx context.Context) ([]entity.DoctorProfile, error)
	Upsert(ctx context.Context, profile entity.DoctorProfile) error
}

type DirectMessageDatabase interface {
	Create(ctx context.Context, msg entity.DirectMessage) (*entity.DirectMessage, error)
	GetByConversation(ctx context.Context, conversationID uuid.UUID) ([]entity.DirectMessage, error)
}

type CalendarDatabase interface {
	CreateCalendarEvent(ctx context.Context, event entity.CalendarEvent) (*entity.CalendarEvent, error)
	GetCalendarEventsByUserUUID(ctx context.Context, userUUID uuid.UUID) ([]entity.CalendarEvent, error)
	DeleteCalendarEvent(ctx context.Context, eventID, userUUID uuid.UUID) error
}

type ArticleDatabase interface {
	Create(ctx context.Context, article entity.Article) (*entity.Article, error)
	GetApproved(ctx context.Context) ([]entity.Article, error)
	GetAll(ctx context.Context) ([]entity.Article, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.ArticleStatus) error
}
