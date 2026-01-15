package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"

	"profzom/internal/common"
	"profzom/internal/domain/profile"
)

type StudentProfileRepository struct {
	db *sql.DB
}

func NewStudentProfileRepository(db *sql.DB) *StudentProfileRepository {
	return &StudentProfileRepository{db: db}
}

func (r *StudentProfileRepository) GetByUserID(ctx context.Context, userID common.UUID) (*profile.StudentProfile, error) {
	row := r.db.QueryRowContext(ctx, `SELECT user_id, name, university, course, specialty, bio, skills, created_at, updated_at
		FROM student_profiles WHERE user_id = $1`, userID)
	var p profile.StudentProfile
	if err := row.Scan(&p.UserID, &p.Name, &p.University, &p.Course, &p.Specialty, &p.About, pq.Array(&p.Skills), &p.CreatedAt, &p.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common.NewError(common.CodeNotFound, "student profile not found", err)
		}
		return nil, common.NewError(common.CodeInternal, "failed to load student profile", err)
	}
	return &p, nil
}

func (r *StudentProfileRepository) Upsert(ctx context.Context, profile profile.StudentProfile) (*profile.StudentProfile, error) {
	now := time.Now().UTC()
	profile.UpdatedAt = now
	_, err := r.db.ExecContext(ctx, `INSERT INTO student_profiles (user_id, name, university, course, specialty, bio, skills, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (user_id) DO UPDATE SET name = EXCLUDED.name, university = EXCLUDED.university, course = EXCLUDED.course,
		specialty = EXCLUDED.specialty, bio = EXCLUDED.bio, skills = EXCLUDED.skills, updated_at = EXCLUDED.updated_at`,
		profile.UserID, profile.Name, profile.University, profile.Course, profile.Specialty, profile.About, pq.Array(profile.Skills), now, now)
	if err != nil {
		return nil, common.NewError(common.CodeInternal, "failed to upsert student profile", err)
	}
	profile.CreatedAt = now
	return &profile, nil
}

type CompanyProfileRepository struct {
	db *sql.DB
}

func NewCompanyProfileRepository(db *sql.DB) *CompanyProfileRepository {
	return &CompanyProfileRepository{db: db}
}

func (r *CompanyProfileRepository) GetByUserID(ctx context.Context, userID common.UUID) (*profile.CompanyProfile, error) {
	row := r.db.QueryRowContext(ctx, `SELECT user_id, name, industry, description, contact_name, contact_email, contact_phone, created_at, updated_at
		FROM company_profiles WHERE user_id = $1`, userID)
	var p profile.CompanyProfile
	if err := row.Scan(&p.UserID, &p.Name, &p.Industry, &p.Description, &p.ContactName, &p.ContactEmail, &p.ContactPhone, &p.CreatedAt, &p.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common.NewError(common.CodeNotFound, "company profile not found", err)
		}
		return nil, common.NewError(common.CodeInternal, "failed to load company profile", err)
	}
	return &p, nil
}

func (r *CompanyProfileRepository) Upsert(ctx context.Context, profile profile.CompanyProfile) (*profile.CompanyProfile, error) {
	now := time.Now().UTC()
	profile.UpdatedAt = now
	_, err := r.db.ExecContext(ctx, `INSERT INTO company_profiles (user_id, name, industry, description, contact_name, contact_email, contact_phone, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (user_id) DO UPDATE SET name = EXCLUDED.name, industry = EXCLUDED.industry, description = EXCLUDED.description,
		contact_name = EXCLUDED.contact_name, contact_email = EXCLUDED.contact_email, contact_phone = EXCLUDED.contact_phone, updated_at = EXCLUDED.updated_at`,
		profile.UserID, profile.Name, profile.Industry, profile.Description,
		profile.ContactName, profile.ContactEmail, profile.ContactPhone,
		now, now)
	if err != nil {
		return nil, common.NewError(common.CodeInternal, "failed to upsert company profile", err)
	}
	profile.CreatedAt = now
	return &profile, nil
}
