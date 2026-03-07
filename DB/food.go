package database

import (
	"fmt"
	"log"
	"strings"
)

func CreateFood(name string, fat float64, carb float64, fiber float64, protein float64, grams float64, userID int) (int, error) {
	return CreateFoodWithBarcode(name, fat, carb, fiber, protein, grams, userID, "")
}

func CreateFoodWithBarcode(name string, fat float64, carb float64, fiber float64, protein float64, grams float64, userID int, barcode string) (int, error) {
	res, err := sqlDB.Exec(
		`INSERT INTO foods (name, protein_per_gram, fat_per_gram, carb_per_gram, fiber_per_gram, grams, creator_user_id, barcode) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		name, protein/grams, fat/grams, carb/grams, fiber/grams, grams, userID, barcode,
	)
	if err != nil {
		return 0, fmt.Errorf("CreateFoodWithBarcode: %w", err)
	}
	id, _ := res.LastInsertId()
	return int(id), nil
}

func FindFoodByBarcode(barcode string) *FoodRecord {
	var f FoodRecord
	var ppg, fpg, cpg, fibpg, grams float64
	err := sqlDB.QueryRow(
		`SELECT id, name, protein_per_gram, fat_per_gram, carb_per_gram, fiber_per_gram, grams, creator_user_id, barcode FROM foods WHERE barcode = ?`,
		barcode,
	).Scan(&f.ID, &f.Name, &ppg, &fpg, &cpg, &fibpg, &grams, &f.CreatorUserID, &f.Barcode)
	if err != nil {
		return nil
	}
	f.ProteinPerGram = float32(ppg)
	f.FatPerGram = float32(fpg)
	f.CarbPerGram = float32(cpg)
	f.FiberPerGram = float32(fibpg)
	f.Grams = float32(grams)
	return &f
}

func DeleteFood(foodID int, userID int) error {
	tx, err := sqlDB.Begin()
	if err != nil {
		return err
	}
	res, err := tx.Exec(`DELETE FROM foods WHERE id = ? AND creator_user_id = ?`, foodID, userID)
	if err != nil {
		tx.Rollback()
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		tx.Rollback()
		return ErrNotOwned
	}
	if _, err := tx.Exec(`DELETE FROM meal_items WHERE food_id = ?`, foodID); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func FoodSearch(name string, userID int) []Food {
	likePattern := "%" + strings.ToLower(name) + "%"
	rows, err := sqlDB.Query(
		`SELECT id, name, protein_per_gram, fat_per_gram, carb_per_gram, fiber_per_gram, grams
		FROM foods
		WHERE (creator_user_id = ? OR creator_user_id = ?)
		  AND (? = '' OR LOWER(name) LIKE ?)
		ORDER BY name`,
		SystemUserID, userID, name, likePattern,
	)
	if err != nil {
		log.Printf("FoodSearch: %v", err)
		return nil
	}
	defer rows.Close()

	var result []Food
	for rows.Next() {
		var id int
		var fname string
		var ppg, fpg, cpg, fibpg, grams float64
		if err := rows.Scan(&id, &fname, &ppg, &fpg, &cpg, &fibpg, &grams); err != nil {
			continue
		}
		result = append(result, Food{
			ID:     id,
			Name:   fname,
			Grams:  float32(grams),
			Macros: macrosByGrams(makeMPG(ppg, fpg, cpg, fibpg), 100),
		})
	}
	return result
}
