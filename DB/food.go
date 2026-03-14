package database

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
)

var ErrFoodInUse = fmt.Errorf("food is in use")

const USDAFoundationSource = "usda_foundation"

func CreateFood(name string, fat float64, carb float64, fiber float64, protein float64, grams float64, userID int) (int, error) {
	return CreateFoodWithBarcode(name, fat, carb, fiber, protein, grams, userID, "")
}

func CreateFoodWithBarcode(name string, fat float64, carb float64, fiber float64, protein float64, grams float64, userID int, barcode string) (int, error) {
	res, err := sqlDB.Exec(
		`INSERT INTO foods (name, protein_per_gram, fat_per_gram, carb_per_gram, fiber_per_gram, grams, creator_user_id, barcode) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		name, protein/grams, fat/grams, carb/grams, fiber/grams, grams, userID, barcode,
	)
	if err != nil {
		if barcode != "" && strings.Contains(err.Error(), "UNIQUE constraint failed") {
			if existing := FindFoodByBarcode(barcode, userID); existing != nil && existing.CreatorUserID == userID {
				return existing.ID, nil
			}
		}
		return 0, fmt.Errorf("CreateFoodWithBarcode: %w", err)
	}
	id, _ := res.LastInsertId()
	return int(id), nil
}

func CreateFoodAndMealItem(name string, fat float64, carb float64, fiber float64, protein float64, servingGrams float64, userID int, barcode string, mealID int, isPreset bool, itemGrams float64) (int, error) {
	tx, err := sqlDB.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	if err := validateMealAccessTx(tx, mealID, userID, isPreset); err != nil {
		return 0, err
	}

	res, err := tx.Exec(
		`INSERT INTO foods (name, protein_per_gram, fat_per_gram, carb_per_gram, fiber_per_gram, grams, creator_user_id, barcode, is_shared) VALUES (?, ?, ?, ?, ?, ?, ?, ?, 0)`,
		name, protein/servingGrams, fat/servingGrams, carb/servingGrams, fiber/servingGrams, servingGrams, userID, barcode,
	)
	if err != nil {
		return 0, fmt.Errorf("CreateFoodAndMealItem: %w", err)
	}
	foodID, _ := res.LastInsertId()
	if _, err := tx.Exec(`INSERT INTO meal_items (meal_id, food_id, grams) VALUES (?, ?, ?)`, mealID, foodID, itemGrams); err != nil {
		return 0, fmt.Errorf("CreateFoodAndMealItem: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("CreateFoodAndMealItem: %w", err)
	}
	return int(foodID), nil
}

func FindFoodByBarcode(barcode string, userID int) *FoodRecord {
	var f FoodRecord
	var ppg, fpg, cpg, fibpg, grams float64
	err := sqlDB.QueryRow(
		`SELECT id, name, protein_per_gram, fat_per_gram, carb_per_gram, fiber_per_gram, grams, creator_user_id, barcode, source, source_ref
		FROM foods
		WHERE barcode = ? AND (creator_user_id = ? OR is_shared = 1)
		ORDER BY CASE WHEN creator_user_id = ? THEN 0 ELSE 1 END, id
		LIMIT 1`,
		barcode, userID, userID,
	).Scan(&f.ID, &f.Name, &ppg, &fpg, &cpg, &fibpg, &grams, &f.CreatorUserID, &f.Barcode, &f.Source, &f.SourceRef)
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
	var refCount int
	if err := tx.QueryRow(
		`SELECT COUNT(*) FROM meal_items mi
		JOIN foods f ON f.id = mi.food_id
		WHERE mi.food_id = ? AND f.creator_user_id = ?`,
		foodID, userID,
	).Scan(&refCount); err != nil {
		tx.Rollback()
		return err
	}
	if refCount > 0 {
		tx.Rollback()
		return ErrFoodInUse
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
	return tx.Commit()
}

func FoodSearch(name string, userID int) []Food {
	likePattern := "%" + strings.ToLower(name) + "%"
	rows, err := sqlDB.Query(
		`SELECT id, name, protein_per_gram, fat_per_gram, carb_per_gram, fiber_per_gram, grams,
		        CASE WHEN creator_user_id = ? THEN 1 ELSE 0 END AS owned, source
		FROM foods
		WHERE (creator_user_id = ? OR is_shared = 1)
		  AND (? = '' OR LOWER(name) LIKE ?)
		ORDER BY CASE WHEN creator_user_id = ? THEN 0 ELSE 1 END, name`,
		userID, userID, name, likePattern, userID,
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
		var owned int
		var source string
		var ppg, fpg, cpg, fibpg, grams float64
		if err := rows.Scan(&id, &fname, &ppg, &fpg, &cpg, &fibpg, &grams, &owned, &source); err != nil {
			continue
		}
		result = append(result, Food{
			ID:     id,
			Name:   fname,
			Grams:  float32(grams),
			Macros: macrosByGrams(makeMPG(ppg, fpg, cpg, fibpg), 100),
			Owned:  owned == 1,
			Source: source,
		})
	}
	return result
}

func UpsertSharedFood(name string, fat float64, carb float64, fiber float64, protein float64, grams float64, ownerUserID int, source string, sourceRef string) error {
	source = strings.TrimSpace(source)
	sourceRef = strings.TrimSpace(sourceRef)
	if source == "" || sourceRef == "" {
		return fmt.Errorf("shared food source and source_ref are required")
	}

	tx, err := sqlDB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var existingID int
	err = tx.QueryRow(`SELECT id FROM foods WHERE source = ? AND source_ref = ? LIMIT 1`, source, sourceRef).Scan(&existingID)
	switch err {
	case nil:
		_, err = tx.Exec(
			`UPDATE foods
			SET name = ?, protein_per_gram = ?, fat_per_gram = ?, carb_per_gram = ?, fiber_per_gram = ?, grams = ?, creator_user_id = ?, barcode = '', is_shared = 1
			WHERE id = ?`,
			name, protein/grams, fat/grams, carb/grams, fiber/grams, grams, ownerUserID, existingID,
		)
		if err != nil {
			return fmt.Errorf("UpsertSharedFood update: %w", err)
		}
	case sql.ErrNoRows:
		_, err = tx.Exec(
			`INSERT INTO foods (name, protein_per_gram, fat_per_gram, carb_per_gram, fiber_per_gram, grams, creator_user_id, barcode, is_shared, source, source_ref)
			VALUES (?, ?, ?, ?, ?, ?, ?, '', 1, ?, ?)`,
			name, protein/grams, fat/grams, carb/grams, fiber/grams, grams, ownerUserID, source, sourceRef,
		)
		if err != nil {
			return fmt.Errorf("UpsertSharedFood insert: %w", err)
		}
	default:
		return fmt.Errorf("UpsertSharedFood lookup: %w", err)
	}

	return tx.Commit()
}
