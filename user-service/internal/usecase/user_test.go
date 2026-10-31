package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"shared/entity"
	sharedjwt "shared/jwt"
	"user-service/api/dto"
	"user-service/internal/storage/database"
	"user-service/internal/usecase"
)

// --- Mocks ---

type mockUserDB struct {
	CreateUserFn  func(ctx context.Context, user entity.User) error
	GetUserByIDFn func(ctx context.Context, id uuid.UUID) (*entity.User, error)
	UpdateUserFn  func(ctx context.Context, id uuid.UUID, body dto.UpdateProfileBody) error
}

func (m *mockUserDB) CreateUser(ctx context.Context, user entity.User) error {
	return m.CreateUserFn(ctx, user)
}

func (m *mockUserDB) GetUserByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	return m.GetUserByIDFn(ctx, id)
}

func (m *mockUserDB) UpdateUser(ctx context.Context, id uuid.UUID, body dto.UpdateProfileBody) error {
	return m.UpdateUserFn(ctx, id, body)
}

type mockToothDB struct {
	CreateInitialTeethFn func(ctx context.Context, userUUID uuid.UUID) ([]entity.Tooth, error)
	GetTeethByUserUUIDFn func(ctx context.Context, userUUID uuid.UUID) ([]entity.Tooth, error)
	UpdateToothFn        func(ctx context.Context, tooth entity.Tooth) error
}

func (m *mockToothDB) CreateInitialTeeth(ctx context.Context, userUUID uuid.UUID) ([]entity.Tooth, error) {
	return m.CreateInitialTeethFn(ctx, userUUID)
}

func (m *mockToothDB) GetTeethByUserUUID(ctx context.Context, userUUID uuid.UUID) ([]entity.Tooth, error) {
	if m.GetTeethByUserUUIDFn != nil {
		return m.GetTeethByUserUUIDFn(ctx, userUUID)
	}
	return nil, nil
}

func (m *mockToothDB) UpdateTooth(ctx context.Context, tooth entity.Tooth) error {
	if m.UpdateToothFn != nil {
		return m.UpdateToothFn(ctx, tooth)
	}
	return nil
}

type mockBrushDB struct {
	GetBrushingTimesFn func(ctx context.Context, userUUID uuid.UUID) ([]entity.BrushTime, error)
	ReplaceAllFn       func(ctx context.Context, userUUID uuid.UUID, times []entity.BrushTime) error
}

func (m *mockBrushDB) CreateBrushingTime(ctx context.Context, bt entity.BrushTime) error { return nil }
func (m *mockBrushDB) GetBrushingTimesByUserUUID(ctx context.Context, userUUID uuid.UUID) ([]entity.BrushTime, error) {
	if m.GetBrushingTimesFn != nil {
		return m.GetBrushingTimesFn(ctx, userUUID)
	}
	return nil, nil
}
func (m *mockBrushDB) UpdateBrushingTime(ctx context.Context, bt entity.BrushTime) error { return nil }
func (m *mockBrushDB) ReplaceAllBrushingTimes(ctx context.Context, userUUID uuid.UUID, times []entity.BrushTime) error {
	if m.ReplaceAllFn != nil {
		return m.ReplaceAllFn(ctx, userUUID, times)
	}
	return nil
}

type mockCalendarDB struct {
	CreateCalendarEventFn         func(ctx context.Context, event entity.CalendarEvent) (*entity.CalendarEvent, error)
	GetCalendarEventsByUserUUIDFn func(ctx context.Context, userUUID uuid.UUID) ([]entity.CalendarEvent, error)
	DeleteCalendarEventFn         func(ctx context.Context, eventID, userUUID uuid.UUID) error
}

var _ database.CalendarDatabase = (*mockCalendarDB)(nil)

func (m *mockCalendarDB) CreateCalendarEvent(ctx context.Context, event entity.CalendarEvent) (*entity.CalendarEvent, error) {
	if m.CreateCalendarEventFn != nil {
		return m.CreateCalendarEventFn(ctx, event)
	}
	return &event, nil
}

func (m *mockCalendarDB) GetCalendarEventsByUserUUID(ctx context.Context, userUUID uuid.UUID) ([]entity.CalendarEvent, error) {
	if m.GetCalendarEventsByUserUUIDFn != nil {
		return m.GetCalendarEventsByUserUUIDFn(ctx, userUUID)
	}
	return nil, nil
}

func (m *mockCalendarDB) DeleteCalendarEvent(ctx context.Context, eventID, userUUID uuid.UUID) error {
	if m.DeleteCalendarEventFn != nil {
		return m.DeleteCalendarEventFn(ctx, eventID, userUUID)
	}
	return nil
}

// --- Helpers ---

// newUserUC собирает UserUseCase с пустым моком календаря.
func newUserUC(userDB database.UserDatabase, toothDB database.ToothDatabase, brushDB database.BrushingTimeDatabase) *usecase.UserUseCase {
	return usecase.NewUserUseCase(userDB, toothDB, brushDB, &mockCalendarDB{})
}

func ctxWithClaims(userID uuid.UUID) context.Context {
	return context.WithValue(context.Background(), sharedjwt.CtxKeyClaims{}, &sharedjwt.AccessClaims{UserID: userID})
}

// --- Tests: CreateUserInitially ---

func TestCreateUserInitially_Success(t *testing.T) {
	var capturedUser entity.User
	db := &mockUserDB{
		CreateUserFn: func(ctx context.Context, user entity.User) error {
			capturedUser = user
			return nil
		},
	}
	uc := newUserUC(db, &mockToothDB{}, &mockBrushDB{})
	user := entity.User{UUID: uuid.New(), Phone: "+71234567890"}
	err := uc.CreateUserInitially(context.Background(), user)
	require.NoError(t, err)

	assert.Equal(t, 32, len(capturedUser.Teeth))
	for i, tooth := range capturedUser.Teeth {
		assert.Equal(t, entity.ToothTypeMolar, tooth.ToothType)
		assert.Equal(t, entity.ToothStateHealthy, tooth.ToothState)
		assert.Equal(t, int16(i), tooth.ToothIndex)
		assert.Equal(t, user.UUID, tooth.UserUUID)
	}
}

func TestCreateUserInitially_DBError(t *testing.T) {
	db := &mockUserDB{
		CreateUserFn: func(ctx context.Context, user entity.User) error {
			return errors.New("db error")
		},
	}
	uc := newUserUC(db, &mockToothDB{}, &mockBrushDB{})
	err := uc.CreateUserInitially(context.Background(), entity.User{UUID: uuid.New()})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}

// --- Tests: GetUserByUUID ---

func TestGetUserByUUID_Success(t *testing.T) {
	userID := uuid.New()
	expected := &entity.User{UUID: userID, Phone: "+71234567890"}
	db := &mockUserDB{
		GetUserByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.User, error) {
			assert.Equal(t, userID, id)
			return expected, nil
		},
	}
	uc := newUserUC(db, &mockToothDB{}, &mockBrushDB{})
	result, err := uc.GetUserByUUID(ctxWithClaims(userID))
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestGetUserByUUID_NoClaims(t *testing.T) {
	uc := newUserUC(&mockUserDB{}, &mockToothDB{}, &mockBrushDB{})
	result, err := uc.GetUserByUUID(context.Background())
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, usecase.ErrNoAuthClaims))
}

func TestGetUserByUUID_DBError(t *testing.T) {
	userID := uuid.New()
	db := &mockUserDB{
		GetUserByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.User, error) {
			return nil, errors.New("db error")
		},
	}
	uc := newUserUC(db, &mockToothDB{}, &mockBrushDB{})
	result, err := uc.GetUserByUUID(ctxWithClaims(userID))
	assert.Nil(t, result)
	require.Error(t, err)
}

// --- Tests: UpdateProfile ---

func TestUpdateProfile_Success(t *testing.T) {
	userID := uuid.New()
	var capturedID uuid.UUID
	db := &mockUserDB{
		UpdateUserFn: func(ctx context.Context, id uuid.UUID, body dto.UpdateProfileBody) error {
			capturedID = id
			return nil
		},
	}
	uc := newUserUC(db, &mockToothDB{}, &mockBrushDB{})
	err := uc.UpdateProfile(ctxWithClaims(userID), dto.UpdateProfileBody{})
	require.NoError(t, err)
	assert.Equal(t, userID, capturedID)
}

func TestUpdateProfile_NoClaims(t *testing.T) {
	updateCalled := false
	db := &mockUserDB{
		UpdateUserFn: func(ctx context.Context, id uuid.UUID, body dto.UpdateProfileBody) error {
			updateCalled = true
			return nil
		},
	}
	uc := newUserUC(db, &mockToothDB{}, &mockBrushDB{})
	err := uc.UpdateProfile(context.Background(), dto.UpdateProfileBody{})
	assert.True(t, errors.Is(err, usecase.ErrNoAuthClaims))
	assert.False(t, updateCalled)
}

func TestUpdateProfile_DBError(t *testing.T) {
	userID := uuid.New()
	db := &mockUserDB{
		UpdateUserFn: func(ctx context.Context, id uuid.UUID, body dto.UpdateProfileBody) error {
			return errors.New("update error")
		},
	}
	uc := newUserUC(db, &mockToothDB{}, &mockBrushDB{})
	err := uc.UpdateProfile(ctxWithClaims(userID), dto.UpdateProfileBody{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "update error")
}

// --- Tests: GetTeeth ---

func TestGetTeeth_Success(t *testing.T) {
	userID := uuid.New()
	expected := []entity.Tooth{{UserUUID: userID, ToothIndex: 0, ToothState: entity.ToothStateHealthy}}
	toothDB := &mockToothDB{
		GetTeethByUserUUIDFn: func(ctx context.Context, id uuid.UUID) ([]entity.Tooth, error) {
			assert.Equal(t, userID, id)
			return expected, nil
		},
	}
	uc := newUserUC(&mockUserDB{}, toothDB, &mockBrushDB{})
	result, err := uc.GetTeeth(ctxWithClaims(userID))
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestGetTeeth_NoClaims(t *testing.T) {
	uc := newUserUC(&mockUserDB{}, &mockToothDB{}, &mockBrushDB{})
	result, err := uc.GetTeeth(context.Background())
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, usecase.ErrNoAuthClaims))
}

func TestGetTeeth_DBError(t *testing.T) {
	userID := uuid.New()
	toothDB := &mockToothDB{
		GetTeethByUserUUIDFn: func(ctx context.Context, id uuid.UUID) ([]entity.Tooth, error) {
			return nil, errors.New("db error")
		},
	}
	uc := newUserUC(&mockUserDB{}, toothDB, &mockBrushDB{})
	result, err := uc.GetTeeth(ctxWithClaims(userID))
	assert.Nil(t, result)
	require.Error(t, err)
}

// --- Tests: UpdateTeeth ---

func TestUpdateTeeth_Success(t *testing.T) {
	userID := uuid.New()
	teeth := make([]dto.ToothUpdateItem, 32)
	for i := range teeth {
		teeth[i] = dto.ToothUpdateItem{ToothState: string(entity.ToothStateHealthy)}
	}

	var captured []entity.Tooth
	toothDB := &mockToothDB{
		UpdateToothFn: func(ctx context.Context, tooth entity.Tooth) error {
			captured = append(captured, tooth)
			return nil
		},
	}
	uc := newUserUC(&mockUserDB{}, toothDB, &mockBrushDB{})
	err := uc.UpdateTeeth(ctxWithClaims(userID), dto.UpdateTeethBody{Teeth: teeth})
	require.NoError(t, err)
	assert.Len(t, captured, 32)
	for i, tooth := range captured {
		assert.Equal(t, int16(i), tooth.ToothIndex)
		assert.Equal(t, userID, tooth.UserUUID)
		assert.Equal(t, entity.ToothStateHealthy, tooth.ToothState)
	}
}

func TestUpdateTeeth_NoClaims(t *testing.T) {
	uc := newUserUC(&mockUserDB{}, &mockToothDB{}, &mockBrushDB{})
	err := uc.UpdateTeeth(context.Background(), dto.UpdateTeethBody{Teeth: []dto.ToothUpdateItem{}})
	assert.True(t, errors.Is(err, usecase.ErrNoAuthClaims))
}

func TestUpdateTeeth_DBError(t *testing.T) {
	userID := uuid.New()
	toothDB := &mockToothDB{
		UpdateToothFn: func(ctx context.Context, tooth entity.Tooth) error {
			return errors.New("update error")
		},
	}
	uc := newUserUC(&mockUserDB{}, toothDB, &mockBrushDB{})
	err := uc.UpdateTeeth(ctxWithClaims(userID), dto.UpdateTeethBody{Teeth: []dto.ToothUpdateItem{{ToothState: string(entity.ToothStateHealthy)}}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "update error")
}

// --- Tests: GetBrushSchedule ---

func TestGetBrushSchedule_Success(t *testing.T) {
	userID := uuid.New()
	expected := []entity.BrushTime{{UserUUID: userID}}
	brushDB := &mockBrushDB{
		GetBrushingTimesFn: func(ctx context.Context, id uuid.UUID) ([]entity.BrushTime, error) {
			assert.Equal(t, userID, id)
			return expected, nil
		},
	}
	uc := newUserUC(&mockUserDB{}, &mockToothDB{}, brushDB)
	result, err := uc.GetBrushSchedule(ctxWithClaims(userID))
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestGetBrushSchedule_NoClaims(t *testing.T) {
	uc := newUserUC(&mockUserDB{}, &mockToothDB{}, &mockBrushDB{})
	result, err := uc.GetBrushSchedule(context.Background())
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, usecase.ErrNoAuthClaims))
}

func TestGetBrushSchedule_DBError(t *testing.T) {
	userID := uuid.New()
	brushDB := &mockBrushDB{
		GetBrushingTimesFn: func(ctx context.Context, id uuid.UUID) ([]entity.BrushTime, error) {
			return nil, errors.New("db error")
		},
	}
	uc := newUserUC(&mockUserDB{}, &mockToothDB{}, brushDB)
	result, err := uc.GetBrushSchedule(ctxWithClaims(userID))
	assert.Nil(t, result)
	require.Error(t, err)
}

// --- Tests: UpdateBrushSchedule ---

func TestUpdateBrushSchedule_Success(t *testing.T) {
	userID := uuid.New()
	var capturedTimes []entity.BrushTime
	brushDB := &mockBrushDB{
		ReplaceAllFn: func(ctx context.Context, id uuid.UUID, times []entity.BrushTime) error {
			capturedTimes = times
			return nil
		},
	}
	uc := newUserUC(&mockUserDB{}, &mockToothDB{}, brushDB)
	err := uc.UpdateBrushSchedule(ctxWithClaims(userID), dto.UpdateBrushScheduleBody{TimesBrush: []string{"08:00", "21:00"}})
	require.NoError(t, err)
	assert.Len(t, capturedTimes, 2)
}

func TestUpdateBrushSchedule_NoClaims(t *testing.T) {
	uc := newUserUC(&mockUserDB{}, &mockToothDB{}, &mockBrushDB{})
	err := uc.UpdateBrushSchedule(context.Background(), dto.UpdateBrushScheduleBody{TimesBrush: []string{"08:00"}})
	assert.True(t, errors.Is(err, usecase.ErrNoAuthClaims))
}

func TestUpdateBrushSchedule_InvalidTime(t *testing.T) {
	userID := uuid.New()
	uc := newUserUC(&mockUserDB{}, &mockToothDB{}, &mockBrushDB{})
	err := uc.UpdateBrushSchedule(ctxWithClaims(userID), dto.UpdateBrushScheduleBody{TimesBrush: []string{"not-a-time"}})
	require.Error(t, err)
}

func TestUpdateBrushSchedule_DBError(t *testing.T) {
	userID := uuid.New()
	brushDB := &mockBrushDB{
		ReplaceAllFn: func(ctx context.Context, id uuid.UUID, times []entity.BrushTime) error {
			return errors.New("db error")
		},
	}
	uc := newUserUC(&mockUserDB{}, &mockToothDB{}, brushDB)
	err := uc.UpdateBrushSchedule(ctxWithClaims(userID), dto.UpdateBrushScheduleBody{TimesBrush: []string{"08:00"}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}
