package cache

import "errors"

var (
	ErrCodeExpired = errors.New("auth code has expired")
)
