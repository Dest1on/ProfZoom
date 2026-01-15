package vacancy

import (
	"context"

	"profzom/internal/common"
)

type Repository interface {
	Create(ctx context.Context, vacancy Vacancy) (*Vacancy, error)
	Update(ctx context.Context, vacancy Vacancy) (*Vacancy, error)
	GetByID(ctx context.Context, id common.UUID) (*Vacancy, error)
	ListPublished(ctx context.Context, limit, offset int) ([]Vacancy, error)
	ListByCompany(ctx context.Context, companyID common.UUID) ([]Vacancy, error)
}
