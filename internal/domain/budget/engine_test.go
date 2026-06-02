package budget_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/hafis915/fintrack/internal/domain/budget"
)

func idA() uuid.UUID { return uuid.New() }
func idB() uuid.UUID { return uuid.New() }
func idC() uuid.UUID { return uuid.New() }
func idD() uuid.UUID { return uuid.New() }

func TestEngine_BebasUtang_AssignsDebtFocusFlag(t *testing.T) {
	in := budget.OnboardingInput{
		Income:          8_000_000,
		Goal:            "debt",
		HousingType:     "kpr",
		LifestyleStyle:  "balanced",
		EmergencyMonths: 1,
		DebtTypes:       []string{"cc"},
		ExpenseItems: []budget.OnboardingItem{
			{Name: "Sewa kosan", Type: "fixed", Amount: 1_200_000, CategoryID: idA()},
			{Name: "Cicilan KPR", Type: "fixed", Amount: 1_500_000, CategoryID: idB()},
			{Name: "Makan & minum", Type: "variable", Amount: 1_200_000, CategoryID: idC()},
			{Name: "Kartu kredit", Type: "debt", Amount: 400_000, CategoryID: idD()},
		},
	}
	out, err := budget.GenerateAllocation(in)
	require.NoError(t, err)
	require.Equal(t, "bebas_utang", out.Program)
	require.Equal(t, int64(8_000_000), out.Summary.Total)
	require.True(t, out.HasDebtFocus)
}

func TestEngine_Pondasi_NoDebt_NoEmergency(t *testing.T) {
	in := budget.OnboardingInput{
		Income: 6_000_000, Goal: "emergency", HousingType: "keluarga",
		LifestyleStyle: "balanced", EmergencyMonths: 0,
		ExpenseItems: []budget.OnboardingItem{
			{Name: "Makan", Type: "variable", Amount: 1_500_000, CategoryID: idA()},
		},
	}
	out, _ := budget.GenerateAllocation(in)
	require.Equal(t, "pondasi", out.Program)
}

func TestEngine_Tumbuh_EmergencyDone(t *testing.T) {
	in := budget.OnboardingInput{
		Income: 12_000_000, Goal: "invest", EmergencyMonths: 6, LifestyleStyle: "balanced",
		ExpenseItems: []budget.OnboardingItem{
			{Name: "Sewa", Type: "fixed", Amount: 2_500_000, CategoryID: idA()},
		},
	}
	out, _ := budget.GenerateAllocation(in)
	require.Equal(t, "tumbuh", out.Program)
}

func TestEngine_RejectsExpensesOverIncome(t *testing.T) {
	in := budget.OnboardingInput{
		Income: 5_000_000, Goal: "balance", LifestyleStyle: "balanced",
		ExpenseItems: []budget.OnboardingItem{
			{Name: "Sewa", Type: "fixed", Amount: 6_000_000, CategoryID: idA()},
		},
	}
	_, err := budget.GenerateAllocation(in)
	require.Error(t, err)
}
