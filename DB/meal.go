package database

import (
	"log"
	"time"
)

const mealTable string = `
CREATE TABLE IF NOT EXISTS "Meals" (
	"meal_id"	INTEGER NOT NULL UNIQUE,
	"user_id"	INTEGER NOT NULL,
	"meal_name"	TEXT NOT NULL DEFAULT 'snack',
	"meal_date"	date DEFAULT CURRENT_DATE NOT NULL,
	PRIMARY KEY("meal_id" AUTOINCREMENT),
	FOREIGN KEY("user_id") REFERENCES "Users"("user_id")
);
`

type Meal struct {
	Name  string
	ID    string
	Foods []Join
}

type Join struct {
	Macros Macro
	Name   string
	JoinID int
	Grams  float32
}

type Food struct {
	Macros Macro
	Name   string
	ID     int
	Grams  float32
}

func CreateMeal(name string, userID int) int {
	start := time.Now()
	defer func() {
		log.Println("Create Meal:", time.Since(start))
	}()
	today := time.Now().Format("2006-01-02")
	result, err := Db.Exec(`INSERT INTO Meals(meal_name, user_id, meal_date) VALUES(?, ?, ?);`, name, userID, today)
	if err != nil {
		panic(err.Error())
	}
	id, err := result.LastInsertId()
	if err != nil {
		panic(err.Error())
	}

	return int(id)
}

func DeleteMeal(mealID string) {
	// NOTE: you can't delete multiple things with a single query
	_, err := Db.Exec(`
	BEGIN TRANSACTION;

	-- First delete the related records in the MealFoods table
	DELETE FROM MealFoods WHERE meal_id = ?;

	-- TIhen delete the meal itself
	DELETE FROM Meals WHERE meal_id = ?;

	COMMIT;`, mealID, mealID)
	if err != nil {
		panic(err.Error())
	}
}

func GetMealByID(mealID string) Meal {
	result, err := Db.Query(
		`SELECT f.food_name, f.fat_per_gram, f.protein_per_gram, f.carbs_per_gram, 
		f.fiber_per_gram, mf.grams, mf.join_id, m.meal_name FROM Foods f
		JOIN MealFoods mf ON f.food_id = mf.food_id
		JOIN Meals m ON mf.meal_id = m.meal_id
		WHERE mf.meal_id = ?`, mealID)
	if err != nil {
		panic(err.Error())
	}
	defer result.Close()

	var model Meal
	model.ID = mealID
	for result.Next() {
		var m MacroPerGram
		var j Join
		result.Scan(&j.Name, &m.FatPerGram, &m.ProteinPerGram, &m.CarbPerGram, &m.FiberPerGram, &j.Grams, &j.JoinID, &model.Name)
		j.Macros = macrosByGrams(m, j.Grams)
		model.Foods = append(model.Foods, j)
	}

	if err = result.Err(); err != nil {
		panic(err.Error())
	}
	return model
}

func UpdateMealName(mealID string, name string) {
	_, err := Db.Exec(`UPDATE Meals SET meal_name = ? WHERE meal_id = ?`, name, mealID)
	if err != nil {
		panic(err.Error())
	}
}
