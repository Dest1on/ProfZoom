package message

import (
	"context"

	"github.com/Dest1on/ProfZoom-backend/internal/common"
)

type Repository interface {
	Create(ctx context.Context, message Message) (*Message, error)
	ListByApplication(ctx context.Context, applicationID common.UUID, limit, offset int) ([]Message, error)
	LatestByApplication(ctx context.Context, applicationID common.UUID) (*Message, error)
}
