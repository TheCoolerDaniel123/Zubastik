package postgres

import (
	"context"

	"github.com/google/uuid"
	pgclient "shared/clients/postgres"
	"shared/entity"
)

type BrushingTimesRepo struct {
	cli *pgclient.PostgresClient
}

func NewBrushingTimesRepo(cli *pgclient.PostgresClient) *BrushingTimesRepo {
	return &BrushingTimesRepo{cli: cli}
}

func (r *BrushingTimesRepo) CreateBrushingTime(ctx context.Context, brushingTime entity.BrushTime) error {
	_, err := r.cli.DB.ExecContext(ctx,
		`INSERT INTO brushing_times (user_uuid, time) VALUES ($1, $2)`,
		brushingTime.UserUUID, brushingTime.Time,
	)
	return err
}

func (r *BrushingTimesRepo) GetBrushingTimesByUserUUID(ctx context.Context, userUUID uuid.UUID) ([]entity.BrushTime, error) {
	rows, err := r.cli.DB.QueryContext(ctx,
		`SELECT user_uuid, time FROM brushing_times WHERE user_uuid = $1`,
		userUUID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var times []entity.BrushTime
	for rows.Next() {
		var t entity.BrushTime
		if err := rows.Scan(&t.UserUUID, &t.Time); err != nil {
			return nil, err
		}
		times = append(times, t)
	}
	return times, nil
}

func (r *BrushingTimesRepo) UpdateBrushingTime(ctx context.Context, brushingTime entity.BrushTime) error {
	_, err := r.cli.DB.ExecContext(ctx,
		`UPDATE brushing_times SET time=$1 WHERE user_uuid=$2 AND time=$3`,
		brushingTime.Time, brushingTime.UserUUID, brushingTime.Time,
	)
	return err
}

func (r *BrushingTimesRepo) ReplaceAllBrushingTimes(ctx context.Context, userUUID uuid.UUID, times []entity.BrushTime) error {
	tx, err := r.cli.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err = tx.ExecContext(ctx, `DELETE FROM brushing_times WHERE user_uuid=$1`, userUUID); err != nil {
		return err
	}

	for _, t := range times {
		if _, err = tx.ExecContext(ctx, `INSERT INTO brushing_times (user_uuid, time) VALUES ($1, $2)`, userUUID, t.Time); err != nil {
			return err
		}
	}

	return tx.Commit()
}
