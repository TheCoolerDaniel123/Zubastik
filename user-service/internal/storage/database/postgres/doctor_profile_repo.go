package postgres

import (
	"context"

	pgclient "shared/clients/postgres"
	"shared/entity"
)

type DoctorProfileRepo struct {
	cli *pgclient.PostgresClient
}

func NewDoctorProfileRepo(cli *pgclient.PostgresClient) *DoctorProfileRepo {
	return &DoctorProfileRepo{cli: cli}
}

func (r *DoctorProfileRepo) GetAll(ctx context.Context) ([]entity.DoctorProfile, error) {
	rows, err := r.cli.DB.QueryContext(ctx,
		`SELECT dp.user_uuid, COALESCE(u.name, ''), dp.specialization, dp.experience_years, dp.rating, dp.photo_url
		 FROM doctor_profiles dp
		 JOIN users u ON u.uuid = dp.user_uuid
		 ORDER BY dp.rating DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []entity.DoctorProfile
	for rows.Next() {
		var p entity.DoctorProfile
		if err := rows.Scan(&p.UserUUID, &p.Name, &p.Specialization, &p.ExperienceYears, &p.Rating, &p.PhotoURL); err != nil {
			return nil, err
		}
		profiles = append(profiles, p)
	}
	return profiles, rows.Err()
}

func (r *DoctorProfileRepo) Upsert(ctx context.Context, profile entity.DoctorProfile) error {
	_, err := r.cli.DB.ExecContext(ctx,
		`INSERT INTO doctor_profiles (user_uuid, specialization, experience_years, rating, photo_url)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (user_uuid) DO UPDATE SET
		     specialization   = EXCLUDED.specialization,
		     experience_years = EXCLUDED.experience_years,
		     rating           = EXCLUDED.rating,
		     photo_url        = EXCLUDED.photo_url`,
		profile.UserUUID, profile.Specialization, profile.ExperienceYears, profile.Rating, profile.PhotoURL,
	)
	return err
}
