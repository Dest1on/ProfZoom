package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Dest1on/ProfZoom-backend/internal/common"
	"github.com/Dest1on/ProfZoom-backend/internal/domain/user"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByPhone(ctx context.Context, phone string) (*user.User, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, phone, created_at, updated_at FROM users WHERE phone = $1`, phone)
	var u user.User
	if err := row.Scan(&u.ID, &u.Phone, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common.NewError(common.CodeNotFound, "user not found", err)
		}
		return nil, common.NewError(common.CodeInternal, "failed to load user", err)
	}
	roles, err := r.ListRoles(ctx, u.ID)
	if err != nil {
		return nil, err
	}
	u.Roles = roles
	return &u, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id common.UUID) (*user.User, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, phone, created_at, updated_at FROM users WHERE id = $1`, id)
	var u user.User
	if err := row.Scan(&u.ID, &u.Phone, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common.NewError(common.CodeNotFound, "user not found", err)
		}
		return nil, common.NewError(common.CodeInternal, "failed to load user", err)
	}
	roles, err := r.ListRoles(ctx, u.ID)
	if err != nil {
		return nil, err
	}
	u.Roles = roles
	return &u, nil
}

func (r *UserRepository) Create(ctx context.Context, phone string) (*user.User, error) {
	id := common.NewUUID()
	now := time.Now().UTC()
	_, err := r.db.ExecContext(ctx, `INSERT INTO users (id, phone, created_at, updated_at) VALUES ($1, $2, $3, $4)`, id, phone, now, now)
	if err != nil {
		return nil, common.NewError(common.CodeInternal, "failed to create user", err)
	}
	return &user.User{ID: id, Phone: phone, CreatedAt: now, UpdatedAt: now}, nil
}

func (r *UserRepository) SetRoles(ctx context.Context, userID common.UUID, roles []user.Role) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM user_roles WHERE user_id = $1`, userID)
	if err != nil {
		return common.NewError(common.CodeInternal, "failed to reset roles", err)
	}
	for _, role := range roles {
		_, err := r.db.ExecContext(ctx, `INSERT INTO user_roles (user_id, role) VALUES ($1, $2)`, userID, role)
		if err != nil {
			return common.NewError(common.CodeInternal, "failed to set role", err)
		}
	}
	return nil
}

func (r *UserRepository) ListRoles(ctx context.Context, userID common.UUID) ([]user.Role, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT role FROM user_roles WHERE user_id = $1`, userID)
	if err != nil {
		return nil, common.NewError(common.CodeInternal, "failed to list roles", err)
	}
	defer rows.Close()
	var roles []user.Role
	for rows.Next() {
		var role user.Role
		if err := rows.Scan(&role); err != nil {
			return nil, common.NewError(common.CodeInternal, "failed to scan role", err)
		}
		roles = append(roles, role)
	}
	return roles, nil
}
