package database

import (
	"database/sql"
	"fmt"
	"os"

	"time"
	// _ "github.com/mattn/go-sqlite3"    		  //cgo sql3
	_ "github.com/libsql/libsql-client-go/libsql" //Turso
	_ "modernc.org/sqlite"                        //no cgo sql3
)

var Db *sql.DB = nil
var localDB = "t.db"

func ClearTables() {
	fmt.Println("Clearing DB:")
	// sqlite_sequence
	names := []string{"MealFoods", "TemplateFoods", "Templates", "Meals", "Foods", "Users"}
	// Db.Exec(`SELECT name from sqlite_master where type="table";`)

	for _, name := range names {
		fmt.Println("Droping table", name)
		sql := fmt.Sprintf("DROP TABLE IF EXISTS %s", name)
		if _, err := Db.Exec(sql); err != nil {
			panic("Error with Clearing table: " + name + " | " + err.Error())
		}
	}
}

func CreateTables() {
	if _, err := Db.Exec(userTable); err != nil {
		panic("Error Exec sql - create user\n" + err.Error())
	}
	if _, err := Db.Exec(foodTable); err != nil {
		panic("Error Exec sql - create food\n" + err.Error())
	}
	if _, err := Db.Exec(mealTable); err != nil {
		panic("Error Exec sql - create meal\n" + err.Error())
	}
	if _, err := Db.Exec(mealJoinTable); err != nil {
		panic("Error Exec sql - create meal join\n" + err.Error())
	}
	if _, err := Db.Exec(templateTable); err != nil {
		panic("Error Exec sql - create template\n" + err.Error())
	}
	if _, err := Db.Exec(templateJoinTable); err != nil {
		panic("Error Exec sql - create template join\n" + err.Error())
	}
}

func tursoURL() string {
	token := os.Getenv("TURSO_TOKEN")
	db := os.Getenv("TURSO_URL")

	fmt.Println("Connecting to Turso:")
	fmt.Println("	URL:", db)
	fmt.Println("	Token:", token)
	URL := fmt.Sprint(db, "?authToken=", token)
	return URL
}

func createOrOpenDatabaseTurso() {
	dbUrl := tursoURL()
	db, err := sql.Open("libsql", dbUrl) //Turso - Prod
	if err != nil {
		panic("Failed to open or create DB\n" + err.Error())
	}
	Db = db
	// NOTE: Journal mode is WAL by default with turso
}

func createOrOpenDatabaseLocal() {
	// NOTE: "sqlite3" for cgo and "sqlite" for no cgo lib.
	// db, err := sql.Open("libsql", fmt.Sprint("file:", file)) //Turso - Local
	db, err := sql.Open("sqlite", localDB) //Local
	if err != nil {
		panic("Failed to open or create DB\n" + err.Error())
	}
	Db = db
	db.Exec("PRAGMA journal_mode=WAL;")
}

func CreateOrOpenDatabase(prod bool) {
	fmt.Println("Creating / Opening Database")

	if prod {
		createOrOpenDatabaseTurso()
	} else {
		createOrOpenDatabaseLocal()
	}
}

// ----------------------------------------------
const ProteinKcalPerGram = 4
const FatKcalPerGram = 9
const CarbKcalPerGram = 4
const AlcoholKcalPerGram = 4

type Macro struct {
	Calories float32
	Fat      float32
	Carb     float32
	Fiber    float32
	Protein  float32
}

type MacroPerGram struct {
	ProteinPerGram float32
	FatPerGram     float32
	CarbPerGram    float32
	FiberPerGram   float32
}

// TODO: Merge structs in a way that makes sense.
// NOTE: used for Meal - Food - Template
type MacroOverview struct {
	Macros Macro
	Name   string
	ID     int
}

type Meal struct {
	Name  string
	ID    string
	Foods []Join
}

type Food struct {
	Macros Macro
	Name   string
	ID     int
	Grams  float32
	//PreferedServing
}

type Join struct {
	Macros Macro
	Name   string
	JoinID int
	Grams  float32
}

func GetEntriessByDate(userID int, dateTime time.Time) []MacroOverview {
	date := dateTime.Format("2006-01-02")
	result, err := Db.Query(
		`SELECT f.fat_per_gram, f.protein_per_gram, f.carbs_per_gram, f.fiber_per_gram, j.grams, j.meal_id, m.meal_name
	FROM Foods f
	JOIN MealFoods j ON f.food_id = j.food_id
	JOIN Meals m ON j.meal_id = m.meal_id
	JOIN users u ON u.user_id = m.user_id
	WHERE m.meal_date = ? AND u.user_id = ?;`, date, userID)
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

func macrosByGrams(macro MacroPerGram, grams float32) Macro {
	gramsOfProtein := macro.ProteinPerGram * grams
	protein := gramsOfProtein * ProteinKcalPerGram
	gramsOfFat := macro.FatPerGram * grams
	fat := gramsOfFat * FatKcalPerGram
	gramsOfCarb := macro.CarbPerGram * grams
	carb := gramsOfCarb * CarbKcalPerGram
	gramsOfFiber := macro.FiberPerGram * grams
	return Macro{
		Calories: protein + fat + carb,
		Protein:  gramsOfProtein,
		Fat:      gramsOfFat,
		Carb:     gramsOfCarb,
		Fiber:    gramsOfFiber,
	}
}

func SumMacros(macros []MacroOverview) Macro {
	var macro Macro

	for _, m := range macros {
		macro.Calories += m.Macros.Calories
		macro.Protein += m.Macros.Protein
		macro.Fat += m.Macros.Fat
		macro.Carb += m.Macros.Carb
		macro.Fiber += m.Macros.Fiber
	}
	return macro
}

func SumMacrosByID(macros []MacroOverview) []MacroOverview {
	dict := map[int][]MacroOverview{}
	for _, m := range macros {
		dict[m.ID] = append(dict[m.ID], m)
	}

	var newMacros []MacroOverview
	for id, m := range dict {
		var newMacro MacroOverview
		newMacro.Macros = SumMacros(m)
		newMacro.ID = id
		newMacro.Name = dict[id][0].Name

		newMacros = append(newMacros, newMacro)
	}
	return newMacros
}
