package usecase

import (
	"auth-service/internal/entity"
	"context"
)

type SMSSender interface {
	SendSMS(phone string, code string) error
}

type RemoteUserDatabase interface {
	CreateUser(ctx context.Context, user entity.User) error
}
