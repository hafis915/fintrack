package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/hafis915/fintrack/database/sqlc/generated"
	"github.com/hafis915/fintrack/pkg/apperr"
)

// UserProfile is the domain shape of a user_profiles row. Nullable enum
// columns from the DB become typed pointer fields so handlers can tell
// "not set yet" from "set to empty string".
type UserProfile struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	IncomeEncrypted  *string
	IncomeHint       *string
	HousingType      *string
	LifestyleStyle   *string
	EmergencyMonths  int16
	ActiveProgram    *string
	OnboardingDone   bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// UpsertOnboardingParams is the call shape used by the onboarding handler.
type UpsertOnboardingParams struct {
	UserID          uuid.UUID
	IncomeEncrypted string
	IncomeHint      string
	HousingType     string
	LifestyleStyle  string
	EmergencyMonths int16
	ActiveProgram   string
}

type UserProfilesRepo interface {
	Get(ctx context.Context, userID uuid.UUID) (UserProfile, error)
	UpsertOnboarding(ctx context.Context, p UpsertOnboardingParams) (UserProfile, error)
}

type userProfilesRepo struct {
	q *generated.Queries
}

func NewUserProfilesRepo(pool *pgxpool.Pool) UserProfilesRepo {
	return &userProfilesRepo{q: generated.New(pool)}
}

func (r *userProfilesRepo) Get(ctx context.Context, userID uuid.UUID) (UserProfile, error) {
	row, err := r.q.GetUserProfile(ctx, toPgUUID(userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return UserProfile{}, apperr.ErrNotFound
		}
		return UserProfile{}, fmt.Errorf("getting user profile: %w", err)
	}
	return toUserProfile(row), nil
}

func (r *userProfilesRepo) UpsertOnboarding(ctx context.Context, p UpsertOnboardingParams) (UserProfile, error) {
	housing := generated.HousingType(p.HousingType)
	style := generated.LifestyleStyle(p.LifestyleStyle)
	program := generated.FinancialProgram(p.ActiveProgram)

	row, err := r.q.UpsertUserProfileForOnboarding(ctx, generated.UpsertUserProfileForOnboardingParams{
		UserID:          toPgUUID(p.UserID),
		IncomeEncrypted: &p.IncomeEncrypted,
		IncomeHint:      &p.IncomeHint,
		HousingType:     &housing,
		LifestyleStyle:  &style,
		EmergencyMonths: p.EmergencyMonths,
		ActiveProgram:   &program,
	})
	if err != nil {
		return UserProfile{}, fmt.Errorf("upserting user profile: %w", err)
	}
	return toUserProfile(row), nil
}

func toUserProfile(row generated.UserProfile) UserProfile {
	p := UserProfile{
		ID:              fromPgUUID(row.ID),
		UserID:          fromPgUUID(row.UserID),
		IncomeEncrypted: row.IncomeEncrypted,
		IncomeHint:      row.IncomeHint,
		EmergencyMonths: row.EmergencyMonths,
		OnboardingDone:  row.OnboardingDone,
		CreatedAt:       fromPgTime(row.CreatedAt),
		UpdatedAt:       fromPgTime(row.UpdatedAt),
	}
	if row.HousingType != nil {
		s := string(*row.HousingType)
		p.HousingType = &s
	}
	if row.LifestyleStyle != nil {
		s := string(*row.LifestyleStyle)
		p.LifestyleStyle = &s
	}
	if row.ActiveProgram != nil {
		s := string(*row.ActiveProgram)
		p.ActiveProgram = &s
	}
	return p
}
