package usecase

import (
	"auth-service/internal/config"
	"auth-service/internal/entity"
	"auth-service/internal/storage/cache"
	"auth-service/internal/storage/database"
	e "shared/e"
	"shared/jwt"

	"github.com/google/uuid"

	"context"
	"errors"
	// "crypto/rand"
	// "fmt"
	// "math/big"
)

type AuthUseCase struct {
	remoteUserRepo RemoteUserDatabase
	codeRepo       cache.CodeDatabase
	userRepo       database.UserDatabase
	smsSender      SMSSender
	authConfig     *config.Auth
}

func NewAuthUseCase(rur RemoteUserDatabase, c_db cache.CodeDatabase, u_db database.UserDatabase, sms SMSSender, conf *config.Auth) *AuthUseCase {
	return &AuthUseCase{
		remoteUserRepo: rur,
		codeRepo:       c_db,
		userRepo:       u_db,
		smsSender:      sms,
		authConfig:     conf,
	}
}

func (uc *AuthUseCase) SendAuthorizationCode(ctx context.Context, phone string) (err error) {
	defer func() { err = e.WrapIfErr(err, "AuthUseCase.SendAuthorizationCode") }()

	// code, err := rand.Int(rand.Reader, big.NewInt(10000))
	// if err != nil {
	// 	code = big.NewInt(8394)
	// }
	// codeStr := fmt.Sprintf("%04d", code)

	codeStr := "1234"

	err = uc.codeRepo.SetValue(ctx, phone, codeStr, uc.authConfig.VerificationCodeTTL)
	if err != nil {
		return err
	}
	err = uc.smsSender.SendSMS(phone, codeStr)
	if err != nil {
		return err
	}
	return nil
}

func (uc *AuthUseCase) VerifyCode(ctx context.Context, phone, code string) (tokenPair *jwt.TokenPair, isNewUser bool, err error) {

	savedCode, err := uc.codeRepo.GetValue(ctx, phone)
	if err != nil {
		return nil, false, err
	}

	if *savedCode != code {
		return nil, false, ErrInvalidCode
	}

	if err := uc.codeRepo.DeleteValue(ctx, phone); err != nil {
		return nil, false, err
	}

	// Находим или создаём пользователя
	user, err := uc.userRepo.GetUserByPhone(ctx, phone)
	if err != nil {
		if errors.Is(err, database.ErrUserNotFound) {
			newUser := entity.User{UUID: uuid.New(), Phone: phone, Role: entity.RoleUser}
			if err := uc.userRepo.CreateUser(ctx, newUser); err != nil {
				return nil, false, err
			}
			if err := uc.remoteUserRepo.CreateUser(ctx, newUser); err != nil {
				return nil, false, err
			}
			user = &newUser
			isNewUser = true
		} else {
			return nil, false, err
		}
	}

	accessClaims := jwt.AccessClaims{UserID: user.UUID, Role: int(user.Role)}
	refreshClaims := jwt.RefreshClaims{UserID: user.UUID}

	conf := jwt.JWTConfig{
		JWTSecret:       uc.authConfig.JWTSecret,
		AccessTokenTTL:  uc.authConfig.AccessTokenTTL,
		RefreshTokenTTL: uc.authConfig.RefreshTokenTTL,
	}
	tokenPair, err = jwt.GenerateTokenPair(&conf, accessClaims, refreshClaims)
	if err != nil {
		return nil, false, err
	}
	return tokenPair, isNewUser, nil
}

func (uc *AuthUseCase) RefreshToken(ctx context.Context, accessTokenStr string, refreshTokenStr string) (*jwt.TokenPair, error) {
	conf := jwt.JWTConfig{
		JWTSecret:       uc.authConfig.JWTSecret,
		AccessTokenTTL:  uc.authConfig.AccessTokenTTL,
		RefreshTokenTTL: uc.authConfig.RefreshTokenTTL}

	accessClaims, err := jwt.ParseAndValidateAccessTokenWithoutExp(&conf, accessTokenStr)
	if err != nil {
		return nil, err
	}

	refreshClaims, err := jwt.ParseAndValidateRefreshToken(&conf, refreshTokenStr)
	if err != nil {
		return nil, err
	}

	if refreshClaims.UserID != accessClaims.UserID {
		return nil, ErrDifferentTokens
	}

	user, err := uc.userRepo.GetUserByID(ctx, refreshClaims.UserID)
	if err != nil {
		return nil, err
	}

	newAccessClaims := jwt.AccessClaims{
		UserID: user.UUID,
		Role:   int(user.Role),
	}

	newRefreshClaims := jwt.RefreshClaims{
		UserID: user.UUID,
	}

	return jwt.GenerateTokenPair(&conf, newAccessClaims, newRefreshClaims)
}
