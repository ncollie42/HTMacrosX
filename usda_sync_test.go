package main

import (
	"context"
	db "myapp/DB"
	"myapp/view"
	"strings"
	"testing"
)

func TestSyncUSDAFoundationFoodsImportsSearchableSharedFoods(t *testing.T) {
	userID := setupMainTestDB(t)

	if err := syncUSDAFoundationFoods(); err != nil {
		t.Fatalf("syncUSDAFoundationFoods returned error: %v", err)
	}

	results := db.FoodSearch("broccoli", userID)
	if len(results) == 0 {
		t.Fatal("expected USDA broccoli results after sync")
	}

	foundUSDA := false
	for _, food := range results {
		if food.Source == db.USDAFoundationSource {
			foundUSDA = true
			if food.Owned {
				t.Fatalf("USDA food unexpectedly marked owned: %+v", food)
			}
			break
		}
	}
	if !foundUSDA {
		t.Fatalf("results missing USDA food: %+v", results)
	}
}

func TestSyncUSDAFoundationFoodsIsIdempotentAndUserFoodsSortFirst(t *testing.T) {
	userID := setupMainTestDB(t)

	if _, err := db.CreateFood("Broccoli Custom", 5, 6, 1, 7, 100, userID); err != nil {
		t.Fatalf("CreateFood: %v", err)
	}
	if err := syncUSDAFoundationFoods(); err != nil {
		t.Fatalf("first syncUSDAFoundationFoods returned error: %v", err)
	}

	firstResults := db.FoodSearch("broccoli", userID)
	if len(firstResults) == 0 {
		t.Fatal("expected broccoli results after first sync")
	}
	if firstResults[0].Name != "Broccoli Custom" || !firstResults[0].Owned {
		t.Fatalf("first result = %+v, want owned food first", firstResults[0])
	}

	firstUSDA := 0
	for _, food := range firstResults {
		if food.Source == db.USDAFoundationSource {
			firstUSDA++
		}
	}

	if err := syncUSDAFoundationFoods(); err != nil {
		t.Fatalf("second syncUSDAFoundationFoods returned error: %v", err)
	}

	secondResults := db.FoodSearch("broccoli", userID)
	secondUSDA := 0
	for _, food := range secondResults {
		if food.Source == db.USDAFoundationSource {
			secondUSDA++
		}
	}
	if secondUSDA != firstUSDA {
		t.Fatalf("USDA result count changed after second sync: first=%d second=%d", firstUSDA, secondUSDA)
	}
}

func TestUSDAFoodDoesNotRenderDeleteAction(t *testing.T) {
	rendered := new(strings.Builder)
	component := view.FoodSearchResults([]db.Food{{
		ID:     99,
		Name:   "Broccoli, raw",
		Source: db.USDAFoundationSource,
		Owned:  false,
		Macros: db.Macro{Fat: 0.4, Carb: 7, Fiber: 3, Protein: 2.8},
	}}, "/meal/1/join/", "")

	if err := component.Render(context.Background(), rendered); err != nil {
		t.Fatalf("render returned error: %v", err)
	}

	body := rendered.String()
	if !strings.Contains(body, "USDA") {
		t.Fatalf("body missing USDA badge: %q", body)
	}
	if strings.Contains(body, "Delete this ingredient?") {
		t.Fatalf("body unexpectedly includes delete action: %q", body)
	}
}

func TestUSDAFoodCannotBeDeletedByNormalUser(t *testing.T) {
	userID := setupMainTestDB(t)

	if err := syncUSDAFoundationFoods(); err != nil {
		t.Fatalf("syncUSDAFoundationFoods returned error: %v", err)
	}

	results := db.FoodSearch("broccoli", userID)
	for _, food := range results {
		if food.Source == db.USDAFoundationSource {
			if err := db.DeleteFood(food.ID, userID); err != db.ErrNotOwned {
				t.Fatalf("DeleteFood err = %v, want %v", err, db.ErrNotOwned)
			}
			return
		}
	}
	t.Fatal("expected USDA broccoli result for delete test")
}
