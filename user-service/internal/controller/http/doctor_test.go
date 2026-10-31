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

	"github.com/go-chi/chi"
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

type drUserDB struct {
	GetUserByIDFn func(ctx context.Context, id uuid.UUID) (*entity.User, error)
}

func (m *drUserDB) CreateUser(ctx context.Context, user entity.User) error { return nil }
func (m *drUserDB) UpdateUser(ctx context.Context, id uuid.UUID, b dto.UpdateProfileBody) error {
	return nil
}
func (m *drUserDB) GetUserByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	if m.GetUserByIDFn != nil {
		return m.GetUserByIDFn(ctx, id)
	}
	return nil, errors.New("not found")
}

type drToothDB struct {
	GetTeethByUserUUIDFn func(ctx context.Context, userUUID uuid.UUID) ([]entity.Tooth, error)
}

func (m *drToothDB) CreateInitialTeeth(ctx context.Context, userUUID uuid.UUID) ([]entity.Tooth, error) {
	return nil, nil
}
func (m *drToothDB) GetTeethByUserUUID(ctx context.Context, userUUID uuid.UUID) ([]entity.Tooth, error) {
	if m.GetTeethByUserUUIDFn != nil {
		return m.GetTeethByUserUUIDFn(ctx, userUUID)
	}
	return nil, nil
}
func (m *drToothDB) UpdateTooth(ctx context.Context, tooth entity.Tooth) error { return nil }

type drNoteDB struct {
	CreateToothNoteFn        func(ctx context.Context, note entity.ToothNote) (*entity.ToothNote, error)
	GetToothNotesByPatientFn func(ctx context.Context, doctorUUID, patientUUID uuid.UUID) ([]entity.ToothNote, error)
	UpdateToothNoteFn        func(ctx context.Context, noteID, doctorUUID uuid.UUID, newNote string) error
	DeleteToothNoteFn        func(ctx context.Context, noteID, doctorUUID uuid.UUID) error
}

func (m *drNoteDB) CreateToothNote(ctx context.Context, note entity.ToothNote) (*entity.ToothNote, error) {
	if m.CreateToothNoteFn != nil {
		return m.CreateToothNoteFn(ctx, note)
	}
	return &note, nil
}
func (m *drNoteDB) GetToothNotesByPatient(ctx context.Context, dID, pID uuid.UUID) ([]entity.ToothNote, error) {
	if m.GetToothNotesByPatientFn != nil {
		return m.GetToothNotesByPatientFn(ctx, dID, pID)
	}
	return nil, nil
}
func (m *drNoteDB) UpdateToothNote(ctx context.Context, noteID, dID uuid.UUID, newNote string) error {
	if m.UpdateToothNoteFn != nil {
		return m.UpdateToothNoteFn(ctx, noteID, dID, newNote)
	}
	return nil
}
func (m *drNoteDB) DeleteToothNote(ctx context.Context, noteID, dID uuid.UUID) error {
	if m.DeleteToothNoteFn != nil {
		return m.DeleteToothNoteFn(ctx, noteID, dID)
	}
	return nil
}

type drRecDB struct {
	CreateRecommendationFn        func(ctx context.Context, rec entity.Recommendation) (*entity.Recommendation, error)
	GetRecommendationsByPatientFn func(ctx context.Context, dID, pID uuid.UUID) ([]entity.Recommendation, error)
}

func (m *drRecDB) CreateRecommendation(ctx context.Context, rec entity.Recommendation) (*entity.Recommendation, error) {
	if m.CreateRecommendationFn != nil {
		return m.CreateRecommendationFn(ctx, rec)
	}
	return &rec, nil
}
func (m *drRecDB) GetRecommendationsByPatient(ctx context.Context, dID, pID uuid.UUID) ([]entity.Recommendation, error) {
	if m.GetRecommendationsByPatientFn != nil {
		return m.GetRecommendationsByPatientFn(ctx, dID, pID)
	}
	return nil, nil
}

// --- Helpers ---

func buildDoctorController(userDB *drUserDB, toothDB *drToothDB, noteDB *drNoteDB, recDB *drRecDB) *controllerhttp.DoctorController {
	uc := usecase.NewDoctorUseCase(userDB, toothDB, noteDB, recDB)
	return controllerhttp.NewDoctorController(uc, validator.New())
}

// withURLParams injects chi route URL params into the request context.
func withURLParams(r *http.Request, pairs ...string) *http.Request {
	rctx := chi.NewRouteContext()
	for i := 0; i+1 < len(pairs); i += 2 {
		rctx.URLParams.Add(pairs[i], pairs[i+1])
	}
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

// withDoctorClaims injects JWT access claims into the request context.
func withDoctorClaims(r *http.Request, doctorID uuid.UUID) *http.Request {
	claims := &sharedjwt.AccessClaims{UserID: doctorID, Role: 1}
	return r.WithContext(context.WithValue(r.Context(), sharedjwt.CtxKeyClaims{}, claims))
}

// combine applies both chi params and JWT claims to the same request.
func withParamsAndClaims(r *http.Request, doctorID uuid.UUID, pairs ...string) *http.Request {
	return withDoctorClaims(withURLParams(r, pairs...), doctorID)
}

func jsonBody(v any) *bytes.Buffer {
	b, _ := json.Marshal(v)
	return bytes.NewBuffer(b)
}

// --- Tests: GetPatientProfileHandler ---

func TestGetPatientProfileHandler_OK(t *testing.T) {
	patientID := uuid.New()
	userDB := &drUserDB{
		GetUserByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.User, error) {
			assert.Equal(t, patientID, id)
			return &entity.User{UUID: patientID, Phone: "+71234567890"}, nil
		},
	}
	ctrl := buildDoctorController(userDB, &drToothDB{}, &drNoteDB{}, &drRecDB{})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = withURLParams(req, "patient_uuid", patientID.String())
	w := httptest.NewRecorder()

	ctrl.GetPatientProfileHandler(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetPatientProfileHandler_InvalidUUID(t *testing.T) {
	ctrl := buildDoctorController(&drUserDB{}, &drToothDB{}, &drNoteDB{}, &drRecDB{})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = withURLParams(req, "patient_uuid", "not-a-uuid")
	w := httptest.NewRecorder()

	ctrl.GetPatientProfileHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetPatientProfileHandler_NotFound(t *testing.T) {
	userDB := &drUserDB{
		GetUserByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.User, error) {
			return nil, database.ErrUserNotFound
		},
	}
	ctrl := buildDoctorController(userDB, &drToothDB{}, &drNoteDB{}, &drRecDB{})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = withURLParams(req, "patient_uuid", uuid.New().String())
	w := httptest.NewRecorder()

	ctrl.GetPatientProfileHandler(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// --- Tests: GetPatientTeethHandler ---

func TestGetPatientTeethHandler_OK(t *testing.T) {
	patientID := uuid.New()
	teeth := []entity.Tooth{{UserUUID: patientID, ToothIndex: 0, ToothState: entity.ToothStateHealthy}}
	toothDB := &drToothDB{
		GetTeethByUserUUIDFn: func(ctx context.Context, id uuid.UUID) ([]entity.Tooth, error) {
			return teeth, nil
		},
	}
	ctrl := buildDoctorController(&drUserDB{}, toothDB, &drNoteDB{}, &drRecDB{})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = withURLParams(req, "patient_uuid", patientID.String())
	w := httptest.NewRecorder()

	ctrl.GetPatientTeethHandler(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetPatientTeethHandler_InvalidUUID(t *testing.T) {
	ctrl := buildDoctorController(&drUserDB{}, &drToothDB{}, &drNoteDB{}, &drRecDB{})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = withURLParams(req, "patient_uuid", "bad-uuid")
	w := httptest.NewRecorder()

	ctrl.GetPatientTeethHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// --- Tests: CreateToothNoteHandler ---

func TestCreateToothNoteHandler_OK(t *testing.T) {
	doctorID := uuid.New()
	patientID := uuid.New()
	noteDB := &drNoteDB{
		CreateToothNoteFn: func(ctx context.Context, note entity.ToothNote) (*entity.ToothNote, error) {
			note.ID = uuid.New()
			return &note, nil
		},
	}
	ctrl := buildDoctorController(&drUserDB{}, &drToothDB{}, noteDB, &drRecDB{})

	body := jsonBody(map[string]any{"tooth_index": 3, "note": "crack visible"})
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", "application/json")
	req = withParamsAndClaims(req, doctorID, "patient_uuid", patientID.String())
	w := httptest.NewRecorder()

	ctrl.CreateToothNoteHandler(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestCreateToothNoteHandler_InvalidUUID(t *testing.T) {
	ctrl := buildDoctorController(&drUserDB{}, &drToothDB{}, &drNoteDB{}, &drRecDB{})

	body := jsonBody(map[string]any{"tooth_index": 3, "note": "crack visible"})
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", "application/json")
	req = withURLParams(req, "patient_uuid", "bad-uuid")
	w := httptest.NewRecorder()

	ctrl.CreateToothNoteHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateToothNoteHandler_MissingNote(t *testing.T) {
	doctorID := uuid.New()
	patientID := uuid.New()
	ctrl := buildDoctorController(&drUserDB{}, &drToothDB{}, &drNoteDB{}, &drRecDB{})

	body := strings.NewReader(`{"tooth_index": 3}`)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", "application/json")
	req = withParamsAndClaims(req, doctorID, "patient_uuid", patientID.String())
	w := httptest.NewRecorder()

	ctrl.CreateToothNoteHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateToothNoteHandler_DBError(t *testing.T) {
	doctorID := uuid.New()
	patientID := uuid.New()
	noteDB := &drNoteDB{
		CreateToothNoteFn: func(ctx context.Context, note entity.ToothNote) (*entity.ToothNote, error) {
			return nil, errors.New("db error")
		},
	}
	ctrl := buildDoctorController(&drUserDB{}, &drToothDB{}, noteDB, &drRecDB{})

	body := jsonBody(map[string]any{"tooth_index": 0, "note": "note text"})
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", "application/json")
	req = withParamsAndClaims(req, doctorID, "patient_uuid", patientID.String())
	w := httptest.NewRecorder()

	ctrl.CreateToothNoteHandler(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- Tests: GetToothNotesHandler ---

func TestGetToothNotesHandler_OK(t *testing.T) {
	doctorID := uuid.New()
	patientID := uuid.New()
	noteDB := &drNoteDB{
		GetToothNotesByPatientFn: func(ctx context.Context, dID, pID uuid.UUID) ([]entity.ToothNote, error) {
			return []entity.ToothNote{{DoctorUUID: dID, PatientUUID: pID, Note: "ok"}}, nil
		},
	}
	ctrl := buildDoctorController(&drUserDB{}, &drToothDB{}, noteDB, &drRecDB{})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = withParamsAndClaims(req, doctorID, "patient_uuid", patientID.String())
	w := httptest.NewRecorder()

	ctrl.GetToothNotesHandler(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var result []map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&result))
	assert.Len(t, result, 1)
}

func TestGetToothNotesHandler_InvalidUUID(t *testing.T) {
	ctrl := buildDoctorController(&drUserDB{}, &drToothDB{}, &drNoteDB{}, &drRecDB{})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = withURLParams(req, "patient_uuid", "bad-uuid")
	w := httptest.NewRecorder()

	ctrl.GetToothNotesHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// --- Tests: UpdateToothNoteHandler ---

func TestUpdateToothNoteHandler_OK(t *testing.T) {
	doctorID := uuid.New()
	noteID := uuid.New()
	ctrl := buildDoctorController(&drUserDB{}, &drToothDB{}, &drNoteDB{}, &drRecDB{})

	body := jsonBody(map[string]any{"note": "revised note"})
	req := httptest.NewRequest(http.MethodPut, "/", body)
	req.Header.Set("Content-Type", "application/json")
	req = withParamsAndClaims(req, doctorID, "note_id", noteID.String())
	w := httptest.NewRecorder()

	ctrl.UpdateToothNoteHandler(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUpdateToothNoteHandler_InvalidUUID(t *testing.T) {
	ctrl := buildDoctorController(&drUserDB{}, &drToothDB{}, &drNoteDB{}, &drRecDB{})

	body := jsonBody(map[string]any{"note": "revised note"})
	req := httptest.NewRequest(http.MethodPut, "/", body)
	req.Header.Set("Content-Type", "application/json")
	req = withURLParams(req, "note_id", "bad-uuid")
	w := httptest.NewRecorder()

	ctrl.UpdateToothNoteHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateToothNoteHandler_MissingNote(t *testing.T) {
	doctorID := uuid.New()
	ctrl := buildDoctorController(&drUserDB{}, &drToothDB{}, &drNoteDB{}, &drRecDB{})

	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req = withParamsAndClaims(req, doctorID, "note_id", uuid.New().String())
	w := httptest.NewRecorder()

	ctrl.UpdateToothNoteHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateToothNoteHandler_NoteNotFound(t *testing.T) {
	doctorID := uuid.New()
	noteDB := &drNoteDB{
		UpdateToothNoteFn: func(ctx context.Context, noteID, dID uuid.UUID, newNote string) error {
			return database.ErrNoteNotFound
		},
	}
	ctrl := buildDoctorController(&drUserDB{}, &drToothDB{}, noteDB, &drRecDB{})

	body := jsonBody(map[string]any{"note": "revised"})
	req := httptest.NewRequest(http.MethodPut, "/", body)
	req.Header.Set("Content-Type", "application/json")
	req = withParamsAndClaims(req, doctorID, "note_id", uuid.New().String())
	w := httptest.NewRecorder()

	ctrl.UpdateToothNoteHandler(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// --- Tests: DeleteToothNoteHandler ---

func TestDeleteToothNoteHandler_OK(t *testing.T) {
	doctorID := uuid.New()
	ctrl := buildDoctorController(&drUserDB{}, &drToothDB{}, &drNoteDB{}, &drRecDB{})

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	req = withParamsAndClaims(req, doctorID, "note_id", uuid.New().String())
	w := httptest.NewRecorder()

	ctrl.DeleteToothNoteHandler(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeleteToothNoteHandler_InvalidUUID(t *testing.T) {
	ctrl := buildDoctorController(&drUserDB{}, &drToothDB{}, &drNoteDB{}, &drRecDB{})

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	req = withURLParams(req, "note_id", "bad-uuid")
	w := httptest.NewRecorder()

	ctrl.DeleteToothNoteHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteToothNoteHandler_NoteNotFound(t *testing.T) {
	doctorID := uuid.New()
	noteDB := &drNoteDB{
		DeleteToothNoteFn: func(ctx context.Context, noteID, dID uuid.UUID) error {
			return database.ErrNoteNotFound
		},
	}
	ctrl := buildDoctorController(&drUserDB{}, &drToothDB{}, noteDB, &drRecDB{})

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	req = withParamsAndClaims(req, doctorID, "note_id", uuid.New().String())
	w := httptest.NewRecorder()

	ctrl.DeleteToothNoteHandler(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// --- Tests: CreateRecommendationHandler ---

func TestCreateRecommendationHandler_OK(t *testing.T) {
	doctorID := uuid.New()
	patientID := uuid.New()
	recDB := &drRecDB{
		CreateRecommendationFn: func(ctx context.Context, rec entity.Recommendation) (*entity.Recommendation, error) {
			rec.ID = uuid.New()
			return &rec, nil
		},
	}
	ctrl := buildDoctorController(&drUserDB{}, &drToothDB{}, &drNoteDB{}, recDB)

	body := jsonBody(map[string]any{"text": "visit dentist in 3 months"})
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", "application/json")
	req = withParamsAndClaims(req, doctorID, "patient_uuid", patientID.String())
	w := httptest.NewRecorder()

	ctrl.CreateRecommendationHandler(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestCreateRecommendationHandler_InvalidUUID(t *testing.T) {
	ctrl := buildDoctorController(&drUserDB{}, &drToothDB{}, &drNoteDB{}, &drRecDB{})

	body := jsonBody(map[string]any{"text": "advice"})
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", "application/json")
	req = withURLParams(req, "patient_uuid", "bad-uuid")
	w := httptest.NewRecorder()

	ctrl.CreateRecommendationHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateRecommendationHandler_MissingText(t *testing.T) {
	doctorID := uuid.New()
	patientID := uuid.New()
	ctrl := buildDoctorController(&drUserDB{}, &drToothDB{}, &drNoteDB{}, &drRecDB{})

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req = withParamsAndClaims(req, doctorID, "patient_uuid", patientID.String())
	w := httptest.NewRecorder()

	ctrl.CreateRecommendationHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateRecommendationHandler_DBError(t *testing.T) {
	doctorID := uuid.New()
	patientID := uuid.New()
	recDB := &drRecDB{
		CreateRecommendationFn: func(ctx context.Context, rec entity.Recommendation) (*entity.Recommendation, error) {
			return nil, errors.New("db error")
		},
	}
	ctrl := buildDoctorController(&drUserDB{}, &drToothDB{}, &drNoteDB{}, recDB)

	body := jsonBody(map[string]any{"text": "advice"})
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", "application/json")
	req = withParamsAndClaims(req, doctorID, "patient_uuid", patientID.String())
	w := httptest.NewRecorder()

	ctrl.CreateRecommendationHandler(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- Tests: GetRecommendationsHandler ---

func TestGetRecommendationsHandler_OK(t *testing.T) {
	doctorID := uuid.New()
	patientID := uuid.New()
	recDB := &drRecDB{
		GetRecommendationsByPatientFn: func(ctx context.Context, dID, pID uuid.UUID) ([]entity.Recommendation, error) {
			return []entity.Recommendation{{DoctorUUID: dID, PatientUUID: pID, Text: "advice"}}, nil
		},
	}
	ctrl := buildDoctorController(&drUserDB{}, &drToothDB{}, &drNoteDB{}, recDB)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = withParamsAndClaims(req, doctorID, "patient_uuid", patientID.String())
	w := httptest.NewRecorder()

	ctrl.GetRecommendationsHandler(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var result []map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&result))
	assert.Len(t, result, 1)
}

func TestGetRecommendationsHandler_InvalidUUID(t *testing.T) {
	ctrl := buildDoctorController(&drUserDB{}, &drToothDB{}, &drNoteDB{}, &drRecDB{})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = withURLParams(req, "patient_uuid", "bad-uuid")
	w := httptest.NewRecorder()

	ctrl.GetRecommendationsHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
