package user

import (
	"context"

	"profzom/internal/common"
)

type Repository interface {
	FindByPhone(ctx context.Context, phone string) (*User, error)
	GetByID(ctx context.Context, id common.UUID) (*User, error)
	Create(ctx context.Context, phone string) (*User, error)
	SetRoles(ctx context.Context, userID common.UUID, roles []Role) error
	ListRoles(ctx context.Context, userID common.UUID) ([]Role, error)
}
