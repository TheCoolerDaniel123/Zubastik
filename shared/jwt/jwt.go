package jwt

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func GenerateTokenPair(config *JWTConfig, accessClaims AccessClaims, refreshClaims RefreshClaims) (tokenPair *TokenPair, err error) {
	now := time.Now()

	accessClaims.RegisteredClaims = jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(now.Add(config.AccessTokenTTL)),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
	}
	refreshClaims.RegisteredClaims = jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(now.Add(config.RefreshTokenTTL)),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString([]byte(config.JWTSecret))
	if err != nil {
		return &TokenPair{}, err
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(config.JWTSecret))
	if err != nil {
		return &TokenPair{}, err
	}
	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func ParseAndValidateAccessTokenWithoutExp(config *JWTConfig, tokenStr string) (*AccessClaims, error) {
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())

	token, err := parser.ParseWithClaims(tokenStr, &AccessClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.JWTSecret), nil
	})

	if err != nil {
		// Ignore expired token
		var verr *jwt.ValidationError
		if errors.As(err, &verr) {
			if verr.Errors == jwt.ValidationErrorExpired {
				// Validate token signature
				claims, ok := token.Claims.(*AccessClaims)
				if !ok {
					return nil, ErrInvalidTokenPair
				}
				return claims, nil
			}
		}
		return nil, err
	}
	claims, ok := token.Claims.(*AccessClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidTokenPair
	}
	return claims, nil
}

func ParseAndValidateAccessToken(config *JWTConfig, tokenStr string) (*AccessClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &AccessClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.JWTSecret), nil
	})

	if err != nil {
		var verr *jwt.ValidationError
		if errors.As(err, &verr) && verr.Errors&jwt.ValidationErrorExpired != 0 {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidTokenPair
	}

	claims, ok := token.Claims.(*AccessClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidTokenPair
	}
	return claims, nil
}

func ParseAndValidateRefreshToken(config *JWTConfig, tokenStr string) (*RefreshClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &RefreshClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.JWTSecret), nil
	})

	if err != nil {
		var verr *jwt.ValidationError
		if errors.As(err, &verr) && verr.Errors&jwt.ValidationErrorExpired != 0 {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidTokenPair
	}

	claims, ok := token.Claims.(*RefreshClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidTokenPair
	}

	return claims, nil
}

func GetClaimsFromCtx(ctx context.Context) *AccessClaims {
	claims, _ := ctx.Value(CtxKeyClaims{}).(*AccessClaims)
	return claims
}
