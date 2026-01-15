package postgres

import (
	"context"
	"database/sql"
	"time"

	"profzom/internal/common"
	"profzom/internal/domain/analytics"
)

type AnalyticsRepository struct {
	db *sql.DB
}

func NewAnalyticsRepository(db *sql.DB) *AnalyticsRepository {
	return &AnalyticsRepository{db: db}
}

func (r *AnalyticsRepository) Create(ctx context.Context, event analytics.Event) error {
	if event.ID == "" {
		event.ID = common.NewUUID()
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now().UTC()
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO analytics_events (id, user_id, name, payload, created_at) VALUES ($1, $2, $3, $4, $5)`,
		event.ID, event.UserID, event.Name, event.Payload, event.CreatedAt)
	if err != nil {
		return common.NewError(common.CodeInternal, "failed to store analytics event", err)
	}
	return nil
}
