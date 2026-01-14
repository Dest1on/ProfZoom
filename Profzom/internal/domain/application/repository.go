package application

import (
	"context"

	"github.com/Dest1on/ProfZoom-backend/internal/common"
)

type Repository interface {
	Create(ctx context.Context, application Application) (*Application, error)
	GetByID(ctx context.Context, id common.UUID) (*Application, error)
	ListByVacancy(ctx context.Context, vacancyID common.UUID) ([]Application, error)
	ListByStudent(ctx context.Context, studentID common.UUID) ([]Application, error)
	ListByCompany(ctx context.Context, companyID common.UUID) ([]Application, error)
	UpdateStatus(ctx context.Context, id common.UUID, status Status, feedback string) (*Application, error)
	FindByVacancyAndStudent(ctx context.Context, vacancyID, studentID common.UUID) (*Application, error)
}
