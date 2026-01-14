package profile

import (
	"time"

	"github.com/Dest1on/ProfZoom-backend/internal/common"
)

type CompanyProfile struct {
	UserID       common.UUID `json:"user_id"`
	Name         string      `json:"name"`
	Industry     string      `json:"industry"`
	Description  string      `json:"description"`
	ContactName  string      `json:"contact_name"`
	ContactEmail string      `json:"contact_email"`
	ContactPhone string      `json:"contact_phone"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}
