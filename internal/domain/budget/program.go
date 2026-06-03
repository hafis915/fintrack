// Package budget contains the goal-first onboarding engine: pure functions
// that turn the user's 6-question intake into a recommended Program and a
// concrete allocation plan. No I/O, no DB, no HTTP — these functions are the
// product's algorithmic core and are unit-tested in isolation.
package budget

import (
	"errors"
	"fmt"
)

// Program identifies one of the five MVP financial programs.
// Values match the financial_program enum in migration 0002.
type Program string

const (
	ProgramPondasi    Program = "pondasi"     // Pay Yourself First — no emergency fund yet
	ProgramBebasUtang Program = "bebas_utang" // Debt Snowball / Avalanche — active debt
	ProgramGoalChaser Program = "goal_chaser" // Zero-based budgeting — specific savings goal
	ProgramTumbuh     Program = "tumbuh"      // 50/30/20 + investing — emergency fund safe
	ProgramSeimbang   Program = "seimbang"    // 50/30/20 default — general control
)

// Goal is the user's primary financial goal from question 3 of onboarding.
// It overrides program selection where unambiguous (e.g. goal=debt → bebas_utang).
type Goal string

const (
	GoalEmergency Goal = "emergency"
	GoalDebt      Goal = "debt"
	GoalGoal      Goal = "goal"
	GoalInvest    Goal = "invest"
	GoalBalance   Goal = "balance"
)

// LifestyleStyle controls how aggressively we allocate to wants vs savings.
type LifestyleStyle string

const (
	LifestyleEasy     LifestyleStyle = "easy"
	LifestyleBalanced LifestyleStyle = "balanced"
	LifestyleStrict   LifestyleStyle = "strict"
)

// HousingType from onboarding question 2.
type HousingType string

const (
	HousingKosan    HousingType = "kosan"
	HousingKpr      HousingType = "kpr"
	HousingKeluarga HousingType = "keluarga"
)

// DebtType from onboarding question 4.
type DebtType string

const (
	DebtNone     DebtType = "none"
	DebtCC       DebtType = "cc"
	DebtPaylater DebtType = "paylater"
	DebtMulti    DebtType = "multi" // multiple kinds of debt
)

// IntakeAnswers is the raw input from POST /v1/onboarding — the six
// questions plus the expense_items list the user filled in.
type IntakeAnswers struct {
	Income          int64
	HousingType     HousingType
	Goal            Goal
	DebtTypes       []DebtType
	EmergencyMonths int // 0 | 1 | 3 | 6
	LifestyleStyle  LifestyleStyle
}

// Validate enforces the contract the PRD documents for /onboarding.
func (a IntakeAnswers) Validate() error {
	if a.Income <= 0 {
		return errors.New("income must be > 0")
	}
	if a.Income > 1_000_000_000 {
		// Sanity cap (1 milyar). Anyone with this much income isn't
		// fresh-worker target market — defer to manual support.
		return errors.New("income out of supported range")
	}
	switch a.HousingType {
	case HousingKosan, HousingKpr, HousingKeluarga:
	default:
		return fmt.Errorf("invalid housing_type: %q", a.HousingType)
	}
	switch a.Goal {
	case GoalEmergency, GoalDebt, GoalGoal, GoalInvest, GoalBalance:
	default:
		return fmt.Errorf("invalid goal: %q", a.Goal)
	}
	switch a.LifestyleStyle {
	case LifestyleEasy, LifestyleBalanced, LifestyleStrict:
	default:
		return fmt.Errorf("invalid lifestyle_style: %q", a.LifestyleStyle)
	}
	switch a.EmergencyMonths {
	case 0, 1, 3, 6:
	default:
		return fmt.Errorf("emergency_months must be 0|1|3|6, got %d", a.EmergencyMonths)
	}
	for _, d := range a.DebtTypes {
		switch d {
		case DebtNone, DebtCC, DebtPaylater, DebtMulti:
		default:
			return fmt.Errorf("invalid debt_type: %q", d)
		}
	}
	return nil
}

// SelectProgram is the deterministic program-selection algorithm. Order
// matters: debt trumps emergency fund (you can't safely save while
// compounding interest), and an empty emergency fund trumps a vague
// "balance" goal.
//
// The rules below are derived from the PRD's program table:
//   - Active debt              → Bebas Utang
//   - No emergency fund yet    → Pondasi
//   - Specific savings goal    → Goal Chaser
//   - Wants to invest          → Tumbuh
//   - Otherwise                → Seimbang
func SelectProgram(a IntakeAnswers) Program {
	hasActiveDebt := false
	for _, d := range a.DebtTypes {
		if d != DebtNone {
			hasActiveDebt = true
			break
		}
	}
	if hasActiveDebt || a.Goal == GoalDebt {
		return ProgramBebasUtang
	}
	if a.EmergencyMonths == 0 || a.Goal == GoalEmergency {
		return ProgramPondasi
	}
	if a.Goal == GoalGoal {
		return ProgramGoalChaser
	}
	if a.Goal == GoalInvest && a.EmergencyMonths >= 3 {
		return ProgramTumbuh
	}
	return ProgramSeimbang
}
