package profile

import (
	"time"

	"profzom/internal/common"
)

type StudentProfile struct {
	UserID     common.UUID `json:"user_id"`
	Name       string      `json:"name"`
	University string      `json:"university"`
	Course     int         `json:"course"`
	Specialty  string      `json:"specialty"`
	Skills     []string    `json:"skills"`
	About      string      `json:"about"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}
