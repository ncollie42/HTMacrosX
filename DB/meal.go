package database

import (
	"fmt"
	"time"
)

func CreateMeal(name string, userID int, isPreset bool) (int, error) {
	today := ""
	if !isPreset {
		today = time.Now().Format("2006-01-02")
	}
	isPresetInt := 0
	if isPreset {
		isPresetInt = 1
	}
	res, err := sqlDB.Exec(
		`INSERT INTO meals (user_id, name, meal_date, is_preset) VALUES (?, ?, ?, ?)`,
		userID, name, today, isPresetInt,
	)
	if err != nil {
		return 0, fmt.Errorf("CreateMeal: %w", err)
	}
	id, _ := res.LastInsertId()
	return int(id), nil
}

func DeleteMeal(mealID int, userID int) error {
	tx, err := sqlDB.Begin()
	if err != nil {
		return err
	}
	res, err := tx.Exec(`DELETE FROM meals WHERE id = ? AND user_id = ?`, mealID, userID)
	if err != nil {
		tx.Rollback()
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		tx.Rollback()
		return ErrNotOwned
	}
	if _, err := tx.Exec(`DELETE FROM meal_items WHERE meal_id = ?`, mealID); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func GetMealByID(mealID int, userID int) (Meal, error) {
	var model Meal
	model.ID = mealID

	rows, err := sqlDB.Query(`
		SELECT m.name, j.id, j.grams, f.name, f.protein_per_gram, f.fat_per_gram, f.carb_per_gram, f.fiber_per_gram
		FROM meals m
		LEFT JOIN meal_items j ON j.meal_id = m.id
		LEFT JOIN foods f ON f.id = j.food_id
		WHERE m.id = ? AND m.user_id = ?
	`, mealID, userID)
	if err != nil {
		return model, err
	}
	defer rows.Close()

	found := false
	for rows.Next() {
		found = true
		var mealName string
		var jid *int
		var grams, ppg, fpg, cpg, fibpg *float64
		var fname *string
		if err := rows.Scan(&mealName, &jid, &grams, &fname, &ppg, &fpg, &cpg, &fibpg); err != nil {
			continue
		}
		model.Name = mealName
		if jid == nil {
			continue // meal has no foods yet
		}
		model.Items = append(model.Items, MealItem{
			Name:   *fname,
			ItemID: *jid,
			Grams:  float32(*grams),
			Macros: macrosByGrams(makeMPG(*ppg, *fpg, *cpg, *fibpg), float32(*grams)),
		})
	}
	if !found {
		return model, ErrNotOwned
	}
	return model, nil
}

func UpdateMealName(mealID int, userID int, name string) error {
	res, err := sqlDB.Exec(`UPDATE meals SET name = ? WHERE id = ? AND user_id = ?`, name, mealID, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotOwned
	}
	return nil
}

func GetTemplates(userID int) []MealSummary {
	rows, err := sqlDB.Query(`
		SELECT j.grams, m.name, m.id, f.protein_per_gram, f.fat_per_gram, f.carb_per_gram, f.fiber_per_gram
		FROM meal_items j
		JOIN meals m ON m.id = j.meal_id
		JOIN foods f ON f.id = j.food_id
		WHERE m.user_id = ? AND m.is_preset = 1
	`, userID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	return scanMealSummaryRows(rows)
}

func TemplateToMeal(templateID int, userID int) (int, error) {
	tx, err := sqlDB.Begin()
	if err != nil {
		return 0, err
	}

	var presetName string
	if err := tx.QueryRow(`SELECT name FROM meals WHERE id = ? AND user_id = ?`, templateID, userID).Scan(&presetName); err != nil {
		tx.Rollback()
		return 0, ErrNotOwned
	}

	rows, err := tx.Query(`SELECT food_id, grams FROM meal_items WHERE meal_id = ?`, templateID)
	if err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("TemplateToMeal: %w", err)
	}
	type item struct {
		FoodID int
		Grams  float64
	}
	var items []item
	for rows.Next() {
		var it item
		if err := rows.Scan(&it.FoodID, &it.Grams); err != nil {
			continue
		}
		items = append(items, it)
	}
	rows.Close()

	today := time.Now().Format("2006-01-02")
	res, err := tx.Exec(
		`INSERT INTO meals (user_id, name, meal_date, is_preset) VALUES (?, ?, ?, 0)`,
		userID, presetName, today,
	)
	if err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("TemplateToMeal: %w", err)
	}
	mealID, _ := res.LastInsertId()

	for _, it := range items {
		if _, err := tx.Exec(`INSERT INTO meal_items (meal_id, food_id, grams) VALUES (?, ?, ?)`, mealID, it.FoodID, it.Grams); err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("TemplateToMeal: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("TemplateToMeal: %w", err)
	}
	return int(mealID), nil
}
