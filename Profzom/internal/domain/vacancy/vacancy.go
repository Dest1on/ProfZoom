package vacancy

import (
	"time"

	"github.com/Dest1on/ProfZoom-backend/internal/common"
)

type Status string

const (
	StatusDraft     Status = "draft"
	StatusPublished Status = "published"
)

type Vacancy struct {
	ID           common.UUID `json:"id"`
	CompanyID    common.UUID `json:"company_id"`
	Title        string      `json:"title"`
	Type         string      `json:"type"`
	Description  string      `json:"description"`
	Requirements []string    `json:"requirements"`
	Conditions   []string    `json:"conditions"`
	Salary       string      `json:"salary"`
	Location     string      `json:"location"`
	Status       Status      `json:"status"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}
