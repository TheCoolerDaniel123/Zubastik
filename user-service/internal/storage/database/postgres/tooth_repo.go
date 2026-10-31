package postgres

import (
	"context"

	"github.com/google/uuid"
	pgclient "shared/clients/postgres"
	"shared/entity"
)

type ToothRepo struct {
	cli *pgclient.PostgresClient
}

func NewToothRepo(cli *pgclient.PostgresClient) *ToothRepo {
	return &ToothRepo{cli: cli}
}

func (r *ToothRepo) CreateInitialTeeth(ctx context.Context, userUUID uuid.UUID) ([]entity.Tooth, error) {
	teeth := make([]entity.Tooth, 32)
	for i := range teeth {
		teeth[i] = entity.Tooth{
			UserUUID:   userUUID,
			ToothIndex: int16(i),
			ToothType:  entity.ToothTypeMolar,
			ToothState: entity.ToothStateHealthy,
		}
	}

	tx, err := r.cli.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	for _, tooth := range teeth {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO teeth (user_uuid, tooth_index, tooth_type, tooth_state)
			 VALUES ($1, $2, $3, $4)`,
			tooth.UserUUID,
			tooth.ToothIndex,
			tooth.ToothType,
			tooth.ToothState,
		)
		if err != nil {
			return nil, err
		}
	}

	return teeth, tx.Commit()
}

func (r *ToothRepo) GetTeethByUserUUID(ctx context.Context, userUUID uuid.UUID) ([]entity.Tooth, error) {
	rows, err := r.cli.DB.QueryContext(ctx,
		`SELECT tooth_index, tooth_type, tooth_state, note FROM teeth WHERE user_uuid=$1`,
		userUUID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teeth []entity.Tooth
	for rows.Next() {
		tooth := entity.Tooth{UserUUID: userUUID}
		if err := rows.Scan(&tooth.ToothIndex, &tooth.ToothType, &tooth.ToothState, &tooth.ToothNote); err != nil {
			return nil, err
		}
		teeth = append(teeth, tooth)
	}
	return teeth, nil
}

func (r *ToothRepo) UpdateTooth(ctx context.Context, tooth entity.Tooth) error {
	_, err := r.cli.DB.ExecContext(ctx,
		`UPDATE teeth SET tooth_state=$1, note=$2 WHERE user_uuid=$3 AND tooth_index=$4`,
		tooth.ToothState,
		tooth.ToothNote,
		tooth.UserUUID,
		tooth.ToothIndex,
	)
	return err
}
