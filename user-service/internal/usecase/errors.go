package usecase

import "errors"

var (
	ErrNoAuthClaims = errors.New("no auth claims")
	ErrForbidden    = errors.New("forbidden")
)
