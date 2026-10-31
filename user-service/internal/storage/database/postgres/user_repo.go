package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	pgclient "shared/clients/postgres"
	"shared/entity"
	"user-service/api/dto"
	"user-service/internal/storage/database"
)

type UserRepo struct {
	cli *pgclient.PostgresClient
}

func NewUserRepo(cli *pgclient.PostgresClient) *UserRepo {
	return &UserRepo{cli: cli}
}

func (r *UserRepo) GetUserByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	var user entity.User

	var name sql.NullString
	err := r.cli.DB.QueryRowContext(
		ctx,
		`SELECT uuid, phone, role, name, age, gender
		 FROM users
		 WHERE uuid=$1`,
		id,
	).Scan(
		&user.UUID,
		&user.Phone,
		&user.Role,
		&name,
		&user.Age,
		&user.Gender,
	)
	user.Name = name.String

	if errors.Is(err, sql.ErrNoRows) {
		return nil, database.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	rows, err := r.cli.DB.QueryContext(ctx,
		`SELECT tooth_index, tooth_type, tooth_state, note
		 FROM teeth
		 WHERE user_uuid=$1
		 ORDER BY tooth_index`,
		id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var tooth entity.Tooth
		if err := rows.Scan(&tooth.ToothIndex, &tooth.ToothType, &tooth.ToothState, &tooth.ToothNote); err != nil {
			return nil, err
		}
		tooth.UserUUID = id
		user.Teeth = append(user.Teeth, tooth)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	timeRows, err := r.cli.DB.QueryContext(ctx,
		`SELECT user_uuid, time
		 FROM brushing_times
		 WHERE user_uuid=$1
		 ORDER BY time`,
		id)
	if err != nil {
		return nil, err
	}
	defer timeRows.Close()

	for timeRows.Next() {
		var bt entity.BrushTime
		if err := timeRows.Scan(&bt.UserUUID, &bt.Time); err != nil {
			return nil, err
		}
		user.BrushTimes = append(user.BrushTimes, bt)
	}
	if err := timeRows.Err(); err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepo) CreateUser(ctx context.Context, user entity.User) error {
	tx, err := r.cli.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx,
		`INSERT INTO users (uuid, phone, role, age, gender)
		 VALUES ($1, $2, $3, $4, $5)`,
		user.UUID,
		user.Phone,
		user.Role,
		user.Age,
		user.Gender,
	)
	if err != nil {
		return err
	}

	for _, tooth := range user.Teeth {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO teeth (user_uuid, tooth_index, tooth_type, tooth_state)
			 VALUES ($1, $2, $3, $4)`,
			user.UUID,
			tooth.ToothIndex,
			tooth.ToothType,
			tooth.ToothState,
		)
		if err != nil {
			return err
		}
	}

	for _, bt := range user.BrushTimes {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO brushing_times (user_uuid, time)
			 VALUES ($1, $2)`,
			user.UUID,
			bt.Time,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *UserRepo) UpdateUser(ctx context.Context, id uuid.UUID, updateInfo dto.UpdateProfileBody) error {
	setParts := []string{}
	args := []any{}
	argID := 1

	if updateInfo.Name != nil {
		setParts = append(setParts, fmt.Sprintf("name=$%d", argID))
		args = append(args, *updateInfo.Name)
		argID++
	}

	if updateInfo.Age != nil {
		setParts = append(setParts, fmt.Sprintf("age=$%d", argID))
		args = append(args, *updateInfo.Age)
		argID++
	}

	if updateInfo.Gender != nil {
		setParts = append(setParts, fmt.Sprintf("gender=$%d", argID))
		args = append(args, *updateInfo.Gender)
		argID++
	}

	if len(setParts) == 0 {
		return nil
	}

	query := fmt.Sprintf(`UPDATE users SET %s WHERE uuid=$%d`,
		strings.Join(setParts, ", "),
		argID)

	args = append(args, id)

	_, err := r.cli.DB.ExecContext(ctx, query, args...)
	return err
}
