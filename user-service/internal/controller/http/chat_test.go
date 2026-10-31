package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"shared/entity"
	sharedjwt "shared/jwt"
	controllerhttp "user-service/internal/controller/http"
	"user-service/internal/usecase"
)

// --- Mocks ---

type chChatDB struct {
	SaveMessageFn func(ctx context.Context, msg entity.ChatMessage) (*entity.ChatMessage, error)
	GetHistoryFn  func(ctx context.Context, userUUID uuid.UUID) ([]entity.ChatMessage, error)
}

func (m *chChatDB) SaveMessage(ctx context.Context, msg entity.ChatMessage) (*entity.ChatMessage, error) {
	if m.SaveMessageFn != nil {
		return m.SaveMessageFn(ctx, msg)
	}
	msg.ID = uuid.New()
	msg.CreatedAt = time.Now()
	return &msg, nil
}

func (m *chChatDB) GetHistory(ctx context.Context, userUUID uuid.UUID) ([]entity.ChatMessage, error) {
	if m.GetHistoryFn != nil {
		return m.GetHistoryFn(ctx, userUUID)
	}
	return nil, nil
}

type chAIProvider struct {
	CompleteFn func(ctx context.Context, systemPrompt string, history []entity.ChatMessage, userMessage string) (string, error)
}

func (m *chAIProvider) Complete(ctx context.Context, systemPrompt string, history []entity.ChatMessage, userMessage string) (string, error) {
	if m.CompleteFn != nil {
		return m.CompleteFn(ctx, systemPrompt, history, userMessage)
	}
	return "stub reply", nil
}

// --- Helpers ---

func buildChatController(chatDB *chChatDB, ai *chAIProvider) *controllerhttp.AiChatController {
	uc := usecase.NewAiChatUseCase(chatDB, ai)
	return controllerhttp.NewAiChatController(uc, validator.New())
}

func withUserClaims(r *http.Request, userID uuid.UUID) *http.Request {
	claims := &sharedjwt.AccessClaims{UserID: userID, Role: 0}
	return r.WithContext(context.WithValue(r.Context(), sharedjwt.CtxKeyClaims{}, claims))
}

// --- Tests: SendMessageHandler ---

func TestChatSendMessageHandler_OK(t *testing.T) {
	userID := uuid.New()
	ai := &chAIProvider{
		CompleteFn: func(ctx context.Context, sp string, h []entity.ChatMessage, msg string) (string, error) {
			return "floss every day", nil
		},
	}
	ctrl := buildChatController(&chChatDB{}, ai)

	b, _ := json.Marshal(map[string]any{"message": "should I floss?"})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	req = withUserClaims(req, userID)
	w := httptest.NewRecorder()

	ctrl.SendMessageHandler(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// the response is [saved user message, assistant reply]
	var body []map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&body))
	require.Len(t, body, 2)
	assert.Equal(t, "user", body[0]["Role"])
	assert.Equal(t, "should I floss?", body[0]["Content"])
	assert.Equal(t, "assistant", body[1]["Role"])
	assert.Equal(t, "floss every day", body[1]["Content"])
}

func TestChatSendMessageHandler_BadJSON(t *testing.T) {
	ctrl := buildChatController(&chChatDB{}, &chAIProvider{})

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`not json`))
	req.Header.Set("Content-Type", "application/json")
	req = withUserClaims(req, uuid.New())
	w := httptest.NewRecorder()

	ctrl.SendMessageHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestChatSendMessageHandler_EmptyMessage(t *testing.T) {
	ctrl := buildChatController(&chChatDB{}, &chAIProvider{})

	b, _ := json.Marshal(map[string]any{"message": ""})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	req = withUserClaims(req, uuid.New())
	w := httptest.NewRecorder()

	ctrl.SendMessageHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestChatSendMessageHandler_MissingMessage(t *testing.T) {
	ctrl := buildChatController(&chChatDB{}, &chAIProvider{})

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req = withUserClaims(req, uuid.New())
	w := httptest.NewRecorder()

	ctrl.SendMessageHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestChatSendMessageHandler_MessageTooLong(t *testing.T) {
	ctrl := buildChatController(&chChatDB{}, &chAIProvider{})

	b, _ := json.Marshal(map[string]any{"message": strings.Repeat("x", 2001)})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	req = withUserClaims(req, uuid.New())
	w := httptest.NewRecorder()

	ctrl.SendMessageHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestChatSendMessageHandler_AIError(t *testing.T) {
	ai := &chAIProvider{
		CompleteFn: func(ctx context.Context, sp string, h []entity.ChatMessage, msg string) (string, error) {
			return "", errors.New("provider unavailable")
		},
	}
	ctrl := buildChatController(&chChatDB{}, ai)

	b, _ := json.Marshal(map[string]any{"message": "hello"})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	req = withUserClaims(req, uuid.New())
	w := httptest.NewRecorder()

	ctrl.SendMessageHandler(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestChatSendMessageHandler_DBSaveError(t *testing.T) {
	chatDB := &chChatDB{
		SaveMessageFn: func(ctx context.Context, msg entity.ChatMessage) (*entity.ChatMessage, error) {
			return nil, errors.New("db error")
		},
	}
	ai := &chAIProvider{
		CompleteFn: func(ctx context.Context, sp string, h []entity.ChatMessage, msg string) (string, error) {
			return "reply", nil
		},
	}
	ctrl := buildChatController(chatDB, ai)

	b, _ := json.Marshal(map[string]any{"message": "hello"})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	req = withUserClaims(req, uuid.New())
	w := httptest.NewRecorder()

	ctrl.SendMessageHandler(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- Tests: GetHistoryHandler ---

func TestChatGetHistoryHandler_OK(t *testing.T) {
	userID := uuid.New()
	msgs := []entity.ChatMessage{
		{ID: uuid.New(), UserUUID: userID, Role: entity.ChatRoleUser, Content: "q", CreatedAt: time.Now()},
		{ID: uuid.New(), UserUUID: userID, Role: entity.ChatRoleAssistant, Content: "a", CreatedAt: time.Now()},
	}
	chatDB := &chChatDB{
		GetHistoryFn: func(ctx context.Context, id uuid.UUID) ([]entity.ChatMessage, error) {
			assert.Equal(t, userID, id)
			return msgs, nil
		},
	}
	ctrl := buildChatController(chatDB, &chAIProvider{})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = withUserClaims(req, userID)
	w := httptest.NewRecorder()

	ctrl.GetHistoryHandler(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var result []map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&result))
	assert.Len(t, result, 2)
	assert.Equal(t, "user", result[0]["Role"])
	assert.Equal(t, "assistant", result[1]["Role"])
}

func TestChatGetHistoryHandler_Empty(t *testing.T) {
	ctrl := buildChatController(&chChatDB{}, &chAIProvider{})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = withUserClaims(req, uuid.New())
	w := httptest.NewRecorder()

	ctrl.GetHistoryHandler(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestChatGetHistoryHandler_DBError(t *testing.T) {
	chatDB := &chChatDB{
		GetHistoryFn: func(ctx context.Context, id uuid.UUID) ([]entity.ChatMessage, error) {
			return nil, errors.New("db error")
		},
	}
	ctrl := buildChatController(chatDB, &chAIProvider{})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = withUserClaims(req, uuid.New())
	w := httptest.NewRecorder()

	ctrl.GetHistoryHandler(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- Tests: GreetHandler (request validation) ---

func TestChatGreetHandler_MissingLanguage(t *testing.T) {
	ctrl := buildChatController(&chChatDB{}, &chAIProvider{})

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req = withUserClaims(req, uuid.New())
	w := httptest.NewRecorder()

	ctrl.GreetHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestChatGreetHandler_LanguageTooShort(t *testing.T) {
	ctrl := buildChatController(&chChatDB{}, &chAIProvider{})

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"language":"e"}`))
	req.Header.Set("Content-Type", "application/json")
	req = withUserClaims(req, uuid.New())
	w := httptest.NewRecorder()

	ctrl.GreetHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
