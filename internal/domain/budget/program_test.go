package budget

import "testing"

func TestSelectProgram(t *testing.T) {
	cases := []struct {
		name string
		in   IntakeAnswers
		want Program
	}{
		{
			name: "active CC debt overrides everything → bebas_utang",
			in: IntakeAnswers{
				Income: 8_000_000, HousingType: HousingKpr, Goal: GoalInvest,
				DebtTypes: []DebtType{DebtCC}, EmergencyMonths: 6, LifestyleStyle: LifestyleBalanced,
			},
			want: ProgramBebasUtang,
		},
		{
			name: "goal=debt with no listed debts still → bebas_utang",
			in: IntakeAnswers{
				Income: 8_000_000, HousingType: HousingKosan, Goal: GoalDebt,
				DebtTypes: []DebtType{DebtNone}, EmergencyMonths: 1, LifestyleStyle: LifestyleBalanced,
			},
			want: ProgramBebasUtang,
		},
		{
			name: "no emergency fund → pondasi",
			in: IntakeAnswers{
				Income: 8_000_000, HousingType: HousingKeluarga, Goal: GoalBalance,
				DebtTypes: []DebtType{DebtNone}, EmergencyMonths: 0, LifestyleStyle: LifestyleEasy,
			},
			want: ProgramPondasi,
		},
		{
			name: "goal=emergency forces pondasi even with some EF",
			in: IntakeAnswers{
				Income: 8_000_000, HousingType: HousingKosan, Goal: GoalEmergency,
				DebtTypes: []DebtType{DebtNone}, EmergencyMonths: 1, LifestyleStyle: LifestyleBalanced,
			},
			want: ProgramPondasi,
		},
		{
			name: "specific savings goal + safe EF → goal_chaser",
			in: IntakeAnswers{
				Income: 10_000_000, HousingType: HousingKpr, Goal: GoalGoal,
				DebtTypes: []DebtType{DebtNone}, EmergencyMonths: 3, LifestyleStyle: LifestyleStrict,
			},
			want: ProgramGoalChaser,
		},
		{
			name: "invest with healthy EF → tumbuh",
			in: IntakeAnswers{
				Income: 12_000_000, HousingType: HousingKpr, Goal: GoalInvest,
				DebtTypes: []DebtType{DebtNone}, EmergencyMonths: 6, LifestyleStyle: LifestyleBalanced,
			},
			want: ProgramTumbuh,
		},
		{
			name: "invest with insufficient EF falls back to seimbang",
			in: IntakeAnswers{
				Income: 10_000_000, HousingType: HousingKosan, Goal: GoalInvest,
				DebtTypes: []DebtType{DebtNone}, EmergencyMonths: 1, LifestyleStyle: LifestyleBalanced,
			},
			want: ProgramSeimbang,
		},
		{
			name: "general balance with EF in place → seimbang",
			in: IntakeAnswers{
				Income: 8_000_000, HousingType: HousingKosan, Goal: GoalBalance,
				DebtTypes: []DebtType{DebtNone}, EmergencyMonths: 3, LifestyleStyle: LifestyleBalanced,
			},
			want: ProgramSeimbang,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := SelectProgram(tc.in); got != tc.want {
				t.Errorf("SelectProgram: want %q, got %q", tc.want, got)
			}
		})
	}
}

func TestIntakeAnswers_Validate(t *testing.T) {
	base := IntakeAnswers{
		Income: 8_000_000, HousingType: HousingKpr, Goal: GoalBalance,
		DebtTypes: []DebtType{DebtNone}, EmergencyMonths: 1, LifestyleStyle: LifestyleBalanced,
	}
	if err := base.Validate(); err != nil {
		t.Fatalf("baseline answers should validate, got: %v", err)
	}

	mut := func(f func(*IntakeAnswers)) IntakeAnswers {
		a := base
		f(&a)
		return a
	}

	cases := map[string]IntakeAnswers{
		"zero_income":           mut(func(a *IntakeAnswers) { a.Income = 0 }),
		"negative_income":       mut(func(a *IntakeAnswers) { a.Income = -1 }),
		"insane_income":         mut(func(a *IntakeAnswers) { a.Income = 2_000_000_000 }),
		"bad_housing":           mut(func(a *IntakeAnswers) { a.HousingType = "villa" }),
		"bad_goal":              mut(func(a *IntakeAnswers) { a.Goal = "yolo" }),
		"bad_lifestyle":         mut(func(a *IntakeAnswers) { a.LifestyleStyle = "spartan" }),
		"odd_emergency_months":  mut(func(a *IntakeAnswers) { a.EmergencyMonths = 2 }),
		"bad_debt_type":         mut(func(a *IntakeAnswers) { a.DebtTypes = []DebtType{"mortgage"} }),
	}
	for name, in := range cases {
		t.Run(name, func(t *testing.T) {
			if err := in.Validate(); err == nil {
				t.Errorf("expected validation error for %s", name)
			}
		})
	}
}
