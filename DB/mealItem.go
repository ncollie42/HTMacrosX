package database

func CreateMealItem(mealID int, foodID int, grams float64, userID int) error {
	res, err := sqlDB.Exec(
		`INSERT INTO meal_items (meal_id, food_id, grams) SELECT ?, ?, ? WHERE EXISTS (SELECT 1 FROM meals WHERE id = ? AND user_id = ?)`,
		mealID, foodID, grams, mealID, userID,
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

func UpdateMealItem(itemID int, grams float64, userID int) error {
	res, err := sqlDB.Exec(
		`UPDATE meal_items SET grams = ? WHERE id = ? AND meal_id IN (SELECT id FROM meals WHERE user_id = ?)`,
		grams, itemID, userID,
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

func DeleteMealItem(itemID int, userID int) error {
	res, err := sqlDB.Exec(
		`DELETE FROM meal_items WHERE id = ? AND meal_id IN (SELECT id FROM meals WHERE user_id = ?)`,
		itemID, userID,
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
