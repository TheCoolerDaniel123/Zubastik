package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"shared/entity"
	"user-service/api/dto"
	"user-service/internal/usecase"
)

// --- Mocks ---

type mockChatDB struct {
	SaveMessageFn func(ctx context.Context, msg entity.ChatMessage) (*entity.ChatMessage, error)
	GetHistoryFn  func(ctx context.Context, userUUID uuid.UUID) ([]entity.ChatMessage, error)
}

func (m *mockChatDB) SaveMessage(ctx context.Context, msg entity.ChatMessage) (*entity.ChatMessage, error) {
	return m.SaveMessageFn(ctx, msg)
}

func (m *mockChatDB) GetHistory(ctx context.Context, userUUID uuid.UUID) ([]entity.ChatMessage, error) {
	return m.GetHistoryFn(ctx, userUUID)
}

type mockAIProvider struct {
	CompleteFn func(ctx context.Context, systemPrompt string, history []entity.ChatMessage, userMessage string) (string, error)
}

func (m *mockAIProvider) Complete(ctx context.Context, systemPrompt string, history []entity.ChatMessage, userMessage string) (string, error) {
	return m.CompleteFn(ctx, systemPrompt, history, userMessage)
}

func newChatUC(chatDB *mockChatDB, ai *mockAIProvider) *usecase.AiChatUseCase {
	return usecase.NewAiChatUseCase(chatDB, ai)
}

// --- Tests: SendMessage ---

func TestChatSendMessage_Success(t *testing.T) {
	userID := uuid.New()
	existing := []entity.ChatMessage{
		{Role: entity.ChatRoleUser, Content: "previous question"},
		{Role: entity.ChatRoleAssistant, Content: "previous answer"},
	}

	var savedMessages []entity.ChatMessage
	chatDB := &mockChatDB{
		GetHistoryFn: func(ctx context.Context, id uuid.UUID) ([]entity.ChatMessage, error) {
			assert.Equal(t, userID, id)
			return existing, nil
		},
		SaveMessageFn: func(ctx context.Context, msg entity.ChatMessage) (*entity.ChatMessage, error) {
			savedMessages = append(savedMessages, msg)
			msg.ID = uuid.New()
			msg.CreatedAt = time.Now()
			return &msg, nil
		},
	}
	ai := &mockAIProvider{
		CompleteFn: func(ctx context.Context, systemPrompt string, history []entity.ChatMessage, userMessage string) (string, error) {
			assert.Len(t, history, 2)
			assert.Equal(t, "how to brush teeth?", userMessage)
			assert.NotEmpty(t, systemPrompt)
			return "brush for 2 minutes twice a day", nil
		},
	}

	uc := newChatUC(chatDB, ai)
	result, err := uc.SendMessage(ctxWithClaims(userID), dto.SendMessageBody{Message: "how to brush teeth?"})

	require.NoError(t, err)
	require.Len(t, result, 2)
	assert.Equal(t, entity.ChatRoleUser, result[0].Role)
	assert.Equal(t, "how to brush teeth?", result[0].Content)
	assert.Equal(t, entity.ChatRoleAssistant, result[1].Role)
	assert.Equal(t, "brush for 2 minutes twice a day", result[1].Content)

	require.Len(t, savedMessages, 2)
	assert.Equal(t, entity.ChatRoleUser, savedMessages[0].Role)
	assert.Equal(t, "how to brush teeth?", savedMessages[0].Content)
	assert.Equal(t, entity.ChatRoleAssistant, savedMessages[1].Role)
	assert.Equal(t, "brush for 2 minutes twice a day", savedMessages[1].Content)
}

func TestChatSendMessage_NoClaims(t *testing.T) {
	uc := newChatUC(&mockChatDB{}, &mockAIProvider{})
	result, err := uc.SendMessage(context.Background(), dto.SendMessageBody{Message: "hello"})
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, usecase.ErrNoAuthClaims))
}

func TestChatSendMessage_GetHistoryError(t *testing.T) {
	chatDB := &mockChatDB{
		GetHistoryFn: func(ctx context.Context, id uuid.UUID) ([]entity.ChatMessage, error) {
			return nil, errors.New("db error")
		},
	}
	uc := newChatUC(chatDB, &mockAIProvider{})
	result, err := uc.SendMessage(ctxWithClaims(uuid.New()), dto.SendMessageBody{Message: "hello"})
	assert.Nil(t, result)
	require.Error(t, err)
}

func TestChatSendMessage_AIError(t *testing.T) {
	chatDB := &mockChatDB{
		GetHistoryFn: func(ctx context.Context, id uuid.UUID) ([]entity.ChatMessage, error) {
			return nil, nil
		},
	}
	ai := &mockAIProvider{
		CompleteFn: func(ctx context.Context, sp string, h []entity.ChatMessage, msg string) (string, error) {
			return "", errors.New("AI unavailable")
		},
	}
	uc := newChatUC(chatDB, ai)
	result, err := uc.SendMessage(ctxWithClaims(uuid.New()), dto.SendMessageBody{Message: "hello"})
	assert.Nil(t, result)
	require.Error(t, err)
}

func TestChatSendMessage_SaveUserMsgError(t *testing.T) {
	callCount := 0
	chatDB := &mockChatDB{
		GetHistoryFn: func(ctx context.Context, id uuid.UUID) ([]entity.ChatMessage, error) {
			return nil, nil
		},
		SaveMessageFn: func(ctx context.Context, msg entity.ChatMessage) (*entity.ChatMessage, error) {
			callCount++
			if callCount == 1 {
				return nil, errors.New("save user msg failed")
			}
			return &msg, nil
		},
	}
	ai := &mockAIProvider{
		CompleteFn: func(ctx context.Context, sp string, h []entity.ChatMessage, msg string) (string, error) {
			return "reply", nil
		},
	}
	uc := newChatUC(chatDB, ai)
	result, err := uc.SendMessage(ctxWithClaims(uuid.New()), dto.SendMessageBody{Message: "hello"})
	assert.Nil(t, result)
	require.Error(t, err)
}

func TestChatSendMessage_SaveAssistantMsgError(t *testing.T) {
	callCount := 0
	chatDB := &mockChatDB{
		GetHistoryFn: func(ctx context.Context, id uuid.UUID) ([]entity.ChatMessage, error) {
			return nil, nil
		},
		SaveMessageFn: func(ctx context.Context, msg entity.ChatMessage) (*entity.ChatMessage, error) {
			callCount++
			if callCount == 2 {
				return nil, errors.New("save assistant msg failed")
			}
			msg.ID = uuid.New()
			return &msg, nil
		},
	}
	ai := &mockAIProvider{
		CompleteFn: func(ctx context.Context, sp string, h []entity.ChatMessage, msg string) (string, error) {
			return "reply", nil
		},
	}
	uc := newChatUC(chatDB, ai)
	result, err := uc.SendMessage(ctxWithClaims(uuid.New()), dto.SendMessageBody{Message: "hello"})
	assert.Nil(t, result)
	require.Error(t, err)
}

// --- Tests: GetHistory ---

func TestChatGetHistory_Success(t *testing.T) {
	userID := uuid.New()
	expected := []entity.ChatMessage{
		{ID: uuid.New(), UserUUID: userID, Role: entity.ChatRoleUser, Content: "question"},
		{ID: uuid.New(), UserUUID: userID, Role: entity.ChatRoleAssistant, Content: "answer"},
	}
	chatDB := &mockChatDB{
		GetHistoryFn: func(ctx context.Context, id uuid.UUID) ([]entity.ChatMessage, error) {
			assert.Equal(t, userID, id)
			return expected, nil
		},
	}
	uc := newChatUC(chatDB, &mockAIProvider{})
	result, err := uc.GetHistory(ctxWithClaims(userID))
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestChatGetHistory_NoClaims(t *testing.T) {
	uc := newChatUC(&mockChatDB{}, &mockAIProvider{})
	result, err := uc.GetHistory(context.Background())
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, usecase.ErrNoAuthClaims))
}

func TestChatGetHistory_DBError(t *testing.T) {
	chatDB := &mockChatDB{
		GetHistoryFn: func(ctx context.Context, id uuid.UUID) ([]entity.ChatMessage, error) {
			return nil, errors.New("db error")
		},
	}
	uc := newChatUC(chatDB, &mockAIProvider{})
	result, err := uc.GetHistory(ctxWithClaims(uuid.New()))
	assert.Nil(t, result)
	require.Error(t, err)
}

func TestChatGetHistory_EmptyHistory(t *testing.T) {
	chatDB := &mockChatDB{
		GetHistoryFn: func(ctx context.Context, id uuid.UUID) ([]entity.ChatMessage, error) {
			return nil, nil
		},
	}
	uc := newChatUC(chatDB, &mockAIProvider{})
	result, err := uc.GetHistory(ctxWithClaims(uuid.New()))
	require.NoError(t, err)
	assert.Nil(t, result)
}

// --- Tests: Greet ---

func TestChatGreet_Success(t *testing.T) {
	userID := uuid.New()
	var saved entity.ChatMessage
	chatDB := &mockChatDB{
		SaveMessageFn: func(ctx context.Context, msg entity.ChatMessage) (*entity.ChatMessage, error) {
			saved = msg
			msg.ID = uuid.New()
			return &msg, nil
		},
	}
	ai := &mockAIProvider{
		CompleteFn: func(ctx context.Context, systemPrompt string, history []entity.ChatMessage, userMessage string) (string, error) {
			assert.Empty(t, history)
			assert.Contains(t, userMessage, "French")
			return "Bonjour !", nil
		},
	}

	uc := newChatUC(chatDB, ai)
	result, err := uc.Greet(ctxWithClaims(userID), dto.GreetBody{Language: "French"})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, entity.ChatRoleAssistant, result.Role)
	assert.Equal(t, "Bonjour !", result.Content)
	assert.Equal(t, userID, saved.UserUUID)
}

func TestChatGreet_NoClaims(t *testing.T) {
	uc := newChatUC(&mockChatDB{}, &mockAIProvider{})
	result, err := uc.Greet(context.Background(), dto.GreetBody{Language: "en"})
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, usecase.ErrNoAuthClaims))
}

func TestChatGreet_AIError(t *testing.T) {
	ai := &mockAIProvider{
		CompleteFn: func(ctx context.Context, sp string, h []entity.ChatMessage, msg string) (string, error) {
			return "", errors.New("AI unavailable")
		},
	}
	uc := newChatUC(&mockChatDB{}, ai)
	result, err := uc.Greet(ctxWithClaims(uuid.New()), dto.GreetBody{Language: "en"})
	assert.Nil(t, result)
	require.Error(t, err)
}

func TestChatGreet_SaveError(t *testing.T) {
	chatDB := &mockChatDB{
		SaveMessageFn: func(ctx context.Context, msg entity.ChatMessage) (*entity.ChatMessage, error) {
			return nil, errors.New("db error")
		},
	}
	ai := &mockAIProvider{
		CompleteFn: func(ctx context.Context, sp string, h []entity.ChatMessage, msg string) (string, error) {
			return "hello", nil
		},
	}
	uc := newChatUC(chatDB, ai)
	result, err := uc.Greet(ctxWithClaims(uuid.New()), dto.GreetBody{Language: "en"})
	assert.Nil(t, result)
	require.Error(t, err)
}
