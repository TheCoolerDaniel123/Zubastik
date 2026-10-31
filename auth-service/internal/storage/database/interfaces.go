package database

import (
	"auth-service/internal/entity"
	"context"

	"github.com/google/uuid"
)

type UserDatabase interface {
	GetUserByPhone(ctx context.Context, phone string) (*entity.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	CreateUser(ctx context.Context, user entity.User) error
}
