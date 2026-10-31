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
	"user-service/internal/storage/database"
	"user-service/internal/usecase"
)

// --- Mocks (mockUserDB, mockToothDB, ctxWithClaims reused from user_test.go) ---

type mockToothNoteDB struct {
	CreateToothNoteFn        func(ctx context.Context, note entity.ToothNote) (*entity.ToothNote, error)
	GetToothNotesByPatientFn func(ctx context.Context, doctorUUID, patientUUID uuid.UUID) ([]entity.ToothNote, error)
	UpdateToothNoteFn        func(ctx context.Context, noteID, doctorUUID uuid.UUID, newNote string) error
	DeleteToothNoteFn        func(ctx context.Context, noteID, doctorUUID uuid.UUID) error
}

func (m *mockToothNoteDB) CreateToothNote(ctx context.Context, note entity.ToothNote) (*entity.ToothNote, error) {
	return m.CreateToothNoteFn(ctx, note)
}

func (m *mockToothNoteDB) GetToothNotesByPatient(ctx context.Context, doctorUUID, patientUUID uuid.UUID) ([]entity.ToothNote, error) {
	return m.GetToothNotesByPatientFn(ctx, doctorUUID, patientUUID)
}

func (m *mockToothNoteDB) UpdateToothNote(ctx context.Context, noteID, doctorUUID uuid.UUID, newNote string) error {
	return m.UpdateToothNoteFn(ctx, noteID, doctorUUID, newNote)
}

func (m *mockToothNoteDB) DeleteToothNote(ctx context.Context, noteID, doctorUUID uuid.UUID) error {
	return m.DeleteToothNoteFn(ctx, noteID, doctorUUID)
}

type mockRecommendationDB struct {
	CreateRecommendationFn        func(ctx context.Context, rec entity.Recommendation) (*entity.Recommendation, error)
	GetRecommendationsByPatientFn func(ctx context.Context, doctorUUID, patientUUID uuid.UUID) ([]entity.Recommendation, error)
}

func (m *mockRecommendationDB) CreateRecommendation(ctx context.Context, rec entity.Recommendation) (*entity.Recommendation, error) {
	return m.CreateRecommendationFn(ctx, rec)
}

func (m *mockRecommendationDB) GetRecommendationsByPatient(ctx context.Context, doctorUUID, patientUUID uuid.UUID) ([]entity.Recommendation, error) {
	return m.GetRecommendationsByPatientFn(ctx, doctorUUID, patientUUID)
}

// --- Helper ---

func newDoctorUC(userDB *mockUserDB, toothDB *mockToothDB, noteDB *mockToothNoteDB, recDB *mockRecommendationDB) *usecase.DoctorUseCase {
	return usecase.NewDoctorUseCase(userDB, toothDB, noteDB, recDB)
}

// --- Tests: GetPatientByUUID ---

func TestDoctorGetPatientByUUID_Success(t *testing.T) {
	patientID := uuid.New()
	expected := &entity.User{UUID: patientID, Phone: "+71234567890"}
	db := &mockUserDB{
		GetUserByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.User, error) {
			assert.Equal(t, patientID, id)
			return expected, nil
		},
	}
	uc := newDoctorUC(db, &mockToothDB{}, &mockToothNoteDB{}, &mockRecommendationDB{})
	result, err := uc.GetPatientByUUID(context.Background(), patientID)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestDoctorGetPatientByUUID_DBError(t *testing.T) {
	db := &mockUserDB{
		GetUserByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.User, error) {
			return nil, errors.New("db error")
		},
	}
	uc := newDoctorUC(db, &mockToothDB{}, &mockToothNoteDB{}, &mockRecommendationDB{})
	result, err := uc.GetPatientByUUID(context.Background(), uuid.New())
	assert.Nil(t, result)
	require.Error(t, err)
}

// --- Tests: GetPatientTeeth ---

func TestDoctorGetPatientTeeth_Success(t *testing.T) {
	patientID := uuid.New()
	expected := []entity.Tooth{{UserUUID: patientID, ToothIndex: 0, ToothState: entity.ToothStateHealthy}}
	toothDB := &mockToothDB{
		GetTeethByUserUUIDFn: func(ctx context.Context, id uuid.UUID) ([]entity.Tooth, error) {
			assert.Equal(t, patientID, id)
			return expected, nil
		},
	}
	uc := newDoctorUC(&mockUserDB{}, toothDB, &mockToothNoteDB{}, &mockRecommendationDB{})
	result, err := uc.GetPatientTeeth(context.Background(), patientID)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestDoctorGetPatientTeeth_DBError(t *testing.T) {
	toothDB := &mockToothDB{
		GetTeethByUserUUIDFn: func(ctx context.Context, id uuid.UUID) ([]entity.Tooth, error) {
			return nil, errors.New("db error")
		},
	}
	uc := newDoctorUC(&mockUserDB{}, toothDB, &mockToothNoteDB{}, &mockRecommendationDB{})
	result, err := uc.GetPatientTeeth(context.Background(), uuid.New())
	assert.Nil(t, result)
	require.Error(t, err)
}

// --- Tests: CreateToothNote ---

func TestDoctorCreateToothNote_Success(t *testing.T) {
	doctorID := uuid.New()
	patientID := uuid.New()
	noteDB := &mockToothNoteDB{
		CreateToothNoteFn: func(ctx context.Context, note entity.ToothNote) (*entity.ToothNote, error) {
			assert.Equal(t, doctorID, note.DoctorUUID)
			assert.Equal(t, patientID, note.PatientUUID)
			assert.Equal(t, int16(5), note.ToothIndex)
			assert.Equal(t, "cavity detected", note.Note)
			note.ID = uuid.New()
			note.CreatedAt = time.Now()
			return &note, nil
		},
	}
	uc := newDoctorUC(&mockUserDB{}, &mockToothDB{}, noteDB, &mockRecommendationDB{})

	result, err := uc.CreateToothNote(ctxWithClaims(doctorID), entity.ToothNote{
		PatientUUID: patientID,
		ToothIndex:  5,
		Note:        "cavity detected",
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, doctorID, result.DoctorUUID)
	assert.Equal(t, patientID, result.PatientUUID)
}

func TestDoctorCreateToothNote_NoClaims(t *testing.T) {
	uc := newDoctorUC(&mockUserDB{}, &mockToothDB{}, &mockToothNoteDB{}, &mockRecommendationDB{})
	result, err := uc.CreateToothNote(context.Background(), entity.ToothNote{})
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, usecase.ErrNoAuthClaims))
}

func TestDoctorCreateToothNote_DBError(t *testing.T) {
	noteDB := &mockToothNoteDB{
		CreateToothNoteFn: func(ctx context.Context, note entity.ToothNote) (*entity.ToothNote, error) {
			return nil, errors.New("db error")
		},
	}
	uc := newDoctorUC(&mockUserDB{}, &mockToothDB{}, noteDB, &mockRecommendationDB{})
	result, err := uc.CreateToothNote(ctxWithClaims(uuid.New()), entity.ToothNote{})
	assert.Nil(t, result)
	require.Error(t, err)
}

// --- Tests: GetToothNotes ---

func TestDoctorGetToothNotes_Success(t *testing.T) {
	doctorID := uuid.New()
	patientID := uuid.New()
	expected := []entity.ToothNote{{DoctorUUID: doctorID, PatientUUID: patientID, Note: "cavity"}}
	noteDB := &mockToothNoteDB{
		GetToothNotesByPatientFn: func(ctx context.Context, dID, pID uuid.UUID) ([]entity.ToothNote, error) {
			assert.Equal(t, doctorID, dID)
			assert.Equal(t, patientID, pID)
			return expected, nil
		},
	}
	uc := newDoctorUC(&mockUserDB{}, &mockToothDB{}, noteDB, &mockRecommendationDB{})
	result, err := uc.GetToothNotes(ctxWithClaims(doctorID), patientID)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestDoctorGetToothNotes_NoClaims(t *testing.T) {
	uc := newDoctorUC(&mockUserDB{}, &mockToothDB{}, &mockToothNoteDB{}, &mockRecommendationDB{})
	result, err := uc.GetToothNotes(context.Background(), uuid.New())
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, usecase.ErrNoAuthClaims))
}

func TestDoctorGetToothNotes_DBError(t *testing.T) {
	noteDB := &mockToothNoteDB{
		GetToothNotesByPatientFn: func(ctx context.Context, dID, pID uuid.UUID) ([]entity.ToothNote, error) {
			return nil, errors.New("db error")
		},
	}
	uc := newDoctorUC(&mockUserDB{}, &mockToothDB{}, noteDB, &mockRecommendationDB{})
	result, err := uc.GetToothNotes(ctxWithClaims(uuid.New()), uuid.New())
	assert.Nil(t, result)
	require.Error(t, err)
}

// --- Tests: UpdateToothNote ---

func TestDoctorUpdateToothNote_Success(t *testing.T) {
	doctorID := uuid.New()
	noteID := uuid.New()
	noteDB := &mockToothNoteDB{
		UpdateToothNoteFn: func(ctx context.Context, nID, dID uuid.UUID, newNote string) error {
			assert.Equal(t, noteID, nID)
			assert.Equal(t, doctorID, dID)
			assert.Equal(t, "updated text", newNote)
			return nil
		},
	}
	uc := newDoctorUC(&mockUserDB{}, &mockToothDB{}, noteDB, &mockRecommendationDB{})
	err := uc.UpdateToothNote(ctxWithClaims(doctorID), noteID, "updated text")
	require.NoError(t, err)
}

func TestDoctorUpdateToothNote_NoClaims(t *testing.T) {
	uc := newDoctorUC(&mockUserDB{}, &mockToothDB{}, &mockToothNoteDB{}, &mockRecommendationDB{})
	err := uc.UpdateToothNote(context.Background(), uuid.New(), "text")
	assert.True(t, errors.Is(err, usecase.ErrNoAuthClaims))
}

func TestDoctorUpdateToothNote_NoteNotFound(t *testing.T) {
	noteDB := &mockToothNoteDB{
		UpdateToothNoteFn: func(ctx context.Context, nID, dID uuid.UUID, newNote string) error {
			return database.ErrNoteNotFound
		},
	}
	uc := newDoctorUC(&mockUserDB{}, &mockToothDB{}, noteDB, &mockRecommendationDB{})
	err := uc.UpdateToothNote(ctxWithClaims(uuid.New()), uuid.New(), "text")
	assert.True(t, errors.Is(err, database.ErrNoteNotFound))
}

// --- Tests: DeleteToothNote ---

func TestDoctorDeleteToothNote_Success(t *testing.T) {
	doctorID := uuid.New()
	noteID := uuid.New()
	noteDB := &mockToothNoteDB{
		DeleteToothNoteFn: func(ctx context.Context, nID, dID uuid.UUID) error {
			assert.Equal(t, noteID, nID)
			assert.Equal(t, doctorID, dID)
			return nil
		},
	}
	uc := newDoctorUC(&mockUserDB{}, &mockToothDB{}, noteDB, &mockRecommendationDB{})
	err := uc.DeleteToothNote(ctxWithClaims(doctorID), noteID)
	require.NoError(t, err)
}

func TestDoctorDeleteToothNote_NoClaims(t *testing.T) {
	uc := newDoctorUC(&mockUserDB{}, &mockToothDB{}, &mockToothNoteDB{}, &mockRecommendationDB{})
	err := uc.DeleteToothNote(context.Background(), uuid.New())
	assert.True(t, errors.Is(err, usecase.ErrNoAuthClaims))
}

func TestDoctorDeleteToothNote_NoteNotFound(t *testing.T) {
	noteDB := &mockToothNoteDB{
		DeleteToothNoteFn: func(ctx context.Context, nID, dID uuid.UUID) error {
			return database.ErrNoteNotFound
		},
	}
	uc := newDoctorUC(&mockUserDB{}, &mockToothDB{}, noteDB, &mockRecommendationDB{})
	err := uc.DeleteToothNote(ctxWithClaims(uuid.New()), uuid.New())
	assert.True(t, errors.Is(err, database.ErrNoteNotFound))
}

// --- Tests: CreateRecommendation ---

func TestDoctorCreateRecommendation_Success(t *testing.T) {
	doctorID := uuid.New()
	patientID := uuid.New()
	recDB := &mockRecommendationDB{
		CreateRecommendationFn: func(ctx context.Context, rec entity.Recommendation) (*entity.Recommendation, error) {
			assert.Equal(t, doctorID, rec.DoctorUUID)
			assert.Equal(t, patientID, rec.PatientUUID)
			assert.Equal(t, "brush twice daily", rec.Text)
			rec.ID = uuid.New()
			rec.CreatedAt = time.Now()
			return &rec, nil
		},
	}
	uc := newDoctorUC(&mockUserDB{}, &mockToothDB{}, &mockToothNoteDB{}, recDB)

	result, err := uc.CreateRecommendation(ctxWithClaims(doctorID), entity.Recommendation{
		PatientUUID: patientID,
		Text:        "brush twice daily",
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, doctorID, result.DoctorUUID)
	assert.Equal(t, patientID, result.PatientUUID)
}

func TestDoctorCreateRecommendation_NoClaims(t *testing.T) {
	uc := newDoctorUC(&mockUserDB{}, &mockToothDB{}, &mockToothNoteDB{}, &mockRecommendationDB{})
	result, err := uc.CreateRecommendation(context.Background(), entity.Recommendation{})
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, usecase.ErrNoAuthClaims))
}

func TestDoctorCreateRecommendation_DBError(t *testing.T) {
	recDB := &mockRecommendationDB{
		CreateRecommendationFn: func(ctx context.Context, rec entity.Recommendation) (*entity.Recommendation, error) {
			return nil, errors.New("db error")
		},
	}
	uc := newDoctorUC(&mockUserDB{}, &mockToothDB{}, &mockToothNoteDB{}, recDB)
	result, err := uc.CreateRecommendation(ctxWithClaims(uuid.New()), entity.Recommendation{})
	assert.Nil(t, result)
	require.Error(t, err)
}

// --- Tests: GetRecommendations ---

func TestDoctorGetRecommendations_Success(t *testing.T) {
	doctorID := uuid.New()
	patientID := uuid.New()
	expected := []entity.Recommendation{{DoctorUUID: doctorID, PatientUUID: patientID, Text: "brush twice daily"}}
	recDB := &mockRecommendationDB{
		GetRecommendationsByPatientFn: func(ctx context.Context, dID, pID uuid.UUID) ([]entity.Recommendation, error) {
			assert.Equal(t, doctorID, dID)
			assert.Equal(t, patientID, pID)
			return expected, nil
		},
	}
	uc := newDoctorUC(&mockUserDB{}, &mockToothDB{}, &mockToothNoteDB{}, recDB)
	result, err := uc.GetRecommendations(ctxWithClaims(doctorID), patientID)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestDoctorGetRecommendations_NoClaims(t *testing.T) {
	uc := newDoctorUC(&mockUserDB{}, &mockToothDB{}, &mockToothNoteDB{}, &mockRecommendationDB{})
	result, err := uc.GetRecommendations(context.Background(), uuid.New())
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, usecase.ErrNoAuthClaims))
}

func TestDoctorGetRecommendations_DBError(t *testing.T) {
	recDB := &mockRecommendationDB{
		GetRecommendationsByPatientFn: func(ctx context.Context, dID, pID uuid.UUID) ([]entity.Recommendation, error) {
			return nil, errors.New("db error")
		},
	}
	uc := newDoctorUC(&mockUserDB{}, &mockToothDB{}, &mockToothNoteDB{}, recDB)
	result, err := uc.GetRecommendations(ctxWithClaims(uuid.New()), uuid.New())
	assert.Nil(t, result)
	require.Error(t, err)
}
