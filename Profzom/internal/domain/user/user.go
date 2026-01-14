package user

import (
	"time"

	"github.com/Dest1on/ProfZoom-backend/internal/common"
)

type Role string

const (
	RoleStudent Role = "student"
	RoleCompany Role = "company"
)

type User struct {
	ID        common.UUID
	Phone     string
	Roles     []Role
	CreatedAt time.Time
	UpdatedAt time.Time
}
