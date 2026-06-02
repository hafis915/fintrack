package budget

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type OnboardingItem struct {
	Name       string
	Icon       string
	Type       string // fixed|variable|debt|want
	Amount     int64
	CategoryID uuid.UUID
}

type OnboardingInput struct {
	Income          int64
	Goal            string // emergency|debt|goal|invest|balance
	HousingType     string
	LifestyleStyle  string
	EmergencyMonths int
	DebtTypes       []string
	ExpenseItems    []OnboardingItem
}

type AllocationItem struct {
	CategoryID      uuid.UUID
	CategoryName    string
	Icon            string
	Type            string
	AllocatedAmount int64
	Percentage      float64
	IsDebtFocus     bool
}

type SummaryGroup struct {
	Amount     int64
	Percentage float64
}

type AllocationSummary struct {
	Kebutuhan SummaryGroup
	Utang     SummaryGroup
	Keinginan SummaryGroup
	Tabungan  SummaryGroup
	Total     int64
}

type Allocation struct {
	Program      string
	Items        []AllocationItem
	Summary      AllocationSummary
	HasDebtFocus bool
	Warning      string
}

func GenerateAllocation(in OnboardingInput) (*Allocation, error) {
	if in.Income <= 0 {
		return nil, errors.New("income must be > 0")
	}
	var totalExp int64
	for _, it := range in.ExpenseItems {
		if it.Amount < 0 {
			return nil, errors.New("expense amount must be >= 0")
		}
		totalExp += it.Amount
	}
	if totalExp > in.Income {
		return nil, errors.New("expenses exceed income")
	}

	prog := classifyProgram(in)
	out := &Allocation{Program: prog}
	out.Items = make([]AllocationItem, 0, len(in.ExpenseItems))

	var fixed, variable, debt, want int64
	for _, it := range in.ExpenseItems {
		isFocus := prog == "bebas_utang" && it.Type == "debt"
		out.Items = append(out.Items, AllocationItem{
			CategoryID:      it.CategoryID,
			CategoryName:    it.Name,
			Icon:            it.Icon,
			Type:            it.Type,
			AllocatedAmount: it.Amount,
			Percentage:      pct(it.Amount, in.Income),
			IsDebtFocus:     isFocus,
		})
		if isFocus {
			out.HasDebtFocus = true
		}
		switch it.Type {
		case "fixed":
			fixed += it.Amount
		case "variable":
			variable += it.Amount
		case "debt":
			debt += it.Amount
		case "want":
			want += it.Amount
		}
	}

	kebutuhan := fixed + variable
	tabungan := in.Income - kebutuhan - debt - want
	if tabungan < 0 {
		tabungan = 0
	}

	out.Summary = AllocationSummary{
		Kebutuhan: SummaryGroup{kebutuhan, pct(kebutuhan, in.Income)},
		Utang:     SummaryGroup{debt, pct(debt, in.Income)},
		Keinginan: SummaryGroup{want, pct(want, in.Income)},
		Tabungan:  SummaryGroup{tabungan, pct(tabungan, in.Income)},
		Total:     in.Income,
	}

	if pct(kebutuhan, in.Income) > 50 {
		out.Warning = fmt.Sprintf(
			"Kebutuhan pokok %.0f%% — sedikit di atas ideal (50%%), wajar untuk kondisimu.",
			pct(kebutuhan, in.Income))
	}
	return out, nil
}

func classifyProgram(in OnboardingInput) string {
	hasDebt := false
	for _, it := range in.ExpenseItems {
		if it.Type == "debt" && it.Amount > 0 {
			hasDebt = true
			break
		}
	}
	switch in.Goal {
	case "debt":
		if hasDebt {
			return "bebas_utang"
		}
		return "seimbang"
	case "emergency":
		return "pondasi"
	case "goal":
		return "goal_chaser"
	case "invest":
		if in.EmergencyMonths >= 3 {
			return "tumbuh"
		}
		return "pondasi"
	default:
		return "seimbang"
	}
}

func pct(part, total int64) float64 {
	if total == 0 {
		return 0
	}
	return float64(part) / float64(total) * 100
}
