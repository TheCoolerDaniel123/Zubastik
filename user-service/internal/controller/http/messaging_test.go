package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"shared/entity"
	sharedjwt "shared/jwt"
	"user-service/api/dto"
	controllerhttp "user-service/internal/controller/http"
	"user-service/internal/storage/database"
	"user-service/internal/usecase"
)

// --- Mocks ---

type msConvDB struct {
	GetOrCreateFn             func(ctx context.Context, d, p uuid.UUID) (*entity.Conversation, error)
	GetByIDFn                 func(ctx context.Context, id uuid.UUID) (*entity.Conversation, error)
	GetByDoctorFn             func(ctx context.Context, d uuid.UUID) ([]entity.Conversation, error)
	GetByPatientWithDetailsFn func(ctx context.Context, p uuid.UUID) ([]dto.ConversationResponse, error)
}

func (m *msConvDB) GetOrCreate(ctx context.Context, d, p uuid.UUID) (*entity.Conversation, error) {
	if m.GetOrCreateFn != nil {
		return m.GetOrCreateFn(ctx, d, p)
	}
	return &entity.Conversation{ID: uuid.New(), DoctorUUID: d, PatientUUID: p, CreatedAt: time.Now()}, nil
}
func (m *msConvDB) GetByID(ctx context.Context, id uuid.UUID) (*entity.Conversation, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}
	return nil, database.ErrConversationNotFound
}
func (m *msConvDB) GetByDoctor(ctx context.Context, d uuid.UUID) ([]entity.Conversation, error) {
	if m.GetByDoctorFn != nil {
		return m.GetByDoctorFn(ctx, d)
	}
	return nil, nil
}
func (m *msConvDB) GetByPatientWithDetails(ctx context.Context, p uuid.UUID) ([]dto.ConversationResponse, error) {
	if m.GetByPatientWithDetailsFn != nil {
		return m.GetByPatientWithDetailsFn(ctx, p)
	}
	return nil, nil
}

type msDMDB struct {
	CreateFn            func(ctx context.Context, msg entity.DirectMessage) (*entity.DirectMessage, error)
	GetByConversationFn func(ctx context.Context, convID uuid.UUID) ([]entity.DirectMessage, error)
}

func (m *msDMDB) Create(ctx context.Context, msg entity.DirectMessage) (*entity.DirectMessage, error) {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, msg)
	}
	msg.ID = uuid.New()
	msg.CreatedAt = time.Now()
	return &msg, nil
}
func (m *msDMDB) GetByConversation(ctx context.Context, id uuid.UUID) ([]entity.DirectMessage, error) {
	if m.GetByConversationFn != nil {
		return m.GetByConversationFn(ctx, id)
	}
	return nil, nil
}

// --- Helpers ---

func buildMessagingController(convDB *msConvDB, dmDB *msDMDB) *controllerhttp.MessagingController {
	uc := usecase.NewMessagingUseCase(convDB, dmDB)
	return controllerhttp.NewMessagingController(uc, validator.New())
}

func withClaims(r *http.Request, userID uuid.UUID) *http.Request {
	claims := &sharedjwt.AccessClaims{UserID: userID, Role: 0}
	return r.WithContext(context.WithValue(r.Context(), sharedjwt.CtxKeyClaims{}, claims))
}

// --- Tests: GetOrCreateConversationHandler ---

func TestGetOrCreateConversationHandler_OK(t *testing.T) {
	doctorID := uuid.New()
	patientID := uuid.New()
	ctrl := buildMessagingController(&msConvDB{}, &msDMDB{})

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req = withURLParams(withClaims(req, doctorID), "patient_uuid", patientID.String())
	w := httptest.NewRecorder()

	ctrl.GetOrCreateConversationHandler(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetOrCreateConversationHandler_InvalidUUID(t *testing.T) {
	ctrl := buildMessagingController(&msConvDB{}, &msDMDB{})

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req = withURLParams(req, "patient_uuid", "not-a-uuid")
	w := httptest.NewRecorder()

	ctrl.GetOrCreateConversationHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetOrCreateConversationHandler_DBError(t *testing.T) {
	convDB := &msConvDB{
		GetOrCreateFn: func(ctx context.Context, d, p uuid.UUID) (*entity.Conversation, error) {
			return nil, errors.New("db error")
		},
	}
	ctrl := buildMessagingController(convDB, &msDMDB{})

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req = withURLParams(withClaims(req, uuid.New()), "patient_uuid", uuid.New().String())
	w := httptest.NewRecorder()

	ctrl.GetOrCreateConversationHandler(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- Tests: GetDoctorConversationsHandler ---

func TestGetDoctorConversationsHandler_OK(t *testing.T) {
	doctorID := uuid.New()
	convDB := &msConvDB{
		GetByDoctorFn: func(ctx context.Context, d uuid.UUID) ([]entity.Conversation, error) {
			return []entity.Conversation{{ID: uuid.New(), DoctorUUID: d}}, nil
		},
	}
	ctrl := buildMessagingController(convDB, &msDMDB{})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = withClaims(req, doctorID)
	w := httptest.NewRecorder()

	ctrl.GetDoctorConversationsHandler(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var result []map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&result))
	assert.Len(t, result, 1)
}

func TestGetDoctorConversationsHandler_DBError(t *testing.T) {
	convDB := &msConvDB{
		GetByDoctorFn: func(ctx context.Context, d uuid.UUID) ([]entity.Conversation, error) {
			return nil, errors.New("db error")
		},
	}
	ctrl := buildMessagingController(convDB, &msDMDB{})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = withClaims(req, uuid.New())
	w := httptest.NewRecorder()

	ctrl.GetDoctorConversationsHandler(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- Tests: GetPatientConversationsHandler ---

func TestGetPatientConversationsHandler_OK(t *testing.T) {
	patientID := uuid.New()
	convDB := &msConvDB{
		GetByPatientWithDetailsFn: func(ctx context.Context, p uuid.UUID) ([]dto.ConversationResponse, error) {
			return []dto.ConversationResponse{{ID: uuid.New(), DoctorUUID: uuid.New(), DoctorName: "Dr. Smith"}}, nil
		},
	}
	ctrl := buildMessagingController(convDB, &msDMDB{})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = withClaims(req, patientID)
	w := httptest.NewRecorder()

	ctrl.GetPatientConversationsHandler(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

// --- Tests: SendMessageHandler ---

func TestSendDirectMessageHandler_OK(t *testing.T) {
	doctorID := uuid.New()
	patientID := uuid.New()
	convID := uuid.New()
	convDB := &msConvDB{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Conversation, error) {
			return &entity.Conversation{ID: convID, DoctorUUID: doctorID, PatientUUID: patientID}, nil
		},
	}
	ctrl := buildMessagingController(convDB, &msDMDB{})

	b, _ := json.Marshal(map[string]any{"content": "see you tomorrow"})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	req = withURLParams(withClaims(req, doctorID), "conversation_id", convID.String())
	w := httptest.NewRecorder()

	ctrl.SendMessageHandler(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestSendDirectMessageHandler_InvalidUUID(t *testing.T) {
	ctrl := buildMessagingController(&msConvDB{}, &msDMDB{})

	b, _ := json.Marshal(map[string]any{"content": "hi"})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	req = withURLParams(req, "conversation_id", "bad-uuid")
	w := httptest.NewRecorder()

	ctrl.SendMessageHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSendDirectMessageHandler_MissingContent(t *testing.T) {
	ctrl := buildMessagingController(&msConvDB{}, &msDMDB{})

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req = withURLParams(withClaims(req, uuid.New()), "conversation_id", uuid.New().String())
	w := httptest.NewRecorder()

	ctrl.SendMessageHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSendDirectMessageHandler_Forbidden(t *testing.T) {
	convID := uuid.New()
	convDB := &msConvDB{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Conversation, error) {
			return &entity.Conversation{ID: convID, DoctorUUID: uuid.New(), PatientUUID: uuid.New()}, nil
		},
	}
	ctrl := buildMessagingController(convDB, &msDMDB{})

	b, _ := json.Marshal(map[string]any{"content": "hi"})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	req = withURLParams(withClaims(req, uuid.New()), "conversation_id", convID.String())
	w := httptest.NewRecorder()

	ctrl.SendMessageHandler(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestSendDirectMessageHandler_ConversationNotFound(t *testing.T) {
	convDB := &msConvDB{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Conversation, error) {
			return nil, database.ErrConversationNotFound
		},
	}
	ctrl := buildMessagingController(convDB, &msDMDB{})

	b, _ := json.Marshal(map[string]any{"content": "hi"})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	req = withURLParams(withClaims(req, uuid.New()), "conversation_id", uuid.New().String())
	w := httptest.NewRecorder()

	ctrl.SendMessageHandler(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// --- Tests: GetMessagesHandler ---

func TestGetMessagesHandler_OK(t *testing.T) {
	doctorID := uuid.New()
	patientID := uuid.New()
	convID := uuid.New()
	convDB := &msConvDB{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Conversation, error) {
			return &entity.Conversation{ID: convID, DoctorUUID: doctorID, PatientUUID: patientID}, nil
		},
	}
	dmDB := &msDMDB{
		GetByConversationFn: func(ctx context.Context, id uuid.UUID) ([]entity.DirectMessage, error) {
			return []entity.DirectMessage{
				{ID: uuid.New(), ConversationID: convID, SenderUUID: doctorID, Content: "hi"},
			}, nil
		},
	}
	ctrl := buildMessagingController(convDB, dmDB)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = withURLParams(withClaims(req, patientID), "conversation_id", convID.String())
	w := httptest.NewRecorder()

	ctrl.GetMessagesHandler(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var result []map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&result))
	assert.Len(t, result, 1)
}

func TestGetMessagesHandler_InvalidUUID(t *testing.T) {
	ctrl := buildMessagingController(&msConvDB{}, &msDMDB{})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = withURLParams(req, "conversation_id", "bad-uuid")
	w := httptest.NewRecorder()

	ctrl.GetMessagesHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetMessagesHandler_Forbidden(t *testing.T) {
	convID := uuid.New()
	convDB := &msConvDB{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Conversation, error) {
			return &entity.Conversation{ID: convID, DoctorUUID: uuid.New(), PatientUUID: uuid.New()}, nil
		},
	}
	ctrl := buildMessagingController(convDB, &msDMDB{})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = withURLParams(withClaims(req, uuid.New()), "conversation_id", convID.String())
	w := httptest.NewRecorder()

	ctrl.GetMessagesHandler(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}
