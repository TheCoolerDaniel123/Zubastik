package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	pgclient "shared/clients/postgres"
	"shared/entity"
	"user-service/api/dto"
	"user-service/internal/storage/database"
)

type ConversationRepo struct {
	cli *pgclient.PostgresClient
}

func NewConversationRepo(cli *pgclient.PostgresClient) *ConversationRepo {
	return &ConversationRepo{cli: cli}
}

func (r *ConversationRepo) GetOrCreate(ctx context.Context, doctorUUID, patientUUID uuid.UUID) (*entity.Conversation, error) {
	var conv entity.Conversation
	err := r.cli.DB.QueryRowContext(ctx,
		`INSERT INTO conversations (doctor_uuid, patient_uuid)
		 VALUES ($1, $2)
		 ON CONFLICT (doctor_uuid, patient_uuid) DO UPDATE SET created_at = conversations.created_at
		 RETURNING id, doctor_uuid, patient_uuid, created_at`,
		doctorUUID, patientUUID,
	).Scan(&conv.ID, &conv.DoctorUUID, &conv.PatientUUID, &conv.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

func (r *ConversationRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Conversation, error) {
	var conv entity.Conversation
	err := r.cli.DB.QueryRowContext(ctx,
		`SELECT id, doctor_uuid, patient_uuid, created_at FROM conversations WHERE id = $1`,
		id,
	).Scan(&conv.ID, &conv.DoctorUUID, &conv.PatientUUID, &conv.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, database.ErrConversationNotFound
	}
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

func (r *ConversationRepo) GetByDoctor(ctx context.Context, doctorUUID uuid.UUID) ([]entity.Conversation, error) {
	rows, err := r.cli.DB.QueryContext(ctx,
		`SELECT id, doctor_uuid, patient_uuid, created_at FROM conversations
		 WHERE doctor_uuid = $1 ORDER BY created_at DESC`,
		doctorUUID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanConversations(rows)
}

func (r *ConversationRepo) GetByPatientWithDetails(ctx context.Context, patientUUID uuid.UUID) ([]dto.ConversationResponse, error) {
	rows, err := r.cli.DB.QueryContext(ctx,
		`SELECT
		    c.id,
		    c.doctor_uuid,
		    COALESCE(u.name, '') AS doctor_name,
		    COALESCE(
		        (SELECT dm.content FROM direct_messages dm
		         WHERE dm.conversation_id = c.id
		         ORDER BY dm.created_at DESC LIMIT 1),
		        ''
		    ) AS last_message,
		    COALESCE(
		        (SELECT dm.created_at::text FROM direct_messages dm
		         WHERE dm.conversation_id = c.id
		         ORDER BY dm.created_at DESC LIMIT 1),
		        c.created_at::text
		    ) AS last_message_time
		 FROM conversations c
		 LEFT JOIN users u ON u.uuid = c.doctor_uuid
		 WHERE c.patient_uuid = $1
		 ORDER BY last_message_time DESC`,
		patientUUID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []dto.ConversationResponse
	for rows.Next() {
		var cr dto.ConversationResponse
		if err := rows.Scan(&cr.ID, &cr.DoctorUUID, &cr.DoctorName, &cr.LastMessage, &cr.LastMessageTime); err != nil {
			return nil, err
		}
		result = append(result, cr)
	}
	return result, rows.Err()
}

func scanConversations(rows *sql.Rows) ([]entity.Conversation, error) {
	var convs []entity.Conversation
	for rows.Next() {
		var c entity.Conversation
		if err := rows.Scan(&c.ID, &c.DoctorUUID, &c.PatientUUID, &c.CreatedAt); err != nil {
			return nil, err
		}
		convs = append(convs, c)
	}
	return convs, rows.Err()
}
