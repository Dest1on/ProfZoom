package analytics

import (
	"time"

	"profzom/internal/common"
)

type Event struct {
	ID        common.UUID
	UserID    *common.UUID
	Name      string
	Payload   []byte
	CreatedAt time.Time
}
