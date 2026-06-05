package budget

import (
	"errors"
	"sort"

	"github.com/google/uuid"
)

// This file is the deterministic core of the financial planner. The product
// rule (chosen deliberately): all money math is decided here, in pure Go. The
// OpenRouter LLM is *only* the language layer — it parses user intent and
// narrates trade-offs, but never invents budget numbers. Two entry points:
//
//   - SuggestFlexible: given the intake answers + the user's fixed expenses,
//     pick a program, derive a savings target as a fraction of income, then
//     split the leftover discretionary budget across the flexible categories.
//   - Rebalance: when the user nudges one flexible category up/down (via inline
//     edit or chat), absorb the delta from the *other* want/variable categories
//     first — protecting needs and the savings target — keeping the plan within
//     income.

// roundStep is the granularity we round suggested category amounts to. Budgets
// that read in clean Rp 10.000 increments feel intentional, not algorithmic.
const roundStep int64 = 10_000

// FlexItem is one flexible category in the working plan the user is refining.
// It carries the live Amount (which Rebalance mutates) plus the Type so the
// rebalancer can prefer pulling from wants before variable needs.
type FlexItem struct {
	CategoryID uuid.UUID
	Name       string
	Type       ExpenseType
	Amount     int64
}

// Suggestion is the output of SuggestFlexible — the program, the derived
// savings target, the bucket summary, and a suggested amount per flexible
// category. Warning is non-empty only when the fixed expenses + savings target
// already exceed income (the discretionary pool is negative).
type Suggestion struct {
	Program       Program
	SavingsTarget int64
	FixedTotal    int64
	Discretionary int64
	Summary       Summary
	Flexible      []PlanItem // suggested_amount lives in PlanItem.AllocatedAmount
	Warning       string
}

// savingsRate returns the base savings fraction of income for a program. These
// are the deterministic "house rules" of each program; lifestyle and emergency
// posture nudge them in SuggestFlexible.
//
//   - Tumbuh:      emergency fund is already safe → invest aggressively.
//   - Goal Chaser: goal-driven, save hard toward the target.
//   - Seimbang:    classic 50/30/20 → 20% savings.
//   - Pondasi:     building the first emergency fund → moderate, steady.
//   - Bebas Utang: debt comes first; the "savings" line stays low because the
//     residual (income − fixed − discretionary) is steered toward debt payoff,
//     not a savings pile.
func savingsRate(p Program) float64 {
	switch p {
	case ProgramTumbuh:
		return 0.28
	case ProgramGoalChaser:
		return 0.25
	case ProgramSeimbang:
		return 0.20
	case ProgramPondasi:
		return 0.18
	case ProgramBebasUtang:
		return 0.10
	default:
		return 0.20
	}
}

// SuggestFlexible derives a savings target and a deterministic split of the
// remaining discretionary budget across the supplied flexible categories.
//
// flexibleCats is the user's flexible category set (type variable or want) with
// zero/ignored amounts — the handler passes the catalog rows and we fill in the
// suggested amounts. Their CategoryID/Name/Icon/Type are carried straight
// through to the returned PlanItems.
//
// If fixedTotal + savingsTarget already meet or exceed income there is nothing
// left to split: we return the buckets we can compute, an empty Flexible slice,
// and a Warning — never negative splits.
func SuggestFlexible(answers IntakeAnswers, fixed []IntakeItem, flexibleCats []IntakeItem) (*Suggestion, error) {
	if answers.Income <= 0 {
		return nil, errors.New("income must be > 0")
	}

	program := SelectProgram(answers)

	// --- Fixed total, split into kebutuhan (fixed) + utang (debt). ---
	var kebutuhan, utang int64
	for _, it := range fixed {
		if it.Amount < 0 {
			return nil, errors.New("fixed expense amount cannot be negative")
		}
		if it.Amount > maxExpenseItemAmount {
			return nil, ErrIncomeTooLow
		}
		switch it.Type {
		case ExpenseFixed, ExpenseVariable:
			// A "fixed" intake item may be tagged variable (e.g. a utility that
			// varies) but is still non-negotiable — counts as kebutuhan.
			kebutuhan += it.Amount
		case ExpenseDebt:
			utang += it.Amount
		case ExpenseWant:
			// Defensive: a want shouldn't arrive in the fixed set, but if it
			// does, treat it as kebutuhan rather than dropping it from the math.
			kebutuhan += it.Amount
		default:
			return nil, errors.New("invalid expense type: " + string(it.Type))
		}
	}
	fixedTotal := kebutuhan + utang

	// --- Savings target: base program rate, nudged by lifestyle + emergency. ---
	rate := savingsRate(program)
	switch answers.LifestyleStyle {
	case LifestyleStrict:
		rate += 0.05 // disciplined → save more
	case LifestyleEasy:
		rate -= 0.05 // wants more breathing room → save less
	}
	// Still building the first months of runway → lean a little harder on
	// savings; already cushioned → can ease off.
	switch answers.EmergencyMonths {
	case 0:
		rate += 0.02
	case 6:
		rate -= 0.02
	}
	if rate < 0 {
		rate = 0
	}
	savingsTarget := roundTo(int64(float64(answers.Income)*rate), roundStep)
	if savingsTarget < 0 {
		savingsTarget = 0
	}

	discretionary := answers.Income - fixedTotal - savingsTarget

	sug := &Suggestion{
		Program:       program,
		SavingsTarget: savingsTarget,
		FixedTotal:    fixedTotal,
		Discretionary: discretionary,
		Flexible:      []PlanItem{},
	}

	// --- Over-budget: fixed + savings already swallow income. ---
	if discretionary < 0 {
		// Clamp the reported savings to what income could actually cover so the
		// summary doesn't claim a target the user can't fund, and surface a
		// coaching warning instead of negative splits.
		coverable := answers.Income - fixedTotal
		if coverable < 0 {
			coverable = 0
		}
		sug.SavingsTarget = coverable
		sug.Discretionary = 0
		sug.Summary = Summary{
			Kebutuhan: SummaryBucket{Amount: kebutuhan, Percentage: pct(kebutuhan, answers.Income)},
			Utang:     SummaryBucket{Amount: utang, Percentage: pct(utang, answers.Income)},
			Keinginan: SummaryBucket{Amount: 0, Percentage: 0},
			Tabungan:  SummaryBucket{Amount: coverable, Percentage: pct(coverable, answers.Income)},
		}
		sug.Warning = "Pengeluaran tetap + target tabungan sudah melebihi pemasukan — kurangi yang fixed atau target tabungan"
		return sug, nil
	}

	// --- Split discretionary across the flexible categories by weight. ---
	items := splitDiscretionary(discretionary, flexibleCats, answers)
	sug.Flexible = items

	var keinginan int64
	for _, it := range items {
		keinginan += it.AllocatedAmount
	}

	sug.Summary = Summary{
		Kebutuhan: SummaryBucket{Amount: kebutuhan, Percentage: pct(kebutuhan, answers.Income)},
		Utang:     SummaryBucket{Amount: utang, Percentage: pct(utang, answers.Income)},
		Keinginan: SummaryBucket{Amount: keinginan, Percentage: pct(keinginan, answers.Income)},
		Tabungan:  SummaryBucket{Amount: savingsTarget, Percentage: pct(savingsTarget, answers.Income)},
	}

	return sug, nil
}

// flexWeight assigns a relative share to a flexible category. Needs-ish
// variable categories (food biggest, then transport, daily shopping, health)
// get priority; pure wants (entertainment, hangouts, self-care) get the
// remainder and are squeezed harder under a strict lifestyle or aggressive
// savings posture. Matching is by category name substring so the handler can
// pass the live catalog without a brittle ID table here.
func flexWeight(name string, t ExpenseType, lifestyle LifestyleStyle) float64 {
	// wantScale squeezes pure-want categories when the user is disciplined.
	wantScale := 1.0
	if lifestyle == LifestyleStrict {
		wantScale = 0.6
	} else if lifestyle == LifestyleEasy {
		wantScale = 1.2
	}

	base := matchWeight(name, t)
	if t == ExpenseWant {
		return base * wantScale
	}
	return base
}

// matchWeight is the raw weight table by category. Keyed on lowercase name
// substrings of the seed categories described in the planner contract.
func matchWeight(name string, t ExpenseType) float64 {
	n := lower(name)
	switch {
	case contains(n, "makan"), contains(n, "minum"):
		return 4.0
	case contains(n, "transport"):
		return 2.0
	case contains(n, "belanja"):
		return 1.8
	case contains(n, "kesehatan"):
		return 1.5
	case contains(n, "hiburan"):
		return 1.2
	case contains(n, "nongkrong"):
		return 1.0
	case contains(n, "self"), contains(n, "care"):
		return 0.8
	}
	// Unknown category: weight by type so it still gets a sensible share.
	if t == ExpenseWant {
		return 1.0
	}
	return 1.5
}

// splitDiscretionary divides total across flexibleCats by weight, rounds each
// to roundStep, and reconciles rounding drift against the largest category so
// the parts sum back to total. Returns PlanItems with the suggested amount in
// AllocatedAmount and the percentage of income pre-computed.
func splitDiscretionary(total int64, flexibleCats []IntakeItem, answers IntakeAnswers) []PlanItem {
	out := make([]PlanItem, 0, len(flexibleCats))
	if total <= 0 || len(flexibleCats) == 0 {
		for _, c := range flexibleCats {
			out = append(out, PlanItem{
				CategoryID:      c.CategoryID,
				CategoryName:    c.Name,
				Icon:            c.Icon,
				Type:            c.Type,
				AllocatedAmount: 0,
				Percentage:      0,
			})
		}
		return out
	}

	var totalWeight float64
	weights := make([]float64, len(flexibleCats))
	for i, c := range flexibleCats {
		w := flexWeight(c.Name, c.Type, answers.LifestyleStyle)
		weights[i] = w
		totalWeight += w
	}

	amounts := apportion(total, weights, roundStep)

	for i, c := range flexibleCats {
		out = append(out, PlanItem{
			CategoryID:      c.CategoryID,
			CategoryName:    c.Name,
			Icon:            c.Icon,
			Type:            c.Type,
			AllocatedAmount: amounts[i],
			Percentage:      pct(amounts[i], answers.Income),
		})
	}
	return out
}

// apportion divides total across the given weights using the largest-remainder
// (Hamilton) method, in whole units of `step` where possible. It GUARANTEES two
// invariants that the old round-each-then-dump-drift-on-largest approach did not:
//
//   - Conservation: sum(result) == total exactly (no money created or lost).
//   - Non-negativity: every part is >= 0.
//
// The previous code rounded each share to the nearest step then added the whole
// drift to one category, clamping at 0 — so when many small same-weight shares
// each rounded UP, the negative drift overflowed that clamp and the parts summed
// to MORE than total (an over-income suggested budget). Hamilton avoids that by
// flooring to step then handing the leftover whole steps to the largest
// fractional remainders, with any sub-step remainder landing on the top one.
// Deterministic: ties break by index, so the result never depends on map order.
func apportion(total int64, weights []float64, step int64) []int64 {
	n := len(weights)
	out := make([]int64, n)
	if n == 0 || total <= 0 {
		return out
	}
	if step < 1 {
		step = 1
	}

	var totalWeight float64
	for _, w := range weights {
		if w > 0 {
			totalWeight += w
		}
	}

	// No usable weights → split as evenly as possible (still conserves exactly).
	if totalWeight <= 0 {
		base := total / int64(n)
		rem := total - base*int64(n)
		for i := range out {
			out[i] = base
			if int64(i) < rem {
				out[i]++
			}
		}
		return out
	}

	// Work in whole steps; the sub-step remainder (when total isn't a multiple of
	// step) is handed to the top-remainder category at the end.
	units := total / step
	subStep := total - units*step

	type frac struct {
		idx int
		rem float64
	}
	fracs := make([]frac, n)
	var assigned int64
	for i, w := range weights {
		ideal := 0.0
		if w > 0 {
			ideal = float64(units) * w / totalWeight
		}
		fl := int64(ideal) // floor; ideal >= 0
		out[i] = fl
		assigned += fl
		fracs[i] = frac{idx: i, rem: ideal - float64(fl)}
	}

	// Hand out the remaining whole steps (leftover < n) to the largest remainders.
	sort.SliceStable(fracs, func(a, b int) bool {
		if fracs[a].rem != fracs[b].rem {
			return fracs[a].rem > fracs[b].rem
		}
		return fracs[a].idx < fracs[b].idx
	})
	for k := int64(0); k < units-assigned; k++ {
		out[fracs[k].idx]++
	}

	for i := range out {
		out[i] *= step
	}
	if subStep > 0 {
		out[fracs[0].idx] += subStep
	}
	return out
}

// Rebalance sets the target category to targetAmount and absorbs the resulting
// delta from the *other* flexible categories. The absorption order protects the
// plan's priorities:
//
//  1. Pull from / add to other 'want' categories first (proportionally).
//  2. If wants can't absorb it, fall to 'variable' (needs-ish) categories.
//  3. Only touch the savings target if allowTouchSavings is true (the user
//     explicitly said "ambil dari tabungan").
//
// No category goes below 0, and sum(flexible)+savings is kept ≤ income. Returns
// the updated flexible slice and the (possibly reduced) savings target.
func Rebalance(
	income, savingsTarget int64,
	flexible []FlexItem,
	targetCategoryID uuid.UUID,
	targetAmount int64,
	allowTouchSavings bool,
) ([]FlexItem, int64, error) {
	if income <= 0 {
		return nil, 0, errors.New("income must be > 0")
	}
	if targetAmount < 0 {
		return nil, 0, errors.New("target amount cannot be negative")
	}

	// Copy so we never mutate the caller's slice in place.
	updated := make([]FlexItem, len(flexible))
	copy(updated, flexible)

	targetIdx := -1
	for i := range updated {
		if updated[i].CategoryID == targetCategoryID {
			targetIdx = i
			break
		}
	}
	if targetIdx == -1 {
		return nil, 0, errors.New("target category not found in flexible set")
	}

	// delta > 0 means the target grew and we must FIND that much elsewhere.
	// delta < 0 means the target shrank and we REDISTRIBUTE the freed budget
	// back to the other categories (so the plan stays fully allocated).
	delta := targetAmount - updated[targetIdx].Amount
	updated[targetIdx].Amount = targetAmount

	if delta == 0 {
		return updated, savingsTarget, nil
	}

	remaining := delta

	if delta > 0 {
		// Need to claw back `remaining` from others: wants first, then variable.
		remaining = pull(updated, targetIdx, ExpenseWant, remaining)
		if remaining > 0 {
			remaining = pull(updated, targetIdx, ExpenseVariable, remaining)
		}
		// Anything still unfunded comes from savings, but only with permission.
		if remaining > 0 && allowTouchSavings {
			take := remaining
			if take > savingsTarget {
				take = savingsTarget
			}
			savingsTarget -= take
			remaining -= take
		}
		// If `remaining` is still > 0 here, the plan can't fully fund the raise
		// without breaching income. Clamp the target down by the shortfall so
		// the invariant sum(flexible)+savings <= income holds.
		if remaining > 0 {
			updated[targetIdx].Amount -= remaining
			if updated[targetIdx].Amount < 0 {
				updated[targetIdx].Amount = 0
			}
			remaining = 0
		}
	} else {
		// Target shrank: redistribute the freed amount (−delta) to other wants
		// first, then variable, proportionally to their current size.
		freed := -delta
		freed = push(updated, targetIdx, ExpenseWant, freed)
		if freed > 0 {
			freed = push(updated, targetIdx, ExpenseVariable, freed)
		}
		// If nothing else can take it (no other flexible categories), the freed
		// budget simply lifts savings — strictly within income.
		if freed > 0 {
			savingsTarget += freed
		}
	}

	return updated, savingsTarget, nil
}

// pull removes up to `amount` total from all categories of type t (excluding
// skipIdx), proportional to each category's current value, never below 0.
// Returns the amount that could NOT be pulled (0 if fully absorbed).
func pull(items []FlexItem, skipIdx int, t ExpenseType, amount int64) int64 {
	if amount <= 0 {
		return 0
	}
	var pool int64
	for i := range items {
		if i == skipIdx || items[i].Type != t {
			continue
		}
		pool += items[i].Amount
	}
	if pool <= 0 {
		return amount
	}

	take := amount
	if take > pool {
		take = pool // can't pull more than the pool holds
	}

	var pulled int64
	lastIdx := -1
	for i := range items {
		if i == skipIdx || items[i].Type != t || items[i].Amount == 0 {
			continue
		}
		lastIdx = i
		share := int64(float64(take) * float64(items[i].Amount) / float64(pool))
		if share > items[i].Amount {
			share = items[i].Amount
		}
		items[i].Amount -= share
		pulled += share
	}
	// Reconcile integer rounding: any residual comes off the last touched
	// category (bounded by what it still holds).
	if lastIdx != -1 && pulled < take {
		extra := take - pulled
		if extra > items[lastIdx].Amount {
			extra = items[lastIdx].Amount
		}
		items[lastIdx].Amount -= extra
		pulled += extra
	}

	return amount - pulled
}

// push adds up to `amount` total to all categories of type t (excluding
// skipIdx), proportional to each category's current value. If every such
// category is currently 0, it spreads evenly. Returns the amount that could NOT
// be pushed (0 if fully absorbed, i.e. nonzero only when no categories of type
// t exist besides skipIdx).
func push(items []FlexItem, skipIdx int, t ExpenseType, amount int64) int64 {
	if amount <= 0 {
		return 0
	}
	idxs := make([]int, 0, len(items))
	var pool int64
	for i := range items {
		if i == skipIdx || items[i].Type != t {
			continue
		}
		idxs = append(idxs, i)
		pool += items[i].Amount
	}
	if len(idxs) == 0 {
		return amount
	}

	var pushed int64
	if pool > 0 {
		for _, i := range idxs {
			share := int64(float64(amount) * float64(items[i].Amount) / float64(pool))
			items[i].Amount += share
			pushed += share
		}
	} else {
		// All zero — spread evenly.
		per := amount / int64(len(idxs))
		for _, i := range idxs {
			items[i].Amount += per
			pushed += per
		}
	}
	// Reconcile rounding residual onto the first category in the set.
	if pushed < amount {
		items[idxs[0]].Amount += amount - pushed
	}
	return 0
}

// roundTo rounds v to the nearest multiple of step (step > 0). Used to keep
// suggested amounts on clean Rp 10.000 boundaries.
func roundTo(v, step int64) int64 {
	if step <= 0 {
		return v
	}
	if v >= 0 {
		return ((v + step/2) / step) * step
	}
	return -((-v + step/2) / step) * step
}

// --- tiny stdlib-free string helpers (keep this file dependency-light, like
// allocator.go's itoa) ---

func lower(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'A' && c <= 'Z' {
			b[i] = c + ('a' - 'A')
		}
	}
	return string(b)
}

func contains(s, sub string) bool {
	if len(sub) == 0 {
		return true
	}
	if len(sub) > len(s) {
		return false
	}
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
