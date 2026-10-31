package postgres

import (
	"context"

	"github.com/google/uuid"
	pgclient "shared/clients/postgres"
	"shared/entity"
)

type RecommendationRepo struct {
	cli *pgclient.PostgresClient
}

func NewRecommendationRepo(cli *pgclient.PostgresClient) *RecommendationRepo {
	return &RecommendationRepo{cli: cli}
}

func (r *RecommendationRepo) CreateRecommendation(ctx context.Context, rec entity.Recommendation) (*entity.Recommendation, error) {
	rec.ID = uuid.New()
	err := r.cli.DB.QueryRowContext(ctx,
		`INSERT INTO recommendations (id, doctor_uuid, patient_uuid, text)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, created_at`,
		rec.ID, rec.DoctorUUID, rec.PatientUUID, rec.Text,
	).Scan(&rec.ID, &rec.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *RecommendationRepo) GetRecommendationsByPatient(ctx context.Context, doctorUUID, patientUUID uuid.UUID) ([]entity.Recommendation, error) {
	rows, err := r.cli.DB.QueryContext(ctx,
		`SELECT id, doctor_uuid, patient_uuid, text, created_at
		 FROM recommendations
		 WHERE doctor_uuid=$1 AND patient_uuid=$2
		 ORDER BY created_at DESC`,
		doctorUUID, patientUUID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recs []entity.Recommendation
	for rows.Next() {
		var rec entity.Recommendation
		if err := rows.Scan(&rec.ID, &rec.DoctorUUID, &rec.PatientUUID, &rec.Text, &rec.CreatedAt); err != nil {
			return nil, err
		}
		recs = append(recs, rec)
	}
	return recs, nil
}
