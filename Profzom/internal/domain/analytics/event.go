package analytics

import (
	"time"

	"github.com/Dest1on/ProfZoom-backend/internal/common"
)

type Event struct {
	ID        common.UUID
	UserID    *common.UUID
	Name      string
	Payload   []byte
	CreatedAt time.Time
}
