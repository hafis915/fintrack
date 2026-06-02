package budget

import (
	"time"

	"github.com/google/uuid"
)

type Plan struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	PeriodYear  int
	PeriodMonth int
	TotalIncome int64
	Program     string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Item struct {
	ID              uuid.UUID
	BudgetPlanID    uuid.UUID
	CategoryID      uuid.UUID
	CategoryName    string
	CategoryIcon    string
	CategoryType    string
	AllocatedAmount int64
	Percentage      float64
	IsDebtFocus     bool
}

type PlanWithItems struct {
	Plan
	Items []Item
}
