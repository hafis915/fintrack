package budget

import (
	"testing"

	"github.com/google/uuid"
)

// defaultFlexCats returns the standard flexible category set (variable + want)
// with zero amounts — the shape the handler passes to SuggestFlexible.
func defaultFlexCats() []IntakeItem {
	return []IntakeItem{
		{CategoryID: uuid.New(), Name: "Makan & minum", Icon: "🍱", Type: ExpenseVariable},
		{CategoryID: uuid.New(), Name: "Transportasi", Icon: "🚌", Type: ExpenseVariable},
		{CategoryID: uuid.New(), Name: "Belanja", Icon: "🛒", Type: ExpenseVariable},
		{CategoryID: uuid.New(), Name: "Kesehatan", Icon: "💊", Type: ExpenseVariable},
		{CategoryID: uuid.New(), Name: "Hiburan", Icon: "🎬", Type: ExpenseWant},
		{CategoryID: uuid.New(), Name: "Nongkrong", Icon: "☕", Type: ExpenseWant},
		{CategoryID: uuid.New(), Name: "Self-care", Icon: "💅", Type: ExpenseWant},
	}
}

func sumFlexible(items []PlanItem) int64 {
	var s int64
	for _, it := range items {
		s += it.AllocatedAmount
	}
	return s
}

func TestSuggestFlexible(t *testing.T) {
	tests := []struct {
		name          string
		answers       IntakeAnswers
		fixed         []IntakeItem
		wantProgram   Program
		wantSavingsGT int64 // savings target must be strictly greater than this
		wantWarning   bool
	}{
		{
			name: "normal seimbang plan splits discretionary",
			answers: IntakeAnswers{
				Income: 8_000_000, HousingType: HousingKosan, Goal: GoalBalance,
				DebtTypes: []DebtType{DebtNone}, EmergencyMonths: 3, LifestyleStyle: LifestyleBalanced,
			},
			fixed: []IntakeItem{
				{CategoryID: uuid.New(), Name: "Sewa kosan", Type: ExpenseFixed, Amount: 1_500_000},
				{CategoryID: uuid.New(), Name: "Listrik", Type: ExpenseFixed, Amount: 300_000},
			},
			wantProgram:   ProgramSeimbang,
			wantSavingsGT: 0,
		},
		{
			name: "over-budget yields warning and no flexible",
			answers: IntakeAnswers{
				Income: 4_000_000, HousingType: HousingKpr, Goal: GoalInvest,
				DebtTypes: []DebtType{DebtNone}, EmergencyMonths: 6, LifestyleStyle: LifestyleStrict,
			},
			fixed: []IntakeItem{
				{CategoryID: uuid.New(), Name: "Cicilan KPR", Type: ExpenseFixed, Amount: 3_800_000},
			},
			wantProgram: ProgramTumbuh,
			wantWarning: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			flex := defaultFlexCats()
			sug, err := SuggestFlexible(tc.answers, tc.fixed, flex)
			if err != nil {
				t.Fatalf("SuggestFlexible: %v", err)
			}

			if sug.Program != tc.wantProgram {
				t.Errorf("program: want %q, got %q", tc.wantProgram, sug.Program)
			}

			if tc.wantWarning {
				if sug.Warning == "" {
					t.Errorf("expected a warning, got none")
				}
				if len(sug.Flexible) != 0 {
					t.Errorf("over-budget: want empty flexible, got %d items", len(sug.Flexible))
				}
				// No negative splits — there are none, but the savings clamp
				// must be >= 0 and within income.
				if sug.SavingsTarget < 0 {
					t.Errorf("savings target negative: %d", sug.SavingsTarget)
				}
				return
			}

			if sug.Warning != "" {
				t.Errorf("unexpected warning: %q", sug.Warning)
			}
			if sug.SavingsTarget <= tc.wantSavingsGT {
				t.Errorf("savings target: want > %d, got %d", tc.wantSavingsGT, sug.SavingsTarget)
			}

			// Discretionary split must sum exactly to the discretionary pool.
			gotSum := sumFlexible(sug.Flexible)
			if gotSum != sug.Discretionary {
				t.Errorf("flexible sum %d != discretionary %d", gotSum, sug.Discretionary)
			}

			// Plan must stay within income: fixed + flexible + savings <= income.
			total := sug.FixedTotal + gotSum + sug.SavingsTarget
			if total > tc.answers.Income {
				t.Errorf("plan exceeds income: total %d > income %d", total, tc.answers.Income)
			}

			// No negative splits anywhere.
			for _, it := range sug.Flexible {
				if it.AllocatedAmount < 0 {
					t.Errorf("category %q negative: %d", it.CategoryName, it.AllocatedAmount)
				}
			}
		})
	}
}

// Strict lifestyle should squeeze wants harder than easy lifestyle, given the
// same income and fixed expenses.
func TestSuggestFlexible_StrictSqueezesWants(t *testing.T) {
	base := IntakeAnswers{
		Income: 8_000_000, HousingType: HousingKosan, Goal: GoalBalance,
		DebtTypes: []DebtType{DebtNone}, EmergencyMonths: 3,
	}
	fixed := []IntakeItem{
		{CategoryID: uuid.New(), Name: "Sewa kosan", Type: ExpenseFixed, Amount: 1_500_000},
	}

	wantsTotal := func(lifestyle LifestyleStyle) int64 {
		a := base
		a.LifestyleStyle = lifestyle
		sug, err := SuggestFlexible(a, fixed, defaultFlexCats())
		if err != nil {
			t.Fatalf("SuggestFlexible(%s): %v", lifestyle, err)
		}
		var w int64
		for _, it := range sug.Flexible {
			if it.Type == ExpenseWant {
				w += it.AllocatedAmount
			}
		}
		return w
	}

	strictWants := wantsTotal(LifestyleStrict)
	easyWants := wantsTotal(LifestyleEasy)

	if strictWants >= easyWants {
		t.Errorf("strict wants (%d) should be < easy wants (%d)", strictWants, easyWants)
	}
}

func TestRebalance(t *testing.T) {
	makan := uuid.New()
	transport := uuid.New()
	hiburan := uuid.New()
	nongkrong := uuid.New()

	const income int64 = 8_000_000
	const savings int64 = 2_000_000

	// flexible currently sums to 3_000_000; fixed implied = income - savings -
	// flexible = 3_000_000 (not modeled directly here).
	flexible := []FlexItem{
		{CategoryID: makan, Name: "Makan & minum", Type: ExpenseVariable, Amount: 1_200_000},
		{CategoryID: transport, Name: "Transportasi", Type: ExpenseVariable, Amount: 600_000},
		{CategoryID: hiburan, Name: "Hiburan", Type: ExpenseWant, Amount: 700_000},
		{CategoryID: nongkrong, Name: "Nongkrong", Type: ExpenseWant, Amount: 500_000},
	}

	before := int64(0)
	for _, f := range flexible {
		before += f.Amount
	}

	// Raise makan by 400k. Wants (hiburan+nongkrong = 1.2M) should absorb it
	// first; savings must stay untouched.
	updated, newSavings, err := Rebalance(income, savings, flexible, makan, 1_600_000, false)
	if err != nil {
		t.Fatalf("Rebalance: %v", err)
	}

	got := map[uuid.UUID]int64{}
	var after int64
	for _, f := range updated {
		got[f.CategoryID] = f.Amount
		after += f.Amount
	}

	if got[makan] != 1_600_000 {
		t.Errorf("makan: want 1_600_000, got %d", got[makan])
	}
	// Savings untouched.
	if newSavings != savings {
		t.Errorf("savings: want untouched %d, got %d", savings, newSavings)
	}
	// Wants dropped by the full 400k delta.
	wantsAfter := got[hiburan] + got[nongkrong]
	if wantsAfter != 1_200_000-400_000 {
		t.Errorf("wants after: want %d, got %d", 1_200_000-400_000, wantsAfter)
	}
	// Variable transport untouched (wants could absorb the whole delta).
	if got[transport] != 600_000 {
		t.Errorf("transport: want untouched 600_000, got %d", got[transport])
	}
	// Flexible total conserved (delta fully absorbed within flexible).
	if after != before {
		t.Errorf("flexible total changed: before %d, after %d", before, after)
	}
	// Plan stays within income.
	if after+newSavings > income {
		t.Errorf("plan exceeds income: %d > %d", after+newSavings, income)
	}
	// No negatives.
	for _, f := range updated {
		if f.Amount < 0 {
			t.Errorf("category %q negative: %d", f.Name, f.Amount)
		}
	}
}

// When wants can't cover the raise and the user permits it, savings absorbs the
// remainder.
func TestRebalance_TouchesSavingsWhenAllowed(t *testing.T) {
	makan := uuid.New()
	hiburan := uuid.New()

	const income int64 = 8_000_000
	const savings int64 = 2_000_000

	flexible := []FlexItem{
		{CategoryID: makan, Name: "Makan & minum", Type: ExpenseVariable, Amount: 1_000_000},
		{CategoryID: hiburan, Name: "Hiburan", Type: ExpenseWant, Amount: 300_000},
	}

	// Raise makan by 1M. Wants only hold 300k → 700k must come from savings.
	updated, newSavings, err := Rebalance(income, savings, flexible, makan, 2_000_000, true)
	if err != nil {
		t.Fatalf("Rebalance: %v", err)
	}

	got := map[uuid.UUID]int64{}
	for _, f := range updated {
		got[f.CategoryID] = f.Amount
	}
	if got[makan] != 2_000_000 {
		t.Errorf("makan: want 2_000_000, got %d", got[makan])
	}
	if got[hiburan] != 0 {
		t.Errorf("hiburan: want drained to 0, got %d", got[hiburan])
	}
	if newSavings != savings-700_000 {
		t.Errorf("savings: want %d, got %d", savings-700_000, newSavings)
	}
}

// Regression: apportion must conserve (sum == total) and never go negative.
// The old splitDiscretionary rounded each share to a step then dumped the whole
// drift on one category and clamped at 0 — so a small pool spread over many
// same-weight categories that each rounded UP produced parts summing to MORE
// than the pool (money created, an over-income suggested budget). The repro is
// the first case below: 40_000 across 7 equal weights at a 10_000 step.
func TestApportionConservesAndNonNegative(t *testing.T) {
	cases := []struct {
		name    string
		total   int64
		weights []float64
		step    int64
	}{
		{"tiny pool, 7 equal weights (the bug repro)", 40_000, []float64{1, 1, 1, 1, 1, 1, 1}, 10_000},
		{"uneven weights", 1_000_000, []float64{2, 1.5, 1, 0.8}, 10_000},
		{"total not a multiple of step", 43_333, []float64{1, 1, 1}, 10_000},
		{"total smaller than one step", 5_000, []float64{1, 1, 1, 1, 1}, 10_000},
		{"zero total", 0, []float64{1, 1}, 10_000},
		{"all-zero weights fall back to even split", 100_000, []float64{0, 0, 0}, 10_000},
		{"single category", 37_000, []float64{1}, 10_000},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := apportion(tc.total, tc.weights, tc.step)
			if len(got) != len(tc.weights) {
				t.Fatalf("len: want %d, got %d", len(tc.weights), len(got))
			}
			var sum int64
			for i, v := range got {
				if v < 0 {
					t.Errorf("part[%d] negative: %d", i, v)
				}
				sum += v
			}
			if sum != tc.total {
				t.Errorf("sum %d != total %d (money created/lost)", sum, tc.total)
			}
		})
	}
}

// End-to-end guard for the same bug: an income barely above fixed expenses
// yields a tiny discretionary pool. The suggested flexible amounts must still
// sum exactly to that pool and keep the whole plan within income.
func TestSuggestFlexible_TinyDiscretionaryStaysWithinIncome(t *testing.T) {
	a := IntakeAnswers{
		Income: 2_300_000, HousingType: HousingKosan, Goal: GoalBalance,
		DebtTypes: []DebtType{DebtNone}, EmergencyMonths: 1, LifestyleStyle: LifestyleBalanced,
	}
	fixed := []IntakeItem{
		{CategoryID: uuid.New(), Name: "Sewa kosan", Type: ExpenseFixed, Amount: 1_700_000},
	}
	sug, err := SuggestFlexible(a, fixed, defaultFlexCats())
	if err != nil {
		t.Fatalf("SuggestFlexible: %v", err)
	}
	gotSum := sumFlexible(sug.Flexible)
	if gotSum != sug.Discretionary {
		t.Errorf("flexible sum %d != discretionary %d", gotSum, sug.Discretionary)
	}
	if total := sug.FixedTotal + gotSum + sug.SavingsTarget; total > a.Income {
		t.Errorf("plan exceeds income: total %d > income %d", total, a.Income)
	}
	for _, it := range sug.Flexible {
		if it.AllocatedAmount < 0 {
			t.Errorf("category %q negative: %d", it.CategoryName, it.AllocatedAmount)
		}
	}
}
