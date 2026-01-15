package app

import (
	"context"

	"profzom/internal/common"
	"profzom/internal/domain/analytics"
	"profzom/internal/domain/profile"
)

type ProfileService struct {
	students  profile.StudentRepository
	companies profile.CompanyRepository
	analytics analytics.Repository
}

func NewProfileService(students profile.StudentRepository, companies profile.CompanyRepository, analytics analytics.Repository) *ProfileService {
	return &ProfileService{students: students, companies: companies, analytics: analytics}
}

func (s *ProfileService) GetStudent(ctx context.Context, userID common.UUID) (*profile.StudentProfile, error) {
	return s.students.GetByUserID(ctx, userID)
}

func (s *ProfileService) UpsertStudent(ctx context.Context, profile profile.StudentProfile) (*profile.StudentProfile, error) {
	updated, err := s.students.Upsert(ctx, profile)
	if err != nil {
		return nil, err
	}
	_ = s.analytics.Create(ctx, analytics.Event{Name: "profile.student.updated", UserID: &profile.UserID, Payload: analyticsPayload(ctx, map[string]string{"user_id": profile.UserID.String()})})
	return updated, nil
}

func (s *ProfileService) GetCompany(ctx context.Context, userID common.UUID) (*profile.CompanyProfile, error) {
	return s.companies.GetByUserID(ctx, userID)
}

func (s *ProfileService) UpsertCompany(ctx context.Context, profile profile.CompanyProfile) (*profile.CompanyProfile, error) {
	updated, err := s.companies.Upsert(ctx, profile)
	if err != nil {
		return nil, err
	}
	_ = s.analytics.Create(ctx, analytics.Event{Name: "profile.company.updated", UserID: &profile.UserID, Payload: analyticsPayload(ctx, map[string]string{"user_id": profile.UserID.String()})})
	return updated, nil
}
