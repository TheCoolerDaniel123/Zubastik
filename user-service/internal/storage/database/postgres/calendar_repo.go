package postgres

import (
	"context"

	"github.com/google/uuid"
	pgclient "shared/clients/postgres"
	"shared/entity"
)

type CalendarRepo struct {
	cli *pgclient.PostgresClient
}

func NewCalendarRepo(cli *pgclient.PostgresClient) *CalendarRepo {
	return &CalendarRepo{cli: cli}
}

func (r *CalendarRepo) CreateCalendarEvent(ctx context.Context, event entity.CalendarEvent) (*entity.CalendarEvent, error) {
	err := r.cli.DB.QueryRowContext(ctx,
		`INSERT INTO calendar_events (user_uuid, date, time, description, status)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id`,
		event.UserUUID,
		event.Date,
		event.Time,
		event.Description,
		event.Status,
	).Scan(&event.ID)
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *CalendarRepo) GetCalendarEventsByUserUUID(ctx context.Context, userUUID uuid.UUID) ([]entity.CalendarEvent, error) {
	rows, err := r.cli.DB.QueryContext(ctx,
		`SELECT id, user_uuid, date::text, to_char(time, 'HH24:MI'), description, status
		 FROM calendar_events
		 WHERE user_uuid = $1
		 ORDER BY date, time`,
		userUUID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]entity.CalendarEvent, 0)
	for rows.Next() {
		var e entity.CalendarEvent
		if err := rows.Scan(&e.ID, &e.UserUUID, &e.Date, &e.Time, &e.Description, &e.Status); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}

func (r *CalendarRepo) DeleteCalendarEvent(ctx context.Context, eventID, userUUID uuid.UUID) error {
	_, err := r.cli.DB.ExecContext(ctx,
		`DELETE FROM calendar_events WHERE id = $1 AND user_uuid = $2`,
		eventID,
		userUUID,
	)
	return err
}
