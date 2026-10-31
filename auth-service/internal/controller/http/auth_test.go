package http_test

import (
	"auth-service/internal/config"
	authhttp "auth-service/internal/controller/http"
	"auth-service/internal/entity"
	"auth-service/internal/storage/database"
	"auth-service/internal/usecase"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	sharedjwt "shared/jwt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Mocks (redefine for this package) ---

type testCodeRepo struct {
	SetValueFn    func(ctx context.Context, key, value string, ttl time.Duration) error
	GetValueFn    func(ctx context.Context, key string) (*string, error)
	DeleteValueFn func(ctx context.Context, key string) error
}

func (m *testCodeRepo) SetValue(ctx context.Context, key, value string, ttl time.Duration) error {
	if m.SetValueFn != nil {
		return m.SetValueFn(ctx, key, value, ttl)
	}
	return nil
}

func (m *testCodeRepo) GetValue(ctx context.Context, key string) (*string, error) {
	if m.GetValueFn != nil {
		return m.GetValueFn(ctx, key)
	}
	s := ""
	return &s, nil
}

func (m *testCodeRepo) DeleteValue(ctx context.Context, key string) error {
	if m.DeleteValueFn != nil {
		return m.DeleteValueFn(ctx, key)
	}
	return nil
}

type testUserRepo struct {
	GetUserByPhoneFn func(ctx context.Context, phone string) (*entity.User, error)
	GetUserByIDFn    func(ctx context.Context, id uuid.UUID) (*entity.User, error)
	CreateUserFn     func(ctx context.Context, user entity.User) error
}

func (m *testUserRepo) GetUserByPhone(ctx context.Context, phone string) (*entity.User, error) {
	if m.GetUserByPhoneFn != nil {
		return m.GetUserByPhoneFn(ctx, phone)
	}
	return nil, database.ErrUserNotFound
}

func (m *testUserRepo) GetUserByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	if m.GetUserByIDFn != nil {
		return m.GetUserByIDFn(ctx, id)
	}
	return nil, database.ErrUserNotFound
}

func (m *testUserRepo) CreateUser(ctx context.Context, user entity.User) error {
	if m.CreateUserFn != nil {
		return m.CreateUserFn(ctx, user)
	}
	return nil
}

type testRemoteRepo struct {
	CreateUserFn func(ctx context.Context, user entity.User) error
}

func (m *testRemoteRepo) CreateUser(ctx context.Context, user entity.User) error {
	if m.CreateUserFn != nil {
		return m.CreateUserFn(ctx, user)
	}
	return nil
}

type testSMS struct {
	SendSMSFn func(phone, code string) error
}

func (m *testSMS) SendSMS(phone, code string) error {
	if m.SendSMSFn != nil {
		return m.SendSMSFn(phone, code)
	}
	return nil
}

// --- Helper ---

func testConf() *config.Auth {
	return &config.Auth{
		JWTSecret:           "test-secret",
		AccessTokenTTL:      15 * time.Minute,
		RefreshTokenTTL:     24 * time.Hour,
		VerificationCodeTTL: 5 * time.Minute,
	}
}

func buildController(codeRepo *testCodeRepo, userRepo *testUserRepo, remoteRepo *testRemoteRepo, sms *testSMS) *authhttp.AuthController {
	uc := usecase.NewAuthUseCase(remoteRepo, codeRepo, userRepo, sms, testConf())
	return authhttp.NewAuthController(uc)
}

func ptrString(s string) *string { return &s }

// --- Tests ---

func TestSendCodeHandler_OK(t *testing.T) {
	codeRepo := &testCodeRepo{
		SetValueFn: func(ctx context.Context, key, value string, ttl time.Duration) error { return nil },
	}
	sms := &testSMS{
		SendSMSFn: func(phone, code string) error { return nil },
	}
	ctrl := buildController(codeRepo, &testUserRepo{}, &testRemoteRepo{}, sms)

	body := `{"Phone":"+71234567890"}`
	req := httptest.NewRequest(http.MethodPost, "/send-code", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ctrl.SendCodeHandler(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSendCodeHandler_InvalidJSON(t *testing.T) {
	ctrl := buildController(&testCodeRepo{}, &testUserRepo{}, &testRemoteRepo{}, &testSMS{})

	req := httptest.NewRequest(http.MethodPost, "/send-code", strings.NewReader("{invalid}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ctrl.SendCodeHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestVerifyCodeHandler_ExistingUser_OK(t *testing.T) {
	existingUser := &entity.User{UUID: uuid.New(), Phone: "+71234567890"}
	codeRepo := &testCodeRepo{
		GetValueFn:    func(ctx context.Context, key string) (*string, error) { return ptrString("1234"), nil },
		DeleteValueFn: func(ctx context.Context, key string) error { return nil },
	}
	userRepo := &testUserRepo{
		GetUserByPhoneFn: func(ctx context.Context, phone string) (*entity.User, error) {
			return existingUser, nil
		},
	}
	ctrl := buildController(codeRepo, userRepo, &testRemoteRepo{}, &testSMS{})

	body := `{"phone":"+71234567890","code":"1234"}`
	req := httptest.NewRequest(http.MethodPost, "/verify-code", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ctrl.VerifyCodeHandler(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, false, resp["is_new_user"])
	assert.NotEmpty(t, resp["access_token"])
}

func TestVerifyCodeHandler_NewUser_OK(t *testing.T) {
	codeRepo := &testCodeRepo{
		GetValueFn:    func(ctx context.Context, key string) (*string, error) { return ptrString("1234"), nil },
		DeleteValueFn: func(ctx context.Context, key string) error { return nil },
	}
	userRepo := &testUserRepo{
		GetUserByPhoneFn: func(ctx context.Context, phone string) (*entity.User, error) {
			return nil, database.ErrUserNotFound
		},
		CreateUserFn: func(ctx context.Context, user entity.User) error { return nil },
	}
	remoteRepo := &testRemoteRepo{
		CreateUserFn: func(ctx context.Context, user entity.User) error { return nil },
	}
	ctrl := buildController(codeRepo, userRepo, remoteRepo, &testSMS{})

	body := `{"phone":"+71234567890","code":"1234"}`
	req := httptest.NewRequest(http.MethodPost, "/verify-code", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ctrl.VerifyCodeHandler(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, true, resp["is_new_user"])
}

func TestVerifyCodeHandler_WrongCode(t *testing.T) {
	codeRepo := &testCodeRepo{
		GetValueFn: func(ctx context.Context, key string) (*string, error) { return ptrString("9999"), nil },
	}
	ctrl := buildController(codeRepo, &testUserRepo{}, &testRemoteRepo{}, &testSMS{})

	body := `{"phone":"+71234567890","code":"1234"}`
	req := httptest.NewRequest(http.MethodPost, "/verify-code", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ctrl.VerifyCodeHandler(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestVerifyCodeHandler_InvalidBody(t *testing.T) {
	ctrl := buildController(&testCodeRepo{}, &testUserRepo{}, &testRemoteRepo{}, &testSMS{})

	req := httptest.NewRequest(http.MethodPost, "/verify-code", bytes.NewReader([]byte{}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ctrl.VerifyCodeHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRefreshTokenHandler_OK(t *testing.T) {
	userID := uuid.New()
	jwtConf := &sharedjwt.JWTConfig{
		JWTSecret:       "test-secret",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 24 * time.Hour,
	}
	tokenPair, err := sharedjwt.GenerateTokenPair(jwtConf,
		sharedjwt.AccessClaims{UserID: userID},
		sharedjwt.RefreshClaims{UserID: userID},
	)
	require.NoError(t, err)

	userRepo := &testUserRepo{
		GetUserByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.User, error) {
			return &entity.User{UUID: userID, Phone: "+71234567890"}, nil
		},
	}
	ctrl := buildController(&testCodeRepo{}, userRepo, &testRemoteRepo{}, &testSMS{})

	reqBody, err := json.Marshal(map[string]string{
		"access_token":  tokenPair.AccessToken,
		"refresh_token": tokenPair.RefreshToken,
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ctrl.RefreshTokenHandler(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.NotEmpty(t, resp["access_token"])
}

func TestRefreshTokenHandler_InvalidToken(t *testing.T) {
	ctrl := buildController(&testCodeRepo{}, &testUserRepo{}, &testRemoteRepo{}, &testSMS{})

	reqBody, err := json.Marshal(map[string]string{
		"access_token":  "garbage",
		"refresh_token": "garbage",
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ctrl.RefreshTokenHandler(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
