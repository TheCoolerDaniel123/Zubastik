package dto

import (
	"time"

	"shared/jwt"
)

type ConfirmCodeBody struct {
	Phone string `json:"phone" validate:"required,e164"`
	Code  string `json:"code" validate:"required,len=4,numeric"`
}

type ConfirmCode struct {
	ConfirmCodeBody
	TTL time.Duration
}

type VerifyCodeResponse struct {
	*jwt.TokenPair
	IsNewUser bool `json:"is_new_user"`
}
