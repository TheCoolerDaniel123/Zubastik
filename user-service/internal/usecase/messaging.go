package usecase

import (
	"context"

	"github.com/google/uuid"
	"shared/entity"
	sharedjwt "shared/jwt"
	"user-service/api/dto"
	"user-service/internal/storage/database"
)

type MessagingUseCase struct {
	convRepo database.ConversationDatabase
	msgRepo  database.DirectMessageDatabase
}

func NewMessagingUseCase(convRepo database.ConversationDatabase, msgRepo database.DirectMessageDatabase) *MessagingUseCase {
	return &MessagingUseCase{convRepo: convRepo, msgRepo: msgRepo}
}

func (uc *MessagingUseCase) GetOrCreateConversation(ctx context.Context, patientUUID uuid.UUID) (*entity.Conversation, error) {
	claims := sharedjwt.GetClaimsFromCtx(ctx)
	if claims == nil {
		return nil, ErrNoAuthClaims
	}
	return uc.convRepo.GetOrCreate(ctx, claims.UserID, patientUUID)
}

func (uc *MessagingUseCase) GetOrCreateConversationAsPatient(ctx context.Context, doctorUUID uuid.UUID) (*entity.Conversation, error) {
	claims := sharedjwt.GetClaimsFromCtx(ctx)
	if claims == nil {
		return nil, ErrNoAuthClaims
	}
	return uc.convRepo.GetOrCreate(ctx, doctorUUID, claims.UserID)
}

func (uc *MessagingUseCase) GetDoctorConversations(ctx context.Context) ([]entity.Conversation, error) {
	claims := sharedjwt.GetClaimsFromCtx(ctx)
	if claims == nil {
		return nil, ErrNoAuthClaims
	}
	return uc.convRepo.GetByDoctor(ctx, claims.UserID)
}

func (uc *MessagingUseCase) GetPatientConversations(ctx context.Context) ([]dto.ConversationResponse, error) {
	claims := sharedjwt.GetClaimsFromCtx(ctx)
	if claims == nil {
		return nil, ErrNoAuthClaims
	}
	return uc.convRepo.GetByPatientWithDetails(ctx, claims.UserID)
}

func (uc *MessagingUseCase) SendMessage(ctx context.Context, conversationID uuid.UUID, body dto.SendDirectMessageBody) (*entity.DirectMessage, error) {
	claims := sharedjwt.GetClaimsFromCtx(ctx)
	if claims == nil {
		return nil, ErrNoAuthClaims
	}

	conv, err := uc.convRepo.GetByID(ctx, conversationID)
	if err != nil {
		return nil, err
	}

	if claims.UserID != conv.DoctorUUID && claims.UserID != conv.PatientUUID {
		return nil, ErrForbidden
	}

	return uc.msgRepo.Create(ctx, entity.DirectMessage{
		ConversationID: conversationID,
		SenderUUID:     claims.UserID,
		Content:        body.Content,
	})
}

func (uc *MessagingUseCase) GetMessages(ctx context.Context, conversationID uuid.UUID) ([]dto.MessageResponse, error) {
	claims := sharedjwt.GetClaimsFromCtx(ctx)
	if claims == nil {
		return nil, ErrNoAuthClaims
	}

	conv, err := uc.convRepo.GetByID(ctx, conversationID)
	if err != nil {
		return nil, err
	}

	if claims.UserID != conv.DoctorUUID && claims.UserID != conv.PatientUUID {
		return nil, ErrForbidden
	}

	msgs, err := uc.msgRepo.GetByConversation(ctx, conversationID)
	if err != nil {
		return nil, err
	}

	result := make([]dto.MessageResponse, len(msgs))
	for i, m := range msgs {
		result[i] = dto.MessageResponse{
			ID:             m.ID,
			ConversationID: m.ConversationID,
			SenderUUID:     m.SenderUUID,
			Content:        m.Content,
			IsMine:         m.SenderUUID == claims.UserID,
			CreatedAt:      m.CreatedAt,
		}
	}
	return result, nil
}
