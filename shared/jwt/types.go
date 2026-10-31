package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type CtxKeyClaims struct{}

type JWTConfig struct {
	JWTSecret string

	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type AccessClaims struct {
	UserID uuid.UUID `json:"user_id"`
	Role   int       `json:"role"`
	jwt.RegisteredClaims
}

type RefreshClaims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
