// Package fatigue holds the pure rules behind the Category Fatigue Dashboard
// (MVP feature #4). Status thresholds and coaching copy live here so the
// rest of the stack (handler, frontend) only has to transport classifier
// output.
package fatigue

import "fmt"

// Status maps to the fatigue_status enum on the SQL side. The values must
// stay in sync with migration 0002.
type Status string

const (
	Fresh    Status = "fresh"
	Warning  Status = "warning"
	Fatigued Status = "fatigued"
)

// Thresholds (percentage of allocated spent):
//   < 70   → fresh
//   70-99  → warning
//   >= 100 → fatigued
const (
	WarningThreshold  = 70.0
	FatiguedThreshold = 100.0
)

// Result is the per-category state the dashboard renders.
type Result struct {
	Status         Status
	PercentageUsed float64 // 2 d.p.
	Remaining      int64   // can be negative when fatigued
	Coaching       string
}

// Classify returns the Result for a (allocated, spent) pair.
//
// Edge cases:
//   - allocated == 0 → can't compute a ratio. We return fresh with 0%
//     usage; the dashboard should hide unallocated categories from the
//     status grid (they belong in the "unallocated" bucket instead).
//   - spent < 0 (refund net) → treat as 0 for ratio purposes but keep the
//     numeric remaining so the UI shows the credit.
func Classify(allocated, spent int64) Result {
	r := Result{
		Remaining: allocated - spent,
	}

	if allocated <= 0 {
		r.Status = Fresh
		r.Coaching = coachingFor(Fresh, "")
		return r
	}

	effSpent := spent
	if effSpent < 0 {
		effSpent = 0
	}

	pct := float64(effSpent) / float64(allocated) * 100
	r.PercentageUsed = roundTwo(pct)

	switch {
	case pct >= FatiguedThreshold:
		r.Status = Fatigued
	case pct >= WarningThreshold:
		r.Status = Warning
	default:
		r.Status = Fresh
	}
	r.Coaching = coachingFor(r.Status, "")
	return r
}

func roundTwo(v float64) float64 {
	if v >= 0 {
		return float64(int64(v*100+0.5)) / 100
	}
	return float64(int64(v*100-0.5)) / 100
}

// Summary is the overall envelope returned by GET /v1/budget/current.
type Summary struct {
	TotalAllocated    int64
	TotalSpent        int64
	UnallocatedSpent  int64
	OverallPercentage float64 // total_spent / total_income, 2 d.p.
}

// BuildSummary computes the dashboard summary from item totals + income.
// Income is used as the denominator for OverallPercentage (matches the
// PRD example where overall_percentage = spent / total_income).
func BuildSummary(totalAllocated, totalSpent, unallocatedSpent, totalIncome int64) Summary {
	pct := 0.0
	if totalIncome > 0 {
		pct = roundTwo(float64(totalSpent) / float64(totalIncome) * 100)
	}
	return Summary{
		TotalAllocated:    totalAllocated,
		TotalSpent:        totalSpent,
		UnallocatedSpent:  unallocatedSpent,
		OverallPercentage: pct,
	}
}

// coachingFor returns the user-facing copy for each status. categoryName,
// when non-empty, allows future personalisation ("Makan & minum sudah
// 95% — slow down minggu ini") — currently unused, kept in the signature
// so handlers can pass it through without an API churn.
func coachingFor(s Status, categoryName string) string {
	switch s {
	case Fresh:
		return "Masih fresh — pertahankan."
	case Warning:
		if categoryName != "" {
			return fmt.Sprintf("Hampir habis — slow down %s minggu ini.", categoryName)
		}
		return "Hampir habis — slow down minggu ini."
	case Fatigued:
		if categoryName != "" {
			return fmt.Sprintf("%s sudah lewat batas — pertimbangkan stop dulu.", categoryName)
		}
		return "Sudah lewat batas — pertimbangkan stop dulu."
	}
	return ""
}
