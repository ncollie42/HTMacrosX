package database

import (
	"errors"
	"path/filepath"
	"testing"
	"time"
)

func setupTestDB(t *testing.T) {
	t.Helper()
	if sqlDB != nil {
		_ = sqlDB.Close()
	}
	Open(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() {
		if sqlDB != nil {
			_ = sqlDB.Close()
			sqlDB = nil
		}
	})
}

func createTestUser(t *testing.T, username string) int {
	t.Helper()
	userID, err := CreateUser(username, "pass")
	if err != nil {
		t.Fatalf("CreateUser(%q): %v", username, err)
	}
	return userID
}

func createTestFood(t *testing.T, userID int, name string) int {
	t.Helper()
	foodID, err := CreateFood(name, 10, 20, 5, 15, 100, userID)
	if err != nil {
		t.Fatalf("CreateFood(%q): %v", name, err)
	}
	return foodID
}

func createTestMeal(t *testing.T, userID int, isPreset bool, name string) int {
	t.Helper()
	mealID, err := CreateMeal(name, userID, isPreset, time.Date(2026, 3, 8, 12, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("CreateMeal(%q): %v", name, err)
	}
	return mealID
}

func createTestItem(t *testing.T, mealID int, foodID int, userID int, isPreset bool) int {
	t.Helper()
	if err := CreateMealItem(mealID, foodID, 100, userID, isPreset); err != nil {
		t.Fatalf("CreateMealItem(meal=%d, food=%d): %v", mealID, foodID, err)
	}
	var itemID int
	err := sqlDB.QueryRow(`SELECT id FROM meal_items WHERE meal_id = ? AND food_id = ? ORDER BY id DESC LIMIT 1`, mealID, foodID).Scan(&itemID)
	if err != nil {
		t.Fatalf("fetch meal item id: %v", err)
	}
	return itemID
}

func TestMealItemMutationsBoundToMealAndType(t *testing.T) {
	setupTestDB(t)

	userID := createTestUser(t, "alice")
	foodID := createTestFood(t, userID, "Chicken")
	mealA := createTestMeal(t, userID, false, "Meal A")
	mealB := createTestMeal(t, userID, false, "Meal B")
	templateID := createTestMeal(t, userID, true, "Template")
	itemID := createTestItem(t, mealB, foodID, userID, false)

	if err := UpdateMealItem(mealB, itemID, 150, userID, false); err != nil {
		t.Fatalf("UpdateMealItem valid: %v", err)
	}
	if err := UpdateMealItem(mealA, itemID, 175, userID, false); !errors.Is(err, ErrNotOwned) {
		t.Fatalf("UpdateMealItem wrong meal err = %v, want ErrNotOwned", err)
	}
	if err := DeleteMealItem(mealA, itemID, userID, false); !errors.Is(err, ErrNotOwned) {
		t.Fatalf("DeleteMealItem wrong meal err = %v, want ErrNotOwned", err)
	}
	if err := CreateMealItem(mealA, foodID, 100, userID, true); !errors.Is(err, ErrNotOwned) {
		t.Fatalf("CreateMealItem wrong type err = %v, want ErrNotOwned", err)
	}
	if err := CreateMealItem(templateID, foodID, 100, userID, false); !errors.Is(err, ErrNotOwned) {
		t.Fatalf("CreateMealItem wrong template type err = %v, want ErrNotOwned", err)
	}
	if err := DeleteMealItem(mealB, itemID, userID, false); err != nil {
		t.Fatalf("DeleteMealItem valid: %v", err)
	}
}

func TestMealTypeEnforcement(t *testing.T) {
	setupTestDB(t)

	userID := createTestUser(t, "bob")
	foodID := createTestFood(t, userID, "Rice")
	mealID := createTestMeal(t, userID, false, "Lunch")
	templateID := createTestMeal(t, userID, true, "Saved Lunch")
	createTestItem(t, templateID, foodID, userID, true)

	if _, err := GetMealByID(templateID, userID, false); !errors.Is(err, ErrNotOwned) {
		t.Fatalf("GetMealByID template as meal err = %v, want ErrNotOwned", err)
	}
	if _, err := GetMealByID(mealID, userID, true); !errors.Is(err, ErrNotOwned) {
		t.Fatalf("GetMealByID meal as template err = %v, want ErrNotOwned", err)
	}
	if err := UpdateMealName(templateID, userID, false, "Wrong"); !errors.Is(err, ErrNotOwned) {
		t.Fatalf("UpdateMealName template as meal err = %v, want ErrNotOwned", err)
	}
	if err := DeleteMeal(mealID, userID, true); !errors.Is(err, ErrNotOwned) {
		t.Fatalf("DeleteMeal meal as template err = %v, want ErrNotOwned", err)
	}
	if err := DeleteMeal(templateID, userID, false); !errors.Is(err, ErrNotOwned) {
		t.Fatalf("DeleteMeal template as meal err = %v, want ErrNotOwned", err)
	}
	if _, err := TemplateToMeal(mealID, userID, time.Date(2026, 3, 8, 12, 0, 0, 0, time.Local)); !errors.Is(err, ErrNotOwned) {
		t.Fatalf("TemplateToMeal normal meal err = %v, want ErrNotOwned", err)
	}
}
