package jwt

import "errors"

var (
	ErrInvalidTokenPair = errors.New("invalid token pair")
	ErrTokenExpired     = errors.New("token has expired")
)
