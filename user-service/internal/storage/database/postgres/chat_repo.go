package postgres

import (
	"context"

	"github.com/google/uuid"
	pgclient "shared/clients/postgres"
	"shared/entity"
)

type ChatRepo struct {
	cli *pgclient.PostgresClient
}

func NewChatRepo(cli *pgclient.PostgresClient) *ChatRepo {
	return &ChatRepo{cli: cli}
}

func (r *ChatRepo) SaveMessage(ctx context.Context, msg entity.ChatMessage) (*entity.ChatMessage, error) {
	msg.ID = uuid.New()
	row := r.cli.DB.QueryRowContext(ctx,
		`INSERT INTO chat_messages (id, user_uuid, role, content)
		 VALUES ($1, $2, $3, $4)
		 RETURNING created_at`,
		msg.ID, msg.UserUUID, msg.Role, msg.Content,
	)
	if err := row.Scan(&msg.CreatedAt); err != nil {
		return nil, err
	}
	return &msg, nil
}

func (r *ChatRepo) GetHistory(ctx context.Context, userUUID uuid.UUID) ([]entity.ChatMessage, error) {
	rows, err := r.cli.DB.QueryContext(ctx,
		`SELECT id, user_uuid, role, content, created_at
		 FROM chat_messages
		 WHERE user_uuid = $1
		 ORDER BY created_at`,
		userUUID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := make([]entity.ChatMessage, 0)
	for rows.Next() {
		var m entity.ChatMessage
		if err := rows.Scan(&m.ID, &m.UserUUID, &m.Role, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, rows.Err()
}
