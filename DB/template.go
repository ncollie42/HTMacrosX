package database

import (
	"fmt"
	"log"
	"time"
)

const templateTable string = `
CREATE TABLE IF NOT EXISTS "Templates" (
    "template_id"	INTEGER NOT NULL UNIQUE,
    "user_id"	INTEGER NOT NULL,
    "template_name"	TEXT NOT NULL,
    PRIMARY KEY("template_id" AUTOINCREMENT),
    FOREIGN KEY("user_id") REFERENCES "Users"("user_id")
);
`

func CreateTemplate(name string, userID int) int {
	result, err := Db.Exec(
		`INSERT INTO Templates(template_name, user_id) VALUES(?, ?);`, name, userID)
	if err != nil {
		panic(err.Error())
	}
	id, err := result.LastInsertId()
	fmt.Println("Created Template: ", name)
	if err != nil {
		panic(err.Error())
	}

	return int(id)
}

func GetTemplateEntriess(userID int) []MacroOverview {
	result, err := Db.Query(
		`SELECT f.fat_per_gram, f.protein_per_gram, f.carbs_per_gram,
		 f.fiber_per_gram, j.grams, j.template_id, t.template_name
	FROM Foods f
	JOIN TemplateFoods j ON f.food_id = j.food_id
	JOIN Templates t ON j.template_id= t.template_id
	WHERE t.user_id = ?;`, userID)
	if err != nil {
		panic(err.Error())
	}
	defer result.Close()

	var meals []MacroOverview
	for result.Next() {
		var macro MacroPerGram
		var grams float32
		var m MacroOverview
		result.Scan(&macro.FatPerGram, &macro.ProteinPerGram, &macro.CarbPerGram, &macro.FiberPerGram, &grams, &m.ID, &m.Name)
		m.Macros = macrosByGrams(macro, grams)
		meals = append(meals, m)
	}
	if err = result.Err(); err != nil {
		panic(err.Error())
	}
	return meals
}

func getTemplateFoodsByID(templateID string) []Food {
	result, err := Db.Query(
		`SELECT f.food_id, j.grams
	FROM Foods f
	JOIN TemplateFoods j ON f.food_id = j.food_id
	JOIN Templates t ON j.template_id= t.template_id
	WHERE t.template_id = ?;`, templateID)
	if err != nil {
		panic(err.Error())
	}
	defer result.Close()

	var foods []Food
	for result.Next() {
		var f Food
		result.Scan(&f.ID, &f.Grams)
		foods = append(foods, f)
	}
	if err = result.Err(); err != nil {
		panic(err.Error())
	}
	return foods
}

func TemplateToMeal(templateID string, userID int) int {
	start := time.Now()
	defer func() {
		log.Println("Template to Meal:", time.Since(start))
	}()

	tx, err := Db.Begin()
	if err != nil {
		panic(err.Error())
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	foods := getTemplateFoodsByID(templateID)
	mealID := CreateMeal(getTemplateName(templateID), userID)

	stmt := CreateMealJoinPrep()
	defer stmt.Close()
	for _, f := range foods {
		// CreateMealJoin(fmt.Sprint(mealID), fmt.Sprint(f.ID), fmt.Sprint(f.Grams))
		stmt.Exec(fmt.Sprint(mealID), fmt.Sprint(f.ID), fmt.Sprint(f.Grams))
	}

	err = tx.Commit()
	if err != nil {
		panic(err.Error())
	}
	return mealID
}

func GetTemplateByID(templateID string) Meal {
	result, err := Db.Query(`SELECT f.food_name, f.fat_per_gram, 
			f.protein_per_gram, f.carbs_per_gram, f.fiber_per_gram, j.grams, j.join_id, t.template_name FROM Foods f
		JOIN TemplateFoods j ON f.food_id = j.food_id
		JOIN Templates t ON j.template_id = t.template_id
		WHERE t.template_id = ?`, templateID)
	if err != nil {
		panic(err.Error())
	}
	defer result.Close()

	var meal Meal
	meal.ID = templateID
	for result.Next() {
		var m MacroPerGram
		var j Join
		result.Scan(&j.Name, &m.FatPerGram, &m.ProteinPerGram, &m.CarbPerGram, &m.FiberPerGram, &j.Grams, &j.JoinID, &meal.Name)
		j.Macros = macrosByGrams(m, j.Grams)
		meal.Foods = append(meal.Foods, j)
	}

	if err = result.Err(); err != nil {
		panic(err.Error())
	}
	return meal
}

func DeleteTemplate(templateID string) {
	// NOTE: you can't delete multiple things with a single query
	_, err := Db.Exec(`
	BEGIN TRANSACTION;

	-- First delete the related records in the MealFoods table
	DELETE FROM TemplateFoods WHERE template_id = ?;

	-- TIhen delete the meal itself
	DELETE FROM Templates WHERE template_id = ?;

	COMMIT;`, templateID, templateID)
	if err != nil {
		panic(err.Error())
	}
}

func UpdateTemplateName(templateID string, name string) {
	_, err := Db.Exec(`UPDATE Templates SET template_name = ? WHERE template_id = ?`, name, templateID)
	if err != nil {
		panic(err.Error())
	}
}

func getTemplateName(templateID string) string {
	result := Db.QueryRow("SELECT template_name FROM Templates WHERE template_id = ?", templateID)
	var name string
	result.Scan(&name)
	return name
}
