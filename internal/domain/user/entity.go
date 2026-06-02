package user

import (
	"time"

	"github.com/google/uuid"
)

type Profile struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	IncomeEncrypted *string
	IncomeHint      *string
	HousingType     *string
	LifestyleStyle  *string
	EmergencyMonths int
	ActiveProgram   *string
	OnboardingDone  bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type CreateProfileInput struct {
	UserID          uuid.UUID
	IncomeEncrypted string
	IncomeHint      string
	HousingType     string
	LifestyleStyle  string
	EmergencyMonths int
	ActiveProgram   string
	OnboardingDone  bool
}
