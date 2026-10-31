package postgres

import (
	"auth-service/internal/entity"
	"auth-service/internal/storage/database"
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	pgclient "shared/clients/postgres"
)

type UserRepo struct {
	cli *pgclient.PostgresClient
}

func NewUserRepo(cli *pgclient.PostgresClient) *UserRepo {
	return &UserRepo{cli: cli}
}

func (r *UserRepo) GetUserByPhone(ctx context.Context, phone string) (*entity.User, error) {
	var user entity.User
	err := r.cli.DB.QueryRowContext(ctx, `SELECT uuid, phone, role FROM users WHERE phone=$1`, phone).
		Scan(&user.UUID, &user.Phone, &user.Role)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, database.ErrUserNotFound
	}
	return &user, err
}

func (r *UserRepo) GetUserByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	var user entity.User
	err := r.cli.DB.QueryRowContext(ctx, `SELECT uuid, phone, role FROM users WHERE uuid=$1`, id).
		Scan(&user.UUID, &user.Phone, &user.Role)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, database.ErrUserNotFound
	}
	return &user, err
}

func (r *UserRepo) CreateUser(ctx context.Context, user entity.User) error {
	_, err := r.cli.DB.ExecContext(ctx,
		`INSERT INTO users (uuid, phone, role) VALUES ($1, $2, $3)`,
		user.UUID, user.Phone, user.Role,
	)
	return err
}
