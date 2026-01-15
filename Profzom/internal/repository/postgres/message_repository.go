package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"profzom/internal/common"
	"profzom/internal/domain/message"
)

type MessageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) Create(ctx context.Context, msg message.Message) (*message.Message, error) {
	msg.ID = common.NewUUID()
	msg.CreatedAt = time.Now().UTC()
	_, err := r.db.ExecContext(ctx, `INSERT INTO messages (id, application_id, sender_id, body, created_at)
		VALUES ($1, $2, $3, $4, $5)`, msg.ID, msg.ApplicationID, msg.SenderID, msg.Body, msg.CreatedAt)
	if err != nil {
		return nil, common.NewError(common.CodeInternal, "failed to create message", err)
	}
	return &msg, nil
}

func (r *MessageRepository) ListByApplication(ctx context.Context, applicationID common.UUID, limit, offset int) ([]message.Message, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, application_id, sender_id, body, created_at FROM messages WHERE application_id = $1 ORDER BY created_at ASC LIMIT $2 OFFSET $3`, applicationID, limit, offset)
	if err != nil {
		return nil, common.NewError(common.CodeInternal, "failed to list messages", err)
	}
	defer rows.Close()
	var items []message.Message
	for rows.Next() {
		var msg message.Message
		if err := rows.Scan(&msg.ID, &msg.ApplicationID, &msg.SenderID, &msg.Body, &msg.CreatedAt); err != nil {
			return nil, common.NewError(common.CodeInternal, "failed to scan message", err)
		}
		items = append(items, msg)
	}
	return items, nil
}

func (r *MessageRepository) LatestByApplication(ctx context.Context, applicationID common.UUID) (*message.Message, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, application_id, sender_id, body, created_at FROM messages WHERE application_id = $1 ORDER BY created_at DESC LIMIT 1`, applicationID)
	var msg message.Message
	if err := row.Scan(&msg.ID, &msg.ApplicationID, &msg.SenderID, &msg.Body, &msg.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common.NewError(common.CodeNotFound, "message not found", err)
		}
		return nil, common.NewError(common.CodeInternal, "failed to load latest message", err)
	}
	return &msg, nil
}
