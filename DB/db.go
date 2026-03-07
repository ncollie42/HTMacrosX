package database

import (
	"database/sql"
	"time"

	_ "modernc.org/sqlite"
)

var sqlDB *sql.DB

const schema = `
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    hashed_password TEXT NOT NULL,
    target_calories REAL NOT NULL DEFAULT 1751.6,
    target_fat REAL NOT NULL DEFAULT 44.8,
    target_carb REAL NOT NULL DEFAULT 247.1,
    target_fiber REAL NOT NULL DEFAULT 32.0,
    target_protein REAL NOT NULL DEFAULT 90.0
);

CREATE TABLE IF NOT EXISTS foods (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    protein_per_gram REAL NOT NULL,
    fat_per_gram REAL NOT NULL,
    carb_per_gram REAL NOT NULL,
    fiber_per_gram REAL NOT NULL,
    grams REAL NOT NULL,
    creator_user_id INTEGER NOT NULL,
    barcode TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS meals (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    meal_date TEXT NOT NULL DEFAULT '',
    is_preset INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS joins (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    meal_id INTEGER NOT NULL,
    food_id INTEGER NOT NULL,
    grams REAL NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_joins_meal_id ON joins(meal_id);
CREATE INDEX IF NOT EXISTS idx_meals_user_date ON meals(user_id, meal_date);
CREATE INDEX IF NOT EXISTS idx_foods_barcode ON foods(barcode);
CREATE INDEX IF NOT EXISTS idx_foods_creator ON foods(creator_user_id);
`

func Open(path string) {
	var err error
	sqlDB, err = sql.Open("sqlite", path)
	if err != nil {
		panic(err)
	}
	sqlDB.SetMaxOpenConns(1)
	if _, err = sqlDB.Exec(schema); err != nil {
		panic(err)
	}
}

func makeMPG(ppg, fpg, cpg, fibpg float64) MacroPerGram {
	return MacroPerGram{
		ProteinPerGram: float32(ppg),
		FatPerGram:     float32(fpg),
		CarbPerGram:    float32(cpg),
		FiberPerGram:   float32(fibpg),
	}
}

func scanMacroOverviewRows(rows *sql.Rows) []MacroOverview {
	var results []MacroOverview
	for rows.Next() {
		var grams, ppg, fpg, cpg, fibpg float64
		var name string
		var id int
		if err := rows.Scan(&grams, &name, &id, &ppg, &fpg, &cpg, &fibpg); err != nil {
			continue
		}
		results = append(results, MacroOverview{
			Macros: macrosByGrams(makeMPG(ppg, fpg, cpg, fibpg), float32(grams)),
			Name:   name,
			ID:     id,
		})
	}
	return results
}

// ----------------------------------------------
const ProteinKcalPerGram = 4
const FatKcalPerGram = 9
const CarbKcalPerGram = 4

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
}

type Join struct {
	Macros Macro
	Name   string
	JoinID int
	Grams  float32
}

type UserRecord struct {
	ID             int
	Username       string
	HashedPassword string
	Targets        Macro
}

type FoodRecord struct {
	ID             int
	Name           string
	ProteinPerGram float32
	FatPerGram     float32
	CarbPerGram    float32
	FiberPerGram   float32
	Grams          float32
	CreatorUserID  int
	Barcode        string
}

type MealRecord struct {
	ID       int
	UserID   int
	Name     string
	MealDate string
	IsPreset bool
}

type JoinRecord struct {
	ID     int
	MealID int
	FoodID int
	Grams  float32
}

func GetEntriessByDate(userID int, dateTime time.Time) []MacroOverview {
	date := dateTime.Format("2006-01-02")
	rows, err := sqlDB.Query(`
		SELECT j.grams, m.name, m.id, f.protein_per_gram, f.fat_per_gram, f.carb_per_gram, f.fiber_per_gram
		FROM joins j
		JOIN meals m ON m.id = j.meal_id
		JOIN foods f ON f.id = j.food_id
		WHERE m.user_id = ? AND m.meal_date = ? AND m.is_preset = 0
	`, userID, date)
	if err != nil {
		return nil
	}
	defer rows.Close()
	return scanMacroOverviewRows(rows)
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
	var order []int
	for _, m := range macros {
		if _, seen := dict[m.ID]; !seen {
			order = append(order, m.ID)
		}
		dict[m.ID] = append(dict[m.ID], m)
	}

	var newMacros []MacroOverview
	for _, id := range order {
		m := dict[id]
		newMacros = append(newMacros, MacroOverview{
			Macros: SumMacros(m),
			ID:     id,
			Name:   m[0].Name,
		})
	}
	return newMacros
}
