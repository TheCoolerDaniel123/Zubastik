package postgres

import (
	"context"

	"github.com/google/uuid"
	pgclient "shared/clients/postgres"
	"shared/entity"
)

type DirectMessageRepo struct {
	cli *pgclient.PostgresClient
}

func NewDirectMessageRepo(cli *pgclient.PostgresClient) *DirectMessageRepo {
	return &DirectMessageRepo{cli: cli}
}

func (r *DirectMessageRepo) Create(ctx context.Context, msg entity.DirectMessage) (*entity.DirectMessage, error) {
	msg.ID = uuid.New()
	row := r.cli.DB.QueryRowContext(ctx,
		`INSERT INTO direct_messages (id, conversation_id, sender_uuid, content)
		 VALUES ($1, $2, $3, $4)
		 RETURNING created_at`,
		msg.ID, msg.ConversationID, msg.SenderUUID, msg.Content,
	)
	if err := row.Scan(&msg.CreatedAt); err != nil {
		return nil, err
	}
	return &msg, nil
}

func (r *DirectMessageRepo) GetByConversation(ctx context.Context, conversationID uuid.UUID) ([]entity.DirectMessage, error) {
	rows, err := r.cli.DB.QueryContext(ctx,
		`SELECT id, conversation_id, sender_uuid, content, created_at
		 FROM direct_messages
		 WHERE conversation_id = $1
		 ORDER BY created_at`,
		conversationID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []entity.DirectMessage
	for rows.Next() {
		var m entity.DirectMessage
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.SenderUUID, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}
	return msgs, rows.Err()
}
