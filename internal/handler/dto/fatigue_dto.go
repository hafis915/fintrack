package dto

type FatigueCategoryResponse struct {
	CategoryID           string  `json:"category_id"`
	CategoryName         string  `json:"category_name"`
	CategoryIcon         string  `json:"category_icon,omitempty"`
	Type                 string  `json:"type"`
	Allocated            int64   `json:"allocated"`
	Spent                int64   `json:"spent"`
	Remaining            int64   `json:"remaining"`
	Percentage           float64 `json:"percentage"`
	Status               string  `json:"status"`
	DailyBudgetRemaining int64   `json:"daily_budget_remaining"`
	Tip                  string  `json:"tip,omitempty"`
}

type FatigueOverallResponse struct {
	TotalAllocated int64   `json:"total_allocated"`
	TotalSpent     int64   `json:"total_spent"`
	Percentage     float64 `json:"percentage"`
}

type FatigueDashboardResponse struct {
	Period        string                    `json:"period"`
	DayOfMonth    int                       `json:"day_of_month"`
	DaysRemaining int                       `json:"days_remaining"`
	Categories    []FatigueCategoryResponse `json:"categories"`
	Overall       FatigueOverallResponse    `json:"overall"`
}

type FatigueAlertResponse struct {
	Status         string  `json:"status"`
	CategoryName   string  `json:"category_name"`
	PercentageUsed float64 `json:"percentage_used"`
	Message        string  `json:"message"`
}
