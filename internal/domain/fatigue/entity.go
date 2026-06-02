package fatigue

import "github.com/google/uuid"

type CategorySnapshot struct {
	CategoryID           uuid.UUID
	CategoryName         string
	CategoryIcon         string
	Type                 string
	Allocated            int64
	Spent                int64
	Remaining            int64
	Percentage           float64
	Status               string
	DailyBudgetRemaining int64
	Tip                  string
}

type Overall struct {
	TotalAllocated int64
	TotalSpent     int64
	Percentage     float64
}

type Snapshot struct {
	Period        string
	DayOfMonth    int
	DaysRemaining int
	Categories    []CategorySnapshot
	Overall       Overall
}

type Alert struct {
	Status         string
	CategoryName   string
	PercentageUsed float64
	Message        string
}

func ComputeStatus(spent, allocated int64) string {
	if allocated == 0 {
		return "fresh"
	}
	p := float64(spent) / float64(allocated) * 100
	switch {
	case p >= 85:
		return "fatigued"
	case p >= 60:
		return "warning"
	default:
		return "fresh"
	}
}
