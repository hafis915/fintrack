package fatigue

import "testing"

func TestClassify_StatusBuckets(t *testing.T) {
	cases := []struct {
		name      string
		allocated int64
		spent     int64
		want      Status
	}{
		{"zero_spent_is_fresh", 1_000_000, 0, Fresh},
		{"under_70_is_fresh", 1_000_000, 690_000, Fresh},
		{"at_70_is_warning", 1_000_000, 700_000, Warning},
		{"just_below_100_is_warning", 1_000_000, 990_000, Warning},
		{"at_100_is_fatigued", 1_000_000, 1_000_000, Fatigued},
		{"over_100_is_fatigued", 1_000_000, 1_500_000, Fatigued},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Classify(tc.allocated, tc.spent)
			if got.Status != tc.want {
				t.Errorf("status: want %q, got %q (pct=%.2f)", tc.want, got.Status, got.PercentageUsed)
			}
		})
	}
}

func TestClassify_RemainingMatchesArithmetic(t *testing.T) {
	r := Classify(1_000_000, 250_000)
	if r.Remaining != 750_000 {
		t.Errorf("remaining: want 750_000, got %d", r.Remaining)
	}
}

func TestClassify_RemainingNegativeWhenFatigued(t *testing.T) {
	r := Classify(1_000_000, 1_200_000)
	if r.Remaining != -200_000 {
		t.Errorf("remaining: want -200_000 (over budget), got %d", r.Remaining)
	}
}

func TestClassify_ZeroAllocatedYieldsFreshWithoutDivByZero(t *testing.T) {
	r := Classify(0, 50_000)
	if r.Status != Fresh {
		t.Errorf("want Fresh for zero-allocated, got %q", r.Status)
	}
	if r.PercentageUsed != 0 {
		t.Errorf("want 0 percentage, got %.2f", r.PercentageUsed)
	}
}

func TestClassify_NegativeSpendTreatedAsZeroForRatio(t *testing.T) {
	// Net refund — should not yield a negative percentage_used.
	r := Classify(1_000_000, -200_000)
	if r.PercentageUsed != 0 {
		t.Errorf("negative spend should yield 0%% used, got %.2f", r.PercentageUsed)
	}
	if r.Remaining != 1_200_000 {
		t.Errorf("remaining should reflect the credit: want 1_200_000, got %d", r.Remaining)
	}
}

func TestClassify_CoachingCopyIsSet(t *testing.T) {
	for _, s := range []int64{0, 800_000, 1_100_000} {
		r := Classify(1_000_000, s)
		if r.Coaching == "" {
			t.Errorf("missing coaching copy for spent=%d (status=%s)", s, r.Status)
		}
	}
}

func TestBuildSummary_PercentageOverIncome(t *testing.T) {
	s := BuildSummary(8_000_000, 3_640_000, 0, 8_000_000)
	if s.OverallPercentage != 45.50 {
		t.Errorf("overall percentage: want 45.50, got %.2f", s.OverallPercentage)
	}
}

func TestBuildSummary_ZeroIncomeYieldsZeroPercentage(t *testing.T) {
	s := BuildSummary(0, 0, 0, 0)
	if s.OverallPercentage != 0 {
		t.Errorf("zero income should give 0 percentage, got %.2f", s.OverallPercentage)
	}
}
