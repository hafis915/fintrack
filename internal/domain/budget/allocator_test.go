package budget

import (
	"testing"

	"github.com/google/uuid"
)

func TestAllocate_ComputesBucketsAndTabungan(t *testing.T) {
	rent := IntakeItem{CategoryID: uuid.New(), Name: "Sewa kosan", Icon: "🏠", Type: ExpenseFixed, Amount: 1_200_000}
	kpr := IntakeItem{CategoryID: uuid.New(), Name: "Cicilan KPR", Icon: "🏗", Type: ExpenseFixed, Amount: 1_500_000}
	food := IntakeItem{CategoryID: uuid.New(), Name: "Makan & minum", Icon: "🍱", Type: ExpenseVariable, Amount: 1_200_000}
	cc := IntakeItem{CategoryID: uuid.New(), Name: "Kartu kredit", Icon: "💳", Type: ExpenseDebt, Amount: 400_000}
	fun := IntakeItem{CategoryID: uuid.New(), Name: "Hiburan", Icon: "🎬", Type: ExpenseWant, Amount: 500_000}

	plan, err := Allocate(IntakeAnswers{
		Income: 8_000_000, HousingType: HousingKpr, Goal: GoalDebt,
		DebtTypes: []DebtType{DebtCC}, EmergencyMonths: 1, LifestyleStyle: LifestyleBalanced,
	}, []IntakeItem{rent, kpr, food, cc, fun})
	if err != nil {
		t.Fatalf("Allocate: %v", err)
	}

	if plan.Program != ProgramBebasUtang {
		t.Errorf("program: want bebas_utang, got %q", plan.Program)
	}
	if plan.Summary.Kebutuhan.Amount != 3_900_000 {
		t.Errorf("kebutuhan amount: want 3_900_000, got %d", plan.Summary.Kebutuhan.Amount)
	}
	if plan.Summary.Utang.Amount != 400_000 {
		t.Errorf("utang amount: want 400_000, got %d", plan.Summary.Utang.Amount)
	}
	if plan.Summary.Keinginan.Amount != 500_000 {
		t.Errorf("keinginan amount: want 500_000, got %d", plan.Summary.Keinginan.Amount)
	}
	// Tabungan = 8M - 3.9M - 0.4M - 0.5M = 3.2M
	if plan.Summary.Tabungan.Amount != 3_200_000 {
		t.Errorf("tabungan amount: want 3_200_000, got %d", plan.Summary.Tabungan.Amount)
	}
	if plan.Summary.Tabungan.Percentage != 40.00 {
		t.Errorf("tabungan percentage: want 40.00, got %.2f", plan.Summary.Tabungan.Percentage)
	}
}

func TestAllocate_FlagsDebtFocusOnlyForBebasUtang(t *testing.T) {
	cc := IntakeItem{CategoryID: uuid.New(), Name: "Kartu kredit", Type: ExpenseDebt, Amount: 500_000}

	plan, err := Allocate(IntakeAnswers{
		Income: 8_000_000, HousingType: HousingKpr, Goal: GoalDebt,
		DebtTypes: []DebtType{DebtCC}, EmergencyMonths: 1, LifestyleStyle: LifestyleBalanced,
	}, []IntakeItem{cc})
	if err != nil {
		t.Fatal(err)
	}
	if !plan.Items[0].IsDebtFocus {
		t.Error("expected debt item to be flagged is_debt_focus under bebas_utang")
	}

	// Same item but program will be seimbang → not flagged
	plan2, err := Allocate(IntakeAnswers{
		Income: 8_000_000, HousingType: HousingKpr, Goal: GoalBalance,
		DebtTypes: []DebtType{DebtNone}, EmergencyMonths: 3, LifestyleStyle: LifestyleBalanced,
	}, []IntakeItem{cc})
	if err != nil {
		t.Fatal(err)
	}
	if plan2.Items[0].IsDebtFocus {
		t.Error("debt focus flag leaked into non-bebas_utang program")
	}
}

func TestAllocate_WarnsWhenInsolvent(t *testing.T) {
	rent := IntakeItem{CategoryID: uuid.New(), Name: "Sewa", Type: ExpenseFixed, Amount: 6_000_000}
	food := IntakeItem{CategoryID: uuid.New(), Name: "Makan", Type: ExpenseVariable, Amount: 4_000_000}
	plan, err := Allocate(IntakeAnswers{
		Income: 8_000_000, HousingType: HousingKosan, Goal: GoalBalance,
		DebtTypes: []DebtType{DebtNone}, EmergencyMonths: 1, LifestyleStyle: LifestyleBalanced,
	}, []IntakeItem{rent, food})
	if err != nil {
		t.Fatal(err)
	}
	if plan.Summary.Tabungan.Amount >= 0 {
		t.Fatalf("expected negative tabungan, got %d", plan.Summary.Tabungan.Amount)
	}
	if plan.Warning == "" {
		t.Error("expected an insolvency warning, got empty")
	}
}

func TestAllocate_RejectsNegativeItemAmount(t *testing.T) {
	bad := IntakeItem{CategoryID: uuid.New(), Name: "Refund", Type: ExpenseFixed, Amount: -100}
	_, err := Allocate(IntakeAnswers{
		Income: 8_000_000, HousingType: HousingKpr, Goal: GoalBalance,
		DebtTypes: []DebtType{DebtNone}, EmergencyMonths: 3, LifestyleStyle: LifestyleBalanced,
	}, []IntakeItem{bad})
	if err == nil {
		t.Fatal("expected error on negative item amount")
	}
}

func TestAllocate_ItemsSortedByTypeThenName(t *testing.T) {
	a := IntakeItem{CategoryID: uuid.New(), Name: "Nongkrong", Type: ExpenseWant, Amount: 100}
	b := IntakeItem{CategoryID: uuid.New(), Name: "Hiburan", Type: ExpenseWant, Amount: 100}
	c := IntakeItem{CategoryID: uuid.New(), Name: "Sewa", Type: ExpenseFixed, Amount: 100}
	d := IntakeItem{CategoryID: uuid.New(), Name: "Makan", Type: ExpenseVariable, Amount: 100}

	plan, err := Allocate(IntakeAnswers{
		Income: 8_000_000, HousingType: HousingKpr, Goal: GoalBalance,
		DebtTypes: []DebtType{DebtNone}, EmergencyMonths: 3, LifestyleStyle: LifestyleBalanced,
	}, []IntakeItem{a, b, c, d})
	if err != nil {
		t.Fatal(err)
	}
	wantOrder := []string{"Sewa", "Makan", "Hiburan", "Nongkrong"}
	for i, w := range wantOrder {
		if plan.Items[i].CategoryName != w {
			t.Errorf("position %d: want %q, got %q", i, w, plan.Items[i].CategoryName)
		}
	}
}
