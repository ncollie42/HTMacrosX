package database

import (
	"testing"
)

func TestCaloriesFromGrams(t *testing.T) {
	tests := []struct {
		fat, carb, protein float64
		want               float32
	}{
		{10, 20, 30, float32(10*9 + 20*4 + 30*4)},
		{0, 0, 0, 0},
		{1, 1, 1, float32(9 + 4 + 4)},
	}
	for _, tt := range tests {
		got := CaloriesFromGrams(tt.fat, tt.carb, tt.protein)
		if got != tt.want {
			t.Errorf("CaloriesFromGrams(%v, %v, %v) = %v, want %v", tt.fat, tt.carb, tt.protein, got, tt.want)
		}
	}
}

func TestMacrosByGrams(t *testing.T) {
	mpg := MacroPerGram{
		ProteinPerGram: 0.1,
		FatPerGram:     0.05,
		CarbPerGram:    0.2,
		FiberPerGram:   0.02,
	}
	m := macrosByGrams(mpg, 200)

	wantProtein := float32(0.1 * 200)
	wantFat := float32(0.05 * 200)
	wantCarb := float32(0.2 * 200)
	wantFiber := float32(0.02 * 200)
	wantCal := wantProtein*4 + wantFat*9 + wantCarb*4

	if m.Protein != wantProtein {
		t.Errorf("Protein = %v, want %v", m.Protein, wantProtein)
	}
	if m.Fat != wantFat {
		t.Errorf("Fat = %v, want %v", m.Fat, wantFat)
	}
	if m.Carb != wantCarb {
		t.Errorf("Carb = %v, want %v", m.Carb, wantCarb)
	}
	if m.Fiber != wantFiber {
		t.Errorf("Fiber = %v, want %v", m.Fiber, wantFiber)
	}
	if m.Calories != wantCal {
		t.Errorf("Calories = %v, want %v", m.Calories, wantCal)
	}
}

func TestSumMacros(t *testing.T) {
	items := []MealSummary{
		{Macros: Macro{Calories: 100, Protein: 10, Fat: 5, Carb: 15, Fiber: 2}},
		{Macros: Macro{Calories: 200, Protein: 20, Fat: 10, Carb: 30, Fiber: 4}},
	}
	got := SumMacros(items)
	if got.Calories != 300 || got.Protein != 30 || got.Fat != 15 || got.Carb != 45 || got.Fiber != 6 {
		t.Errorf("SumMacros = %+v, want {300 30 15 45 6}", got)
	}
}

func TestSumMacrosEmpty(t *testing.T) {
	got := SumMacros(nil)
	if got.Calories != 0 {
		t.Errorf("SumMacros(nil) = %+v, want zero", got)
	}
}

func TestSumMacrosByID(t *testing.T) {
	items := []MealSummary{
		{ID: 1, Name: "Lunch", Macros: Macro{Calories: 100, Protein: 10}},
		{ID: 2, Name: "Dinner", Macros: Macro{Calories: 200, Protein: 20}},
		{ID: 1, Name: "Lunch", Macros: Macro{Calories: 50, Protein: 5}},
	}
	got := SumMacrosByID(items)

	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].ID != 1 || got[0].Macros.Calories != 150 || got[0].Macros.Protein != 15 {
		t.Errorf("got[0] = %+v, want ID=1 Cal=150 Prot=15", got[0])
	}
	if got[1].ID != 2 || got[1].Macros.Calories != 200 {
		t.Errorf("got[1] = %+v, want ID=2 Cal=200", got[1])
	}
	if got[0].Name != "Lunch" || got[1].Name != "Dinner" {
		t.Errorf("names wrong: %q, %q", got[0].Name, got[1].Name)
	}
}

func TestMakeMPG(t *testing.T) {
	mpg := makeMPG(0.1, 0.2, 0.3, 0.4)
	if mpg.ProteinPerGram != float32(0.1) {
		t.Errorf("ProteinPerGram = %v, want %v", mpg.ProteinPerGram, float32(0.1))
	}
	if mpg.FatPerGram != float32(0.2) {
		t.Errorf("FatPerGram = %v, want %v", mpg.FatPerGram, float32(0.2))
	}
	if mpg.CarbPerGram != float32(0.3) {
		t.Errorf("CarbPerGram = %v, want %v", mpg.CarbPerGram, float32(0.3))
	}
	if mpg.FiberPerGram != float32(0.4) {
		t.Errorf("FiberPerGram = %v, want %v", mpg.FiberPerGram, float32(0.4))
	}
}
