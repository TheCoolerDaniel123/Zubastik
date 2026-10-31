package usecase

import "errors"

var (
	ErrInvalidCode = errors.New("invalid verification code")

	ErrNoAuthClaims    = errors.New("no auth claims")
	ErrDifferentTokens = errors.New("jwt tokens belong to different users")
)
