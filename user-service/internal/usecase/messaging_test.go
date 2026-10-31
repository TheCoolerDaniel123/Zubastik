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
	"user-service/internal/storage/database"
	"user-service/internal/usecase"
)

// --- Mocks ---

type mockConvDB struct {
	GetOrCreateFn             func(ctx context.Context, doctorUUID, patientUUID uuid.UUID) (*entity.Conversation, error)
	GetByIDFn                 func(ctx context.Context, id uuid.UUID) (*entity.Conversation, error)
	GetByDoctorFn             func(ctx context.Context, doctorUUID uuid.UUID) ([]entity.Conversation, error)
	GetByPatientFn            func(ctx context.Context, patientUUID uuid.UUID) ([]entity.Conversation, error)
	GetByPatientWithDetailsFn func(ctx context.Context, patientUUID uuid.UUID) ([]dto.ConversationResponse, error)
}

// Проверка на этапе компиляции: мок реализует интерфейс.
var _ database.ConversationDatabase = (*mockConvDB)(nil)

func (m *mockConvDB) GetOrCreate(ctx context.Context, d, p uuid.UUID) (*entity.Conversation, error) {
	return m.GetOrCreateFn(ctx, d, p)
}
func (m *mockConvDB) GetByID(ctx context.Context, id uuid.UUID) (*entity.Conversation, error) {
	return m.GetByIDFn(ctx, id)
}
func (m *mockConvDB) GetByDoctor(ctx context.Context, d uuid.UUID) ([]entity.Conversation, error) {
	return m.GetByDoctorFn(ctx, d)
}
func (m *mockConvDB) GetByPatient(ctx context.Context, p uuid.UUID) ([]entity.Conversation, error) {
	return m.GetByPatientFn(ctx, p)
}
func (m *mockConvDB) GetByPatientWithDetails(ctx context.Context, p uuid.UUID) ([]dto.ConversationResponse, error) {
	if m.GetByPatientWithDetailsFn != nil {
		return m.GetByPatientWithDetailsFn(ctx, p)
	}
	return nil, nil
}

type mockDMDB struct {
	CreateFn            func(ctx context.Context, msg entity.DirectMessage) (*entity.DirectMessage, error)
	GetByConversationFn func(ctx context.Context, convID uuid.UUID) ([]entity.DirectMessage, error)
}

func (m *mockDMDB) Create(ctx context.Context, msg entity.DirectMessage) (*entity.DirectMessage, error) {
	return m.CreateFn(ctx, msg)
}
func (m *mockDMDB) GetByConversation(ctx context.Context, id uuid.UUID) ([]entity.DirectMessage, error) {
	return m.GetByConversationFn(ctx, id)
}

func newMessagingUC(convDB *mockConvDB, dmDB *mockDMDB) *usecase.MessagingUseCase {
	return usecase.NewMessagingUseCase(convDB, dmDB)
}

// --- Tests: GetOrCreateConversation ---

func TestMessaging_GetOrCreateConversation_Success(t *testing.T) {
	doctorID := uuid.New()
	patientID := uuid.New()
	expected := &entity.Conversation{ID: uuid.New(), DoctorUUID: doctorID, PatientUUID: patientID}
	convDB := &mockConvDB{
		GetOrCreateFn: func(ctx context.Context, d, p uuid.UUID) (*entity.Conversation, error) {
			assert.Equal(t, doctorID, d)
			assert.Equal(t, patientID, p)
			return expected, nil
		},
	}
	uc := newMessagingUC(convDB, &mockDMDB{})
	result, err := uc.GetOrCreateConversation(ctxWithClaims(doctorID), patientID)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestMessaging_GetOrCreateConversation_NoClaims(t *testing.T) {
	uc := newMessagingUC(&mockConvDB{}, &mockDMDB{})
	result, err := uc.GetOrCreateConversation(context.Background(), uuid.New())
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, usecase.ErrNoAuthClaims))
}

func TestMessaging_GetOrCreateConversation_DBError(t *testing.T) {
	convDB := &mockConvDB{
		GetOrCreateFn: func(ctx context.Context, d, p uuid.UUID) (*entity.Conversation, error) {
			return nil, errors.New("db error")
		},
	}
	uc := newMessagingUC(convDB, &mockDMDB{})
	result, err := uc.GetOrCreateConversation(ctxWithClaims(uuid.New()), uuid.New())
	assert.Nil(t, result)
	require.Error(t, err)
}

// --- Tests: GetDoctorConversations ---

func TestMessaging_GetDoctorConversations_Success(t *testing.T) {
	doctorID := uuid.New()
	expected := []entity.Conversation{{ID: uuid.New(), DoctorUUID: doctorID}}
	convDB := &mockConvDB{
		GetByDoctorFn: func(ctx context.Context, d uuid.UUID) ([]entity.Conversation, error) {
			assert.Equal(t, doctorID, d)
			return expected, nil
		},
	}
	uc := newMessagingUC(convDB, &mockDMDB{})
	result, err := uc.GetDoctorConversations(ctxWithClaims(doctorID))
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestMessaging_GetDoctorConversations_NoClaims(t *testing.T) {
	uc := newMessagingUC(&mockConvDB{}, &mockDMDB{})
	result, err := uc.GetDoctorConversations(context.Background())
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, usecase.ErrNoAuthClaims))
}

// --- Tests: GetPatientConversations ---

func TestMessaging_GetPatientConversations_Success(t *testing.T) {
	patientID := uuid.New()
	// Поля dto.ConversationResponse не заполнены намеренно: добавьте нужные при необходимости.
	expected := []dto.ConversationResponse{{}}
	convDB := &mockConvDB{
		GetByPatientWithDetailsFn: func(ctx context.Context, p uuid.UUID) ([]dto.ConversationResponse, error) {
			assert.Equal(t, patientID, p)
			return expected, nil
		},
	}
	uc := newMessagingUC(convDB, &mockDMDB{})
	result, err := uc.GetPatientConversations(ctxWithClaims(patientID))
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestMessaging_GetPatientConversations_NoClaims(t *testing.T) {
	uc := newMessagingUC(&mockConvDB{}, &mockDMDB{})
	result, err := uc.GetPatientConversations(context.Background())
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, usecase.ErrNoAuthClaims))
}

// --- Tests: SendMessage ---

func TestMessaging_SendMessage_ByDoctor(t *testing.T) {
	doctorID := uuid.New()
	patientID := uuid.New()
	convID := uuid.New()
	conv := &entity.Conversation{ID: convID, DoctorUUID: doctorID, PatientUUID: patientID}

	convDB := &mockConvDB{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Conversation, error) {
			return conv, nil
		},
	}
	dmDB := &mockDMDB{
		CreateFn: func(ctx context.Context, msg entity.DirectMessage) (*entity.DirectMessage, error) {
			assert.Equal(t, convID, msg.ConversationID)
			assert.Equal(t, doctorID, msg.SenderUUID)
			assert.Equal(t, "hello patient", msg.Content)
			msg.ID = uuid.New()
			msg.CreatedAt = time.Now()
			return &msg, nil
		},
	}
	uc := newMessagingUC(convDB, dmDB)
	result, err := uc.SendMessage(ctxWithClaims(doctorID), convID, dto.SendDirectMessageBody{Content: "hello patient"})
	require.NoError(t, err)
	assert.Equal(t, doctorID, result.SenderUUID)
}

func TestMessaging_SendMessage_ByPatient(t *testing.T) {
	doctorID := uuid.New()
	patientID := uuid.New()
	convID := uuid.New()
	conv := &entity.Conversation{ID: convID, DoctorUUID: doctorID, PatientUUID: patientID}

	convDB := &mockConvDB{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Conversation, error) { return conv, nil },
	}
	dmDB := &mockDMDB{
		CreateFn: func(ctx context.Context, msg entity.DirectMessage) (*entity.DirectMessage, error) {
			assert.Equal(t, patientID, msg.SenderUUID)
			msg.ID = uuid.New()
			return &msg, nil
		},
	}
	uc := newMessagingUC(convDB, dmDB)
	result, err := uc.SendMessage(ctxWithClaims(patientID), convID, dto.SendDirectMessageBody{Content: "reply"})
	require.NoError(t, err)
	assert.Equal(t, patientID, result.SenderUUID)
}

func TestMessaging_SendMessage_Forbidden(t *testing.T) {
	convID := uuid.New()
	conv := &entity.Conversation{ID: convID, DoctorUUID: uuid.New(), PatientUUID: uuid.New()}
	convDB := &mockConvDB{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Conversation, error) { return conv, nil },
	}
	uc := newMessagingUC(convDB, &mockDMDB{})
	result, err := uc.SendMessage(ctxWithClaims(uuid.New()), convID, dto.SendDirectMessageBody{Content: "x"})
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, usecase.ErrForbidden))
}

func TestMessaging_SendMessage_NoClaims(t *testing.T) {
	uc := newMessagingUC(&mockConvDB{}, &mockDMDB{})
	result, err := uc.SendMessage(context.Background(), uuid.New(), dto.SendDirectMessageBody{Content: "x"})
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, usecase.ErrNoAuthClaims))
}

func TestMessaging_SendMessage_ConversationNotFound(t *testing.T) {
	convDB := &mockConvDB{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Conversation, error) {
			return nil, database.ErrConversationNotFound
		},
	}
	uc := newMessagingUC(convDB, &mockDMDB{})
	result, err := uc.SendMessage(ctxWithClaims(uuid.New()), uuid.New(), dto.SendDirectMessageBody{Content: "x"})
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, database.ErrConversationNotFound))
}

// --- Tests: GetMessages ---

func TestMessaging_GetMessages_Success(t *testing.T) {
	doctorID := uuid.New()
	patientID := uuid.New()
	convID := uuid.New()
	conv := &entity.Conversation{ID: convID, DoctorUUID: doctorID, PatientUUID: patientID}
	expected := []entity.DirectMessage{
		{ID: uuid.New(), ConversationID: convID, SenderUUID: doctorID, Content: "hi"},
		{ID: uuid.New(), ConversationID: convID, SenderUUID: patientID, Content: "hello"},
	}

	convDB := &mockConvDB{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Conversation, error) { return conv, nil },
	}
	dmDB := &mockDMDB{
		GetByConversationFn: func(ctx context.Context, id uuid.UUID) ([]entity.DirectMessage, error) {
			assert.Equal(t, convID, id)
			return expected, nil
		},
	}
	uc := newMessagingUC(convDB, dmDB)
	result, err := uc.GetMessages(ctxWithClaims(doctorID), convID)
	require.NoError(t, err)

	// GetMessages returns dto.MessageResponse (with the IsMine flag), not entity.DirectMessage.
	require.Len(t, result, len(expected))
	for i, m := range expected {
		assert.Equal(t, m.ID, result[i].ID)
		assert.Equal(t, m.ConversationID, result[i].ConversationID)
		assert.Equal(t, m.SenderUUID, result[i].SenderUUID)
		assert.Equal(t, m.Content, result[i].Content)
		assert.Equal(t, m.CreatedAt, result[i].CreatedAt)
		assert.Equal(t, m.SenderUUID == doctorID, result[i].IsMine)
	}
}

func TestMessaging_GetMessages_IsMineDependsOnCaller(t *testing.T) {
	doctorID := uuid.New()
	patientID := uuid.New()
	convID := uuid.New()
	conv := &entity.Conversation{ID: convID, DoctorUUID: doctorID, PatientUUID: patientID}
	msgs := []entity.DirectMessage{
		{ID: uuid.New(), ConversationID: convID, SenderUUID: doctorID, Content: "from doctor"},
		{ID: uuid.New(), ConversationID: convID, SenderUUID: patientID, Content: "from patient"},
	}

	convDB := &mockConvDB{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Conversation, error) { return conv, nil },
	}
	dmDB := &mockDMDB{
		GetByConversationFn: func(ctx context.Context, id uuid.UUID) ([]entity.DirectMessage, error) { return msgs, nil },
	}
	uc := newMessagingUC(convDB, dmDB)

	asDoctor, err := uc.GetMessages(ctxWithClaims(doctorID), convID)
	require.NoError(t, err)
	require.Len(t, asDoctor, 2)
	assert.True(t, asDoctor[0].IsMine)
	assert.False(t, asDoctor[1].IsMine)

	asPatient, err := uc.GetMessages(ctxWithClaims(patientID), convID)
	require.NoError(t, err)
	require.Len(t, asPatient, 2)
	assert.False(t, asPatient[0].IsMine)
	assert.True(t, asPatient[1].IsMine)
}

func TestMessaging_GetMessages_Forbidden(t *testing.T) {
	convID := uuid.New()
	conv := &entity.Conversation{ID: convID, DoctorUUID: uuid.New(), PatientUUID: uuid.New()}
	convDB := &mockConvDB{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Conversation, error) { return conv, nil },
	}
	uc := newMessagingUC(convDB, &mockDMDB{})
	result, err := uc.GetMessages(ctxWithClaims(uuid.New()), convID)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, usecase.ErrForbidden))
}

func TestMessaging_GetMessages_NoClaims(t *testing.T) {
	uc := newMessagingUC(&mockConvDB{}, &mockDMDB{})
	result, err := uc.GetMessages(context.Background(), uuid.New())
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, usecase.ErrNoAuthClaims))
}

func TestMessaging_GetMessages_ConversationNotFound(t *testing.T) {
	convDB := &mockConvDB{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Conversation, error) {
			return nil, database.ErrConversationNotFound
		},
	}
	uc := newMessagingUC(convDB, &mockDMDB{})
	result, err := uc.GetMessages(ctxWithClaims(uuid.New()), uuid.New())
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, database.ErrConversationNotFound))
}
