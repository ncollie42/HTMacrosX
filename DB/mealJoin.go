package database

import (
	"strconv"
)

func CreateMealJoin(mealID string, foodID string, grams string) {
	mid, _ := strconv.Atoi(mealID)
	fid, _ := strconv.Atoi(foodID)
	g, _ := strconv.ParseFloat(grams, 64)
	sqlDB.Exec(`INSERT INTO joins (meal_id, food_id, grams) VALUES (?, ?, ?)`, mid, fid, g)
}

func UpdateMealJoin(joinID string, gramStr string) Join {
	jid, _ := strconv.Atoi(joinID)
	g, _ := strconv.ParseFloat(gramStr, 64)

	sqlDB.Exec(`UPDATE joins SET grams = ? WHERE id = ?`, g, jid)

	var fname string
	var ppg, fpg, cpg, fibpg float64
	err := sqlDB.QueryRow(`
		SELECT f.name, f.protein_per_gram, f.fat_per_gram, f.carb_per_gram, f.fiber_per_gram
		FROM joins j JOIN foods f ON f.id = j.food_id WHERE j.id = ?
	`, jid).Scan(&fname, &ppg, &fpg, &cpg, &fibpg)
	if err != nil {
		return Join{}
	}
	return Join{
		Name:   fname,
		JoinID: jid,
		Grams:  float32(g),
		Macros: macrosByGrams(makeMPG(ppg, fpg, cpg, fibpg), float32(g)),
	}
}

func DeleteMealJoin(joinID string) {
	jid, _ := strconv.Atoi(joinID)
	sqlDB.Exec(`DELETE FROM joins WHERE id = ?`, jid)
}
