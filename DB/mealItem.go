package database

import (
	"strconv"
)

func CreateMealItem(mealID string, foodID string, grams string) {
	mid, _ := strconv.Atoi(mealID)
	fid, _ := strconv.Atoi(foodID)
	g, _ := strconv.ParseFloat(grams, 64)
	sqlDB.Exec(`INSERT INTO meal_items (meal_id, food_id, grams) VALUES (?, ?, ?)`, mid, fid, g)
}

func UpdateMealItem(joinID string, gramStr string) {
	jid, _ := strconv.Atoi(joinID)
	g, _ := strconv.ParseFloat(gramStr, 64)
	sqlDB.Exec(`UPDATE meal_items SET grams = ? WHERE id = ?`, g, jid)
}

func DeleteMealItem(joinID string) {
	jid, _ := strconv.Atoi(joinID)
	sqlDB.Exec(`DELETE FROM meal_items WHERE id = ?`, jid)
}
