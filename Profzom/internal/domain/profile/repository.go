package profile

import (
	"context"

	"github.com/Dest1on/ProfZoom-backend/internal/common"
)

type StudentRepository interface {
	GetByUserID(ctx context.Context, userID common.UUID) (*StudentProfile, error)
	Upsert(ctx context.Context, profile StudentProfile) (*StudentProfile, error)
}

type CompanyRepository interface {
	GetByUserID(ctx context.Context, userID common.UUID) (*CompanyProfile, error)
	Upsert(ctx context.Context, profile CompanyProfile) (*CompanyProfile, error)
}
