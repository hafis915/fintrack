package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	db "github.com/hafis915/fintrack/database/sqlc/generated"
	"github.com/hafis915/fintrack/internal/domain/user"
	"github.com/hafis915/fintrack/pkg/apperror"
)

type userRepo struct {
	q *db.Queries
}

func NewUserRepo(pool *pgxpool.Pool) user.Repository {
	return &userRepo{q: db.New(pool)}
}

func (r *userRepo) GetByUserID(ctx context.Context, id uuid.UUID) (*user.Profile, error) {
	row, err := r.q.GetUserProfileByUserID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.NotFound("user_profile", id.String())
		}
		return nil, apperror.Internal(err)
	}
	return toUserDomain(row), nil
}

func (r *userRepo) Create(ctx context.Context, in user.CreateProfileInput) (*user.Profile, error) {
	row, err := r.q.CreateUserProfile(ctx, db.CreateUserProfileParams{
		UserID:          toPgUUID(in.UserID),
		IncomeEncrypted: optStr(in.IncomeEncrypted),
		IncomeHint:      optStr(in.IncomeHint),
		HousingType:     optHousing(in.HousingType),
		LifestyleStyle:  optLifestyle(in.LifestyleStyle),
		EmergencyMonths: int16(in.EmergencyMonths),
		ActiveProgram:   optProgram(in.ActiveProgram),
		OnboardingDone:  in.OnboardingDone,
	})
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return toUserDomain(row), nil
}

func (r *userRepo) UpdateLifestyle(ctx context.Context, id uuid.UUID, ls *string, em *int) (*user.Profile, error) {
	var emInt *int16
	if em != nil {
		v := int16(*em)
		emInt = &v
	}
	row, err := r.q.UpdateUserProfile(ctx, db.UpdateUserProfileParams{
		UserID:          toPgUUID(id),
		LifestyleStyle:  optLifestyle(strDeref(ls)),
		EmergencyMonths: emInt,
	})
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return toUserDomain(row), nil
}

func (r *userRepo) UpdateIncome(ctx context.Context, id uuid.UUID, encrypted, hint string) (string, error) {
	out, err := r.q.UpdateUserIncome(ctx, db.UpdateUserIncomeParams{
		UserID:          toPgUUID(id),
		IncomeEncrypted: &encrypted,
		IncomeHint:      &hint,
	})
	if err != nil {
		return "", apperror.Internal(err)
	}
	if out == nil {
		return hint, nil
	}
	return *out, nil
}

// --- conversion helpers (db <-> domain) ---

func toUserDomain(r db.UserProfile) *user.Profile {
	p := &user.Profile{
		ID:              fromPgUUID(r.ID),
		UserID:          fromPgUUID(r.UserID),
		IncomeEncrypted: r.IncomeEncrypted,
		IncomeHint:      r.IncomeHint,
		EmergencyMonths: int(r.EmergencyMonths),
		OnboardingDone:  r.OnboardingDone,
		CreatedAt:       r.CreatedAt.Time,
		UpdatedAt:       r.UpdatedAt.Time,
	}
	if r.HousingType != nil {
		s := string(*r.HousingType)
		p.HousingType = &s
	}
	if r.LifestyleStyle != nil {
		s := string(*r.LifestyleStyle)
		p.LifestyleStyle = &s
	}
	if r.ActiveProgram != nil {
		s := string(*r.ActiveProgram)
		p.ActiveProgram = &s
	}
	return p
}

func toPgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func fromPgUUID(p pgtype.UUID) uuid.UUID {
	if !p.Valid {
		return uuid.Nil
	}
	return p.Bytes
}

func optStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func strDeref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func optHousing(s string) *db.HousingType {
	if s == "" {
		return nil
	}
	v := db.HousingType(s)
	return &v
}

func optLifestyle(s string) *db.LifestyleStyle {
	if s == "" {
		return nil
	}
	v := db.LifestyleStyle(s)
	return &v
}

func optProgram(s string) *db.FinancialProgram {
	if s == "" {
		return nil
	}
	v := db.FinancialProgram(s)
	return &v
}
