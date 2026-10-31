package usecase

import (
	"context"

	"shared/entity"
	sharedjwt "shared/jwt"
	"user-service/api/dto"
	"user-service/internal/storage/database"
)

const systemPrompt = `You are a dental health assistant built into the Zub dental health app.
Your exclusive purpose is to help users with topics related to dentistry and oral health.
If a user asks about any topic not related to dentistry or oral health, politely decline and redirect them to dental topics.
Never provide a specific medical diagnosis. For any serious or urgent dental concern, always recommend visiting a licensed dentist.
Keep your responses clear, friendly, and concise.
Always respond in the same language the user writes in.`

type AiChatUseCase struct {
	chatRepo database.ChatDatabase
	ai       AIProvider
}

func NewAiChatUseCase(chatRepo database.ChatDatabase, ai AIProvider) *AiChatUseCase {
	return &AiChatUseCase{chatRepo: chatRepo, ai: ai}
}

func (uc *AiChatUseCase) SendMessage(ctx context.Context, body dto.SendMessageBody) ([]entity.ChatMessage, error) {
	claims := sharedjwt.GetClaimsFromCtx(ctx)
	if claims == nil {
		return nil, ErrNoAuthClaims
	}

	history, err := uc.chatRepo.GetHistory(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}

	reply, err := uc.ai.Complete(ctx, systemPrompt, history, body.Message)
	if err != nil {
		return nil, err
	}

	savedUser, err := uc.chatRepo.SaveMessage(ctx, entity.ChatMessage{
		UserUUID: claims.UserID,
		Role:     entity.ChatRoleUser,
		Content:  body.Message,
	})
	if err != nil {
		return nil, err
	}

	savedAssistant, err := uc.chatRepo.SaveMessage(ctx, entity.ChatMessage{
		UserUUID: claims.UserID,
		Role:     entity.ChatRoleAssistant,
		Content:  reply,
	})
	if err != nil {
		return nil, err
	}

	return []entity.ChatMessage{*savedUser, *savedAssistant}, nil
}

func (uc *AiChatUseCase) Greet(ctx context.Context, body dto.GreetBody) (*entity.ChatMessage, error) {
	claims := sharedjwt.GetClaimsFromCtx(ctx)
	if claims == nil {
		return nil, ErrNoAuthClaims
	}

	prompt := "Greet the user warmly as a dental health assistant. Respond in " + body.Language + ". Be brief and friendly."
	reply, err := uc.ai.Complete(ctx, systemPrompt, nil, prompt)
	if err != nil {
		return nil, err
	}

	saved, err := uc.chatRepo.SaveMessage(ctx, entity.ChatMessage{
		UserUUID: claims.UserID,
		Role:     entity.ChatRoleAssistant,
		Content:  reply,
	})
	if err != nil {
		return nil, err
	}

	return saved, nil
}

func (uc *AiChatUseCase) GetHistory(ctx context.Context) ([]entity.ChatMessage, error) {
	claims := sharedjwt.GetClaimsFromCtx(ctx)
	if claims == nil {
		return nil, ErrNoAuthClaims
	}
	return uc.chatRepo.GetHistory(ctx, claims.UserID)
}
