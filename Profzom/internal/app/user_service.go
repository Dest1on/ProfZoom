package app

import (
    "context"
    "strings"

    "profzom/internal/common"
    "profzom/internal/domain/analytics"
    "profzom/internal/domain/user"
)

type UserService struct {
    users     user.Repository
    analytics analytics.Repository
}

func NewUserService(users user.Repository, analytics analytics.Repository) *UserService {
    return &UserService{users: users, analytics: analytics}
}

func (s *UserService) SetRole(ctx context.Context, userID common.UUID, role user.Role) error {
    normalized := user.Role(strings.ToLower(strings.TrimSpace(string(role))))
    if normalized != user.RoleStudent && normalized != user.RoleCompany {
        return common.NewValidationError("invalid role", map[string]string{"role": "role must be student or company"})
    }
    if _, err := s.users.GetByID(ctx, userID); err != nil {
        return err
    }
    roles, err := s.users.ListRoles(ctx, userID)
    if err != nil {
        return err
    }
    if len(roles) > 0 {
        return common.NewValidationError("role already selected", map[string]string{"role": "role already selected"})
    }
    if err := s.users.SetRoles(ctx, userID, []user.Role{normalized}); err != nil {
        return err
    }
    _ = s.analytics.Create(ctx, analytics.Event{Name: "user.role_selected", UserID: &userID, Payload: analyticsPayload(ctx, map[string]string{"role": string(normalized)})})
    return nil
}
