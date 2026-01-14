package message

import (
	"time"

	"github.com/Dest1on/ProfZoom-backend/internal/common"
)

type Message struct {
	ID            common.UUID `json:"id"`
	ApplicationID common.UUID `json:"application_id"`
	SenderID      common.UUID `json:"sender_id"`
	Body          string      `json:"body"`
	CreatedAt     time.Time   `json:"created_at"`
}
