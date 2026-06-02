package dto

type OnboardingItemRequest struct {
	Name       string `json:"name"        validate:"required,min=1,max=100"`
	Icon       string `json:"icon"`
	Type       string `json:"type"        validate:"required,oneof=fixed variable debt want"`
	Amount     int64  `json:"amount"      validate:"required,gte=0"`
	CategoryID string `json:"category_id" validate:"required,uuid"`
}

type OnboardingRequest struct {
	Income          int64                   `json:"income"          validate:"required,gt=0"`
	Goal            string                  `json:"goal"            validate:"required,oneof=emergency debt goal invest balance"`
	HousingType     string                  `json:"housing_type"    validate:"omitempty,oneof=kosan kpr keluarga"`
	LifestyleStyle  string                  `json:"lifestyle_style" validate:"omitempty,oneof=easy balanced strict"`
	EmergencyMonths int                     `json:"emergency_months" validate:"oneof=0 1 3 6"`
	DebtTypes       []string                `json:"debt_types"`
	ExpenseItems    []OnboardingItemRequest `json:"expense_items"   validate:"required,min=1,dive"`
}

type SummaryGroupResponse struct {
	Amount     int64   `json:"amount"`
	Percentage float64 `json:"percentage"`
}

type AllocationSummaryResponse struct {
	Kebutuhan SummaryGroupResponse `json:"kebutuhan"`
	Utang     SummaryGroupResponse `json:"utang"`
	Keinginan SummaryGroupResponse `json:"keinginan"`
	Tabungan  SummaryGroupResponse `json:"tabungan"`
	Total     int64                `json:"total"`
}

type BudgetItemResponse struct {
	ID              string  `json:"id,omitempty"`
	CategoryID      string  `json:"category_id"`
	CategoryName    string  `json:"category_name"`
	CategoryIcon    string  `json:"category_icon,omitempty"`
	CategoryType    string  `json:"category_type"`
	AllocatedAmount int64   `json:"allocated_amount"`
	Percentage      float64 `json:"percentage"`
	IsDebtFocus     bool    `json:"is_debt_focus"`
}

type OnboardingResponse struct {
	BudgetPlanID string                    `json:"budget_plan_id"`
	Program      string                    `json:"program"`
	IncomeHint   string                    `json:"income_hint"`
	Warning      string                    `json:"warning,omitempty"`
	Summary      AllocationSummaryResponse `json:"summary"`
	Items        []BudgetItemResponse      `json:"items"`
}

type CurrentBudgetResponse struct {
	BudgetPlanID string               `json:"budget_plan_id"`
	Year         int                  `json:"year"`
	Month        int                  `json:"month"`
	Program      string               `json:"program"`
	TotalIncome  int64                `json:"total_income"`
	Items        []BudgetItemResponse `json:"items"`
}
