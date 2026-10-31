package usecase_test

import (
	"auth-service/internal/config"
	"auth-service/internal/entity"
	"auth-service/internal/storage/database"
	"auth-service/internal/usecase"
	"context"
	"errors"
	sharedjwt "shared/jwt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Mocks ---

type mockCodeRepo struct {
	SetValueFn    func(ctx context.Context, key, value string, ttl time.Duration) error
	GetValueFn    func(ctx context.Context, key string) (*string, error)
	DeleteValueFn func(ctx context.Context, key string) error
}

func (m *mockCodeRepo) SetValue(ctx context.Context, key, value string, ttl time.Duration) error {
	return m.SetValueFn(ctx, key, value, ttl)
}

func (m *mockCodeRepo) GetValue(ctx context.Context, key string) (*string, error) {
	return m.GetValueFn(ctx, key)
}

func (m *mockCodeRepo) DeleteValue(ctx context.Context, key string) error {
	return m.DeleteValueFn(ctx, key)
}

type mockUserRepo struct {
	GetUserByPhoneFn func(ctx context.Context, phone string) (*entity.User, error)
	GetUserByIDFn    func(ctx context.Context, id uuid.UUID) (*entity.User, error)
	CreateUserFn     func(ctx context.Context, user entity.User) error
}

func (m *mockUserRepo) GetUserByPhone(ctx context.Context, phone string) (*entity.User, error) {
	return m.GetUserByPhoneFn(ctx, phone)
}

func (m *mockUserRepo) GetUserByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	return m.GetUserByIDFn(ctx, id)
}

func (m *mockUserRepo) CreateUser(ctx context.Context, user entity.User) error {
	return m.CreateUserFn(ctx, user)
}

type mockRemoteRepo struct {
	CreateUserFn func(ctx context.Context, user entity.User) error
}

func (m *mockRemoteRepo) CreateUser(ctx context.Context, user entity.User) error {
	return m.CreateUserFn(ctx, user)
}

type mockSMS struct {
	SendSMSFn func(phone, code string) error
}

func (m *mockSMS) SendSMS(phone, code string) error {
	return m.SendSMSFn(phone, code)
}

// --- Helpers ---

func testAuthConfig() *config.Auth {
	return &config.Auth{
		JWTSecret:           "test-secret",
		AccessTokenTTL:      15 * time.Minute,
		RefreshTokenTTL:     24 * time.Hour,
		VerificationCodeTTL: 5 * time.Minute,
	}
}

func testUseCase(codeRepo *mockCodeRepo, userRepo *mockUserRepo, remoteRepo *mockRemoteRepo, sms *mockSMS) *usecase.AuthUseCase {
	return usecase.NewAuthUseCase(remoteRepo, codeRepo, userRepo, sms, testAuthConfig())
}

// --- Tests: SendAuthorizationCode ---

func TestSendAuthorizationCode_Success(t *testing.T) {
	codeRepo := &mockCodeRepo{
		SetValueFn: func(ctx context.Context, key, value string, ttl time.Duration) error {
			return nil
		},
	}
	sms := &mockSMS{
		SendSMSFn: func(phone, code string) error {
			return nil
		},
	}
	uc := testUseCase(codeRepo, &mockUserRepo{}, &mockRemoteRepo{}, sms)
	err := uc.SendAuthorizationCode(context.Background(), "+71234567890")
	assert.NoError(t, err)
}

func TestSendAuthorizationCode_SMSError(t *testing.T) {
	codeRepo := &mockCodeRepo{
		SetValueFn: func(ctx context.Context, key, value string, ttl time.Duration) error {
			return nil
		},
	}
	sms := &mockSMS{
		SendSMSFn: func(phone, code string) error {
			return errors.New("sms failed")
		},
	}
	uc := testUseCase(codeRepo, &mockUserRepo{}, &mockRemoteRepo{}, sms)
	err := uc.SendAuthorizationCode(context.Background(), "+71234567890")
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "sms failed"))
}

func TestSendAuthorizationCode_CacheError(t *testing.T) {
	smsCalled := false
	codeRepo := &mockCodeRepo{
		SetValueFn: func(ctx context.Context, key, value string, ttl time.Duration) error {
			return errors.New("cache error")
		},
	}
	sms := &mockSMS{
		SendSMSFn: func(phone, code string) error {
			smsCalled = true
			return nil
		},
	}
	uc := testUseCase(codeRepo, &mockUserRepo{}, &mockRemoteRepo{}, sms)
	err := uc.SendAuthorizationCode(context.Background(), "+71234567890")
	require.Error(t, err)
	assert.False(t, smsCalled)
}

// --- Tests: VerifyCode ---

func ptrStr(s string) *string { return &s }

func TestVerifyCode_ExistingUser(t *testing.T) {
	existingUser := &entity.User{UUID: uuid.New(), Phone: "+71234567890", Role: entity.RoleUser}
	codeRepo := &mockCodeRepo{
		GetValueFn:    func(ctx context.Context, key string) (*string, error) { return ptrStr("1234"), nil },
		DeleteValueFn: func(ctx context.Context, key string) error { return nil },
	}
	userRepo := &mockUserRepo{
		GetUserByPhoneFn: func(ctx context.Context, phone string) (*entity.User, error) {
			return existingUser, nil
		},
	}
	uc := testUseCase(codeRepo, userRepo, &mockRemoteRepo{}, &mockSMS{})
	tokenPair, isNewUser, err := uc.VerifyCode(context.Background(), "+71234567890", "1234")
	require.NoError(t, err)
	assert.False(t, isNewUser)
	assert.NotNil(t, tokenPair)
	assert.NotEmpty(t, tokenPair.AccessToken)
}

func TestVerifyCode_NewUser(t *testing.T) {
	codeRepo := &mockCodeRepo{
		GetValueFn:    func(ctx context.Context, key string) (*string, error) { return ptrStr("1234"), nil },
		DeleteValueFn: func(ctx context.Context, key string) error { return nil },
	}
	userRepo := &mockUserRepo{
		GetUserByPhoneFn: func(ctx context.Context, phone string) (*entity.User, error) {
			return nil, database.ErrUserNotFound
		},
		CreateUserFn: func(ctx context.Context, user entity.User) error { return nil },
	}
	remoteRepo := &mockRemoteRepo{
		CreateUserFn: func(ctx context.Context, user entity.User) error { return nil },
	}
	uc := testUseCase(codeRepo, userRepo, remoteRepo, &mockSMS{})
	tokenPair, isNewUser, err := uc.VerifyCode(context.Background(), "+71234567890", "1234")
	require.NoError(t, err)
	assert.True(t, isNewUser)
	assert.NotNil(t, tokenPair)
}

func TestVerifyCode_WrongCode(t *testing.T) {
	codeRepo := &mockCodeRepo{
		GetValueFn: func(ctx context.Context, key string) (*string, error) { return ptrStr("9999"), nil },
	}
	uc := testUseCase(codeRepo, &mockUserRepo{}, &mockRemoteRepo{}, &mockSMS{})
	tokenPair, isNewUser, err := uc.VerifyCode(context.Background(), "+71234567890", "1234")
	require.Error(t, err)
	assert.True(t, errors.Is(err, usecase.ErrInvalidCode))
	assert.Nil(t, tokenPair)
	assert.False(t, isNewUser)
}

func TestVerifyCode_CodeNotFound(t *testing.T) {
	codeRepo := &mockCodeRepo{
		GetValueFn: func(ctx context.Context, key string) (*string, error) {
			return nil, errors.New("not found")
		},
	}
	uc := testUseCase(codeRepo, &mockUserRepo{}, &mockRemoteRepo{}, &mockSMS{})
	tokenPair, isNewUser, err := uc.VerifyCode(context.Background(), "+71234567890", "1234")
	require.Error(t, err)
	assert.Nil(t, tokenPair)
	assert.False(t, isNewUser)
}

func TestVerifyCode_DeleteError(t *testing.T) {
	codeRepo := &mockCodeRepo{
		GetValueFn:    func(ctx context.Context, key string) (*string, error) { return ptrStr("1234"), nil },
		DeleteValueFn: func(ctx context.Context, key string) error { return errors.New("delete error") },
	}
	uc := testUseCase(codeRepo, &mockUserRepo{}, &mockRemoteRepo{}, &mockSMS{})
	tokenPair, isNewUser, err := uc.VerifyCode(context.Background(), "+71234567890", "1234")
	require.Error(t, err)
	assert.Nil(t, tokenPair)
	assert.False(t, isNewUser)
}

func TestVerifyCode_CreateUserError(t *testing.T) {
	codeRepo := &mockCodeRepo{
		GetValueFn:    func(ctx context.Context, key string) (*string, error) { return ptrStr("1234"), nil },
		DeleteValueFn: func(ctx context.Context, key string) error { return nil },
	}
	userRepo := &mockUserRepo{
		GetUserByPhoneFn: func(ctx context.Context, phone string) (*entity.User, error) {
			return nil, database.ErrUserNotFound
		},
		CreateUserFn: func(ctx context.Context, user entity.User) error {
			return errors.New("db create error")
		},
	}
	uc := testUseCase(codeRepo, userRepo, &mockRemoteRepo{}, &mockSMS{})
	tokenPair, isNewUser, err := uc.VerifyCode(context.Background(), "+71234567890", "1234")
	require.Error(t, err)
	assert.Nil(t, tokenPair)
	assert.False(t, isNewUser)
}

func TestVerifyCode_RemoteCreateUserError(t *testing.T) {
	codeRepo := &mockCodeRepo{
		GetValueFn:    func(ctx context.Context, key string) (*string, error) { return ptrStr("1234"), nil },
		DeleteValueFn: func(ctx context.Context, key string) error { return nil },
	}
	userRepo := &mockUserRepo{
		GetUserByPhoneFn: func(ctx context.Context, phone string) (*entity.User, error) {
			return nil, database.ErrUserNotFound
		},
		CreateUserFn: func(ctx context.Context, user entity.User) error { return nil },
	}
	remoteRepo := &mockRemoteRepo{
		CreateUserFn: func(ctx context.Context, user entity.User) error {
			return errors.New("remote error")
		},
	}
	uc := testUseCase(codeRepo, userRepo, remoteRepo, &mockSMS{})
	tokenPair, isNewUser, err := uc.VerifyCode(context.Background(), "+71234567890", "1234")
	require.Error(t, err)
	assert.Nil(t, tokenPair)
	assert.False(t, isNewUser)
}

// --- Tests: RefreshToken ---

func TestRefreshToken_Success(t *testing.T) {
	userID := uuid.New()
	conf := &sharedjwt.JWTConfig{
		JWTSecret:       "test-secret",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 24 * time.Hour,
	}
	tokenPair, err := sharedjwt.GenerateTokenPair(conf,
		sharedjwt.AccessClaims{UserID: userID},
		sharedjwt.RefreshClaims{UserID: userID},
	)
	require.NoError(t, err)

	userRepo := &mockUserRepo{
		GetUserByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.User, error) {
			return &entity.User{UUID: userID, Phone: "+71234567890"}, nil
		},
	}
	uc := testUseCase(&mockCodeRepo{}, userRepo, &mockRemoteRepo{}, &mockSMS{})
	newPair, err := uc.RefreshToken(context.Background(), tokenPair.AccessToken, tokenPair.RefreshToken)
	require.NoError(t, err)
	assert.NotNil(t, newPair)
	assert.NotEmpty(t, newPair.AccessToken)
}

func TestRefreshToken_InvalidAccessToken(t *testing.T) {
	uc := testUseCase(&mockCodeRepo{}, &mockUserRepo{}, &mockRemoteRepo{}, &mockSMS{})
	_, err := uc.RefreshToken(context.Background(), "not-a-token", "not-a-token")
	require.Error(t, err)
}

func TestRefreshToken_InvalidRefreshToken(t *testing.T) {
	userID := uuid.New()
	conf := &sharedjwt.JWTConfig{
		JWTSecret:       "test-secret",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 24 * time.Hour,
	}
	tokenPair, err := sharedjwt.GenerateTokenPair(conf,
		sharedjwt.AccessClaims{UserID: userID},
		sharedjwt.RefreshClaims{UserID: userID},
	)
	require.NoError(t, err)

	uc := testUseCase(&mockCodeRepo{}, &mockUserRepo{}, &mockRemoteRepo{}, &mockSMS{})
	_, err = uc.RefreshToken(context.Background(), tokenPair.AccessToken, "garbage-refresh-token")
	require.Error(t, err)
}

func TestRefreshToken_DifferentUsers(t *testing.T) {
	userA := uuid.New()
	userB := uuid.New()
	conf := &sharedjwt.JWTConfig{
		JWTSecret:       "test-secret",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 24 * time.Hour,
	}
	pairA, err := sharedjwt.GenerateTokenPair(conf,
		sharedjwt.AccessClaims{UserID: userA},
		sharedjwt.RefreshClaims{UserID: userA},
	)
	require.NoError(t, err)
	pairB, err := sharedjwt.GenerateTokenPair(conf,
		sharedjwt.AccessClaims{UserID: userB},
		sharedjwt.RefreshClaims{UserID: userB},
	)
	require.NoError(t, err)

	uc := testUseCase(&mockCodeRepo{}, &mockUserRepo{}, &mockRemoteRepo{}, &mockSMS{})
	_, err = uc.RefreshToken(context.Background(), pairA.AccessToken, pairB.RefreshToken)
	require.Error(t, err)
	assert.True(t, errors.Is(err, usecase.ErrDifferentTokens))
}
