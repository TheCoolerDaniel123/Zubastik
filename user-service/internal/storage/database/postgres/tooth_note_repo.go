package postgres

import (
	"context"

	"github.com/google/uuid"
	pgclient "shared/clients/postgres"
	"shared/entity"
	"user-service/internal/storage/database"
)

type ToothNoteRepo struct {
	cli *pgclient.PostgresClient
}

func NewToothNoteRepo(cli *pgclient.PostgresClient) *ToothNoteRepo {
	return &ToothNoteRepo{cli: cli}
}

func (r *ToothNoteRepo) CreateToothNote(ctx context.Context, note entity.ToothNote) (*entity.ToothNote, error) {
	note.ID = uuid.New()
	err := r.cli.DB.QueryRowContext(ctx,
		`INSERT INTO tooth_notes (id, doctor_uuid, patient_uuid, tooth_index, note)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, created_at`,
		note.ID, note.DoctorUUID, note.PatientUUID, note.ToothIndex, note.Note,
	).Scan(&note.ID, &note.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &note, nil
}

func (r *ToothNoteRepo) GetToothNotesByPatient(ctx context.Context, doctorUUID, patientUUID uuid.UUID) ([]entity.ToothNote, error) {
	rows, err := r.cli.DB.QueryContext(ctx,
		`SELECT id, doctor_uuid, patient_uuid, tooth_index, note, created_at
		 FROM tooth_notes
		 WHERE doctor_uuid=$1 AND patient_uuid=$2
		 ORDER BY created_at DESC`,
		doctorUUID, patientUUID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []entity.ToothNote
	for rows.Next() {
		var n entity.ToothNote
		if err := rows.Scan(&n.ID, &n.DoctorUUID, &n.PatientUUID, &n.ToothIndex, &n.Note, &n.CreatedAt); err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}
	return notes, nil
}

func (r *ToothNoteRepo) UpdateToothNote(ctx context.Context, noteID, doctorUUID uuid.UUID, newNote string) error {
	res, err := r.cli.DB.ExecContext(ctx,
		`UPDATE tooth_notes SET note=$1 WHERE id=$2 AND doctor_uuid=$3`,
		newNote, noteID, doctorUUID,
	)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return database.ErrNoteNotFound
	}
	return nil
}

func (r *ToothNoteRepo) DeleteToothNote(ctx context.Context, noteID, doctorUUID uuid.UUID) error {
	res, err := r.cli.DB.ExecContext(ctx,
		`DELETE FROM tooth_notes WHERE id=$1 AND doctor_uuid=$2`,
		noteID, doctorUUID,
	)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return database.ErrNoteNotFound
	}
	return nil
}
