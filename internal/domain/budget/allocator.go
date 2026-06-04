package budget

import (
	"errors"
	"sort"

	"github.com/google/uuid"
)

// ErrIncomeTooLow is returned when declared expenses dwarf income to a degree
// that can only be a data-entry error (e.g. an income typo). Normal
// overspending — expenses a few times income — is still allowed; this guards
// only against absurd input that would otherwise overflow downstream numeric
// columns. Handlers map it to a clean 400.
var ErrIncomeTooLow = errors.New("income too low relative to declared expenses")

// maxExpenseToIncomeRatio is the ceiling on (total declared expenses / income)
// before we treat the input as degenerate. Overspending is a core, supported
// scenario ("gym for your money" coaches exactly these users), so the ceiling
// is deliberately high — it exists only to catch a missing/typo'd income
// (e.g. 10jt expenses vs 100 income = 100000x) before a per-category
// percentage would overflow the budget_items.percentage numeric(7,2) column
// (max 99999.99%). At 999x total, no single category can exceed ~99900%, so
// the insert is always safe. Normal and even heavy overspending (a few x
// income) flows straight through and gets a coaching Warning instead.
const maxExpenseToIncomeRatio = 999

// maxExpenseItemAmount caps a single declared expense (1 triliun Rupiah) so the
// int64 expense sum can't overflow and defeat the ratio guard above.
const maxExpenseItemAmount = 1_000_000_000_000

// ExpenseType maps to the expense_category_type enum.
type ExpenseType string

const (
	ExpenseFixed    ExpenseType = "fixed"
	ExpenseVariable ExpenseType = "variable"
	ExpenseDebt     ExpenseType = "debt"
	ExpenseWant     ExpenseType = "want"
)

// IntakeItem is one line the user entered during onboarding ("Sewa kosan,
// 1.200.000, fixed"). CategoryID must exist in expense_categories.
type IntakeItem struct {
	CategoryID uuid.UUID
	Name       string
	Icon       string
	Type       ExpenseType
	Amount     int64
}

// PlanItem is the allocated amount for one category in the final plan,
// with its percentage of total income pre-computed for the UI.
type PlanItem struct {
	CategoryID      uuid.UUID
	CategoryName    string
	Icon            string
	Type            ExpenseType
	AllocatedAmount int64
	Percentage      float64 // of total income, 2 d.p.
	IsDebtFocus     bool
}

// SummaryBucket is one of the four PRD-defined buckets shown on the
// onboarding result screen.
type SummaryBucket struct {
	Amount     int64
	Percentage float64
}

// Plan is the full output of the allocator — what the API returns to the
// client and what the repository persists into budget_plans + budget_items.
type Plan struct {
	Program     Program
	TotalIncome int64
	Summary     Summary
	Items       []PlanItem
	Warning     string // human-readable advisory ("Kebutuhan 56% — sedikit di atas ideal")
}

// Summary captures the four buckets from the PRD onboarding response.
type Summary struct {
	Kebutuhan SummaryBucket
	Utang     SummaryBucket
	Keinginan SummaryBucket
	Tabungan  SummaryBucket
}

// Allocate turns the user's intake into a Plan. Behaviour:
//   - User-entered amounts are respected (we don't override what the user
//     told us they spend on rent, etc.). The allocator's job is to compute
//     the *savings* and (for Bebas Utang) *debt focus* buckets that the
//     user didn't enter.
//   - Tabungan is whatever's left after kebutuhan + utang + keinginan.
//     If that goes negative, we surface a Warning rather than silently
//     producing a broken plan.
//   - Percentages are rounded to 2 d.p. for display consistency.
func Allocate(answers IntakeAnswers, items []IntakeItem) (*Plan, error) {
	if answers.Income <= 0 {
		return nil, errors.New("income must be > 0")
	}

	program := SelectProgram(answers)

	plan := &Plan{
		Program:     program,
		TotalIncome: answers.Income,
		Items:       make([]PlanItem, 0, len(items)),
	}

	var kebutuhan, utang, keinginan int64
	for _, it := range items {
		if it.Amount < 0 {
			return nil, errors.New("expense amount cannot be negative")
		}
		// Reject absurd per-item amounts BEFORE summing. Without this, a value
		// near math.MaxInt64 (or several large items) overflows the int64 sum
		// below, wrapping negative — which then slips past the
		// maxExpenseToIncomeRatio guard and overflows the numeric(7,2) column at
		// insert. 1 triliun is far above any real monthly expense.
		if it.Amount > maxExpenseItemAmount {
			return nil, ErrIncomeTooLow
		}
		pi := PlanItem{
			CategoryID:      it.CategoryID,
			CategoryName:    it.Name,
			Icon:            it.Icon,
			Type:            it.Type,
			AllocatedAmount: it.Amount,
			Percentage:      pct(it.Amount, answers.Income),
		}
		switch it.Type {
		case ExpenseFixed, ExpenseVariable:
			kebutuhan += it.Amount
		case ExpenseDebt:
			utang += it.Amount
			if program == ProgramBebasUtang {
				pi.IsDebtFocus = true
			}
		case ExpenseWant:
			keinginan += it.Amount
		default:
			return nil, errors.New("invalid expense type: " + string(it.Type))
		}
		plan.Items = append(plan.Items, pi)
	}

	// Degenerate-input guard: if declared expenses dwarf income beyond any
	// plausible overspend, the income is almost certainly a typo. Reject with
	// a clean domain error rather than letting a wildly out-of-range percentage
	// overflow the budget_items.percentage column at insert time.
	totalExpenses := kebutuhan + utang + keinginan
	if totalExpenses > answers.Income*maxExpenseToIncomeRatio {
		return nil, ErrIncomeTooLow
	}

	// Tabungan is the residual after the user's declared spending.
	tabungan := answers.Income - kebutuhan - utang - keinginan

	plan.Summary = Summary{
		Kebutuhan: SummaryBucket{Amount: kebutuhan, Percentage: pct(kebutuhan, answers.Income)},
		Utang:     SummaryBucket{Amount: utang, Percentage: pct(utang, answers.Income)},
		Keinginan: SummaryBucket{Amount: keinginan, Percentage: pct(keinginan, answers.Income)},
		Tabungan:  SummaryBucket{Amount: tabungan, Percentage: pct(tabungan, answers.Income)},
	}

	plan.Warning = buildWarning(plan.Summary, program)

	// Stable item ordering — fixed first, then variable, debt, want; by name within type.
	sort.SliceStable(plan.Items, func(i, j int) bool {
		a, b := plan.Items[i], plan.Items[j]
		if a.Type != b.Type {
			return typeRank(a.Type) < typeRank(b.Type)
		}
		return a.CategoryName < b.CategoryName
	})

	return plan, nil
}

func typeRank(t ExpenseType) int {
	switch t {
	case ExpenseFixed:
		return 0
	case ExpenseVariable:
		return 1
	case ExpenseDebt:
		return 2
	case ExpenseWant:
		return 3
	}
	return 99
}

func pct(part, total int64) float64 {
	if total == 0 {
		return 0
	}
	return roundTwo(float64(part) / float64(total) * 100)
}

func roundTwo(v float64) float64 {
	// Avoid math.Round import for one line — manual two-decimal rounding.
	if v >= 0 {
		return float64(int64(v*100+0.5)) / 100
	}
	return float64(int64(v*100-0.5)) / 100
}

// buildWarning produces the human-readable advisory shown on the result
// screen. Order matters: insolvent plan → warn first; high kebutuhan is
// a softer informational message.
func buildWarning(s Summary, program Program) string {
	if s.Tabungan.Amount < 0 {
		return "Pengeluaranmu masih di atas pemasukan. Mari kita pangkas Keinginan dulu — coba kurangi 20% di kategori ini bulan depan."
	}
	if s.Kebutuhan.Percentage > 50 && s.Kebutuhan.Percentage <= 60 {
		return "Kebutuhan pokok " + percentString(s.Kebutuhan.Percentage) + " — sedikit di atas ideal (50%), wajar untuk kondisimu."
	}
	if s.Kebutuhan.Percentage > 60 {
		return "Kebutuhan pokok " + percentString(s.Kebutuhan.Percentage) + " sudah > 60% income. Pertimbangkan cari sumber penghasilan tambahan."
	}
	if program == ProgramBebasUtang && s.Utang.Percentage < 5 {
		return "Alokasi utang masih rendah — kalau memungkinkan, tambah 5–10% dari Tabungan untuk percepat lunas."
	}
	return ""
}

func percentString(p float64) string {
	// Single digit precision is enough for advisory text.
	whole := int64(p + 0.5)
	return itoa(whole) + "%"
}

// Tiny stdlib-free int→string to keep this file dependency-light.
func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	buf := make([]byte, 0, 12)
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	if neg {
		buf = append([]byte{'-'}, buf...)
	}
	return string(buf)
}
