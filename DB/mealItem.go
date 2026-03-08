package database

func CreateMealItem(mealID int, foodID int, grams float64, userID int, isPreset bool) error {
	res, err := sqlDB.Exec(
		`INSERT INTO meal_items (meal_id, food_id, grams)
		SELECT ?, ?, ?
		WHERE EXISTS (SELECT 1 FROM meals WHERE id = ? AND user_id = ? AND is_preset = ?)
		  AND EXISTS (SELECT 1 FROM foods WHERE id = ? AND (creator_user_id = ? OR creator_user_id = ?))`,
		mealID, foodID, grams, mealID, userID, presetInt(isPreset), foodID, userID, SystemUserID,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotOwned
	}
	return nil
}

func UpdateMealItem(mealID int, itemID int, grams float64, userID int, isPreset bool) error {
	res, err := sqlDB.Exec(
		`UPDATE meal_items
		SET grams = ?
		WHERE id = ?
		  AND meal_id = ?
		  AND meal_id IN (SELECT id FROM meals WHERE id = ? AND user_id = ? AND is_preset = ?)`,
		grams, itemID, mealID, mealID, userID, presetInt(isPreset),
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotOwned
	}
	return nil
}

func DeleteMealItem(mealID int, itemID int, userID int, isPreset bool) error {
	res, err := sqlDB.Exec(
		`DELETE FROM meal_items
		WHERE id = ?
		  AND meal_id = ?
		  AND meal_id IN (SELECT id FROM meals WHERE id = ? AND user_id = ? AND is_preset = ?)`,
		itemID, mealID, mealID, userID, presetInt(isPreset),
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotOwned
	}
	return nil
}
