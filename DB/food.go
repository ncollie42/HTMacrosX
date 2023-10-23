package database

import (
	"fmt"
)

// ----------------  Food --------------------------
const foodTable string = `
CREATE TABLE IF NOT EXISTS "Foods" (
	"food_id"	INTEGER NOT NULL UNIQUE,
	"food_name"	VARCHAR(255) NOT NULL,
	"protein_per_gram"	FLOAT NOT NULL,
	"fat_per_gram"	FLOAT NOT NULL,
	"carbs_per_gram"	FLOAT NOT NULL,
	"fiber_per_gram"	FLOAT NOT NULL,
	"grams"	FLOAT NOT NULL,
	"creator_user_id" INTEGER NOT NULL,
	PRIMARY KEY("food_id" AUTOINCREMENT),
	FOREIGN KEY("creator_user_id") REFERENCES "Users"("user_id")
);
`

func CreateFood(name string, fat float64, carb float64, fiber float64, protein float64, grams float64, userID int) int {
	// NOTE: if food not created by user userID should be 0
	result, err := Db.Exec(`
		INSERT INTO Foods (food_name, protein_per_gram, fat_per_gram, carbs_per_gram, fiber_per_gram, grams, creator_user_id)
	VALUES(?, ?, ?, ?, ?, ?, ?);`, name, protein/grams, fat/grams, carb/grams, fiber/grams, grams, userID)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("Created Food: ", name)
	foodID, err := result.LastInsertId()
	if err != nil {
		panic(err)
	}
	return int(foodID)
}

func FoodSearch(name string, userID int) []Food {
	// NOTE: Empty string will return all rows, this can replace GetAllFoods
	// NOTE: user_id 1 is a shared profile.

	// Show all foods on base 100grams.
	result, err := Db.Query(`SELECT * FROM Foods WHERE food_name LIKE ? AND creator_user_id IN (1,?) `,
		fmt.Sprint("%", name, "%"), userID)
	if err != nil {
		panic(err.Error())
	}
	defer result.Close()

	var foods []Food
	for result.Next() {
		var m MacroPerGram
		var f Food
		var userID int
		result.Scan(&f.ID, &f.Name, &m.ProteinPerGram, &m.FatPerGram, &m.CarbPerGram, &m.FiberPerGram, &f.Grams, &userID)
		f.Macros = macrosByGrams(m, 100)
		foods = append(foods, f)
	}
	if err = result.Err(); err != nil {
		panic(err.Error())
	}
	return foods
}
