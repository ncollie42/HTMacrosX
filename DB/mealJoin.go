package database

import (
	"database/sql"
	"log"
	"time"
)

const mealJoinTable string = `
CREATE TABLE IF NOT EXISTS "MealFoods" (
	"join_id"	INTEGER NOT NULL UNIQUE,
	"meal_id"	INTEGER NOT NULL,
	"food_id"	INTEGER NOT NULL,
	"grams"		FLOAT NOT NULL,
	PRIMARY KEY("join_id" AUTOINCREMENT),
	FOREIGN KEY("meal_id") REFERENCES "Meals"("meal_id"),
	FOREIGN KEY("food_id") REFERENCES "Foods"("food_id")
);
`

func CreateMealJoin(mealID string, foodID string, grams string) {
	start := time.Now()
	defer func() {
		log.Println("Meal Join:", time.Since(start))
	}()
	_, err := Db.Exec(
		`INSERT INTO MealFoods(meal_id, food_id, grams)
		VALUES(?,?,?);`, mealID, foodID, grams)
	if err != nil {
		panic(err.Error())
	}
}

func CreateMealJoinPrep() *sql.Stmt {
	stmt, err := Db.Prepare(
		`INSERT INTO MealFoods(meal_id, food_id, grams)
		VALUES(?,?,?);`)
	if err != nil {
		panic(err.Error())
	}
	return stmt
}

func UpdateMealJoin(joinID string, gramStr string) Join {
	_, err := Db.Exec(`UPDATE MealFoods SET grams = ? WHERE join_id = ?`, gramStr, joinID)
	if err != nil {
		panic(err.Error())
	}

	result := Db.QueryRow(`
		SELECT f.food_name, f.fat_per_gram, f.protein_per_gram, f.carbs_per_gram, f.fiber_per_gram, mf.grams, mf.join_id FROM Foods f
 		JOIN MealFoods mf ON f.food_id = mf.food_id
 		WHERE mf.join_id = ?`, joinID)

	var m MacroPerGram
	var j Join
	result.Scan(&j.Name, &m.FatPerGram, &m.ProteinPerGram, &m.CarbPerGram, &m.FiberPerGram, &j.Grams, &j.JoinID)
	j.Macros = macrosByGrams(m, j.Grams)
	if err = result.Err(); err != nil {
		panic(err.Error())
	}

	return j
}

func DeleteMealJoin(joinID string) {
	_, err := Db.Exec(`DELETE FROM MealFoods WHERE join_id = ?;`, joinID)
	if err != nil {
		panic(err.Error())
	}
}
