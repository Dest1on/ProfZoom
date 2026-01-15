package message

import (
	"context"

	"profzom/internal/common"
)

type Repository interface {
	Create(ctx context.Context, message Message) (*Message, error)
	ListByApplication(ctx context.Context, applicationID common.UUID, limit, offset int) ([]Message, error)
	LatestByApplication(ctx context.Context, applicationID common.UUID) (*Message, error)
}
