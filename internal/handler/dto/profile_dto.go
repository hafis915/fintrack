package dto

type ProfileResponse struct {
	ID              string `json:"id"`
	IncomeHint      string `json:"income_hint,omitempty"`
	HousingType     string `json:"housing_type,omitempty"`
	LifestyleStyle  string `json:"lifestyle_style,omitempty"`
	EmergencyMonths int    `json:"emergency_months"`
	ActiveProgram   string `json:"active_program,omitempty"`
	OnboardingDone  bool   `json:"onboarding_done"`
}

type UpdateProfileRequest struct {
	LifestyleStyle  *string `json:"lifestyle_style"  validate:"omitempty,oneof=easy balanced strict"`
	EmergencyMonths *int    `json:"emergency_months" validate:"omitempty,oneof=0 1 3 6"`
}

type UpdateIncomeRequest struct {
	Income int64 `json:"income" validate:"required,gt=0"`
}
