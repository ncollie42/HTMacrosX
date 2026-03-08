package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

var sqlDB *sql.DB

var ErrNotOwned = fmt.Errorf("resource not found or not owned by user")
const DefaultSavedMealName = "Saved Meal"
const SharedFoodVisibility = 1

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
    grams REAL NOT NULL CHECK (grams > 0),
    creator_user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    barcode TEXT NOT NULL DEFAULT '',
    is_shared INTEGER NOT NULL DEFAULT 0 CHECK (is_shared IN (0, 1))
);

CREATE TABLE IF NOT EXISTS meals (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    meal_date TEXT NOT NULL DEFAULT '',
    is_preset INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS meal_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    meal_id INTEGER NOT NULL REFERENCES meals(id) ON DELETE CASCADE,
    food_id INTEGER NOT NULL REFERENCES foods(id) ON DELETE CASCADE,
    grams REAL NOT NULL CHECK (grams > 0)
);

CREATE TABLE IF NOT EXISTS sessions (
    session_id TEXT PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token TEXT NOT NULL DEFAULT '',
    expires_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_meal_items_meal_id ON meal_items(meal_id);
CREATE INDEX IF NOT EXISTS idx_meals_user_date ON meals(user_id, meal_date);
CREATE INDEX IF NOT EXISTS idx_foods_barcode ON foods(barcode);
CREATE INDEX IF NOT EXISTS idx_foods_creator ON foods(creator_user_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_foods_user_barcode ON foods(creator_user_id, barcode) WHERE barcode <> '';
CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at);
`

func Open(path string) {
	var err error
	sqlDB, err = sql.Open("sqlite", path)
	if err != nil {
		panic(err)
	}
	sqlDB.SetMaxOpenConns(1)
	if _, err = sqlDB.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		panic(err)
	}
	// Migrate: rename old "joins" table to "meal_items" if it exists
	sqlDB.Exec(`ALTER TABLE joins RENAME TO meal_items`)
	sqlDB.Exec(`DROP INDEX IF EXISTS idx_joins_meal_id`)
	if err := migrateLegacySchema(); err != nil {
		panic(err)
	}
	if _, err = sqlDB.Exec(schema); err != nil {
		panic(err)
	}
}

func migrateLegacySchema() error {
	hasFoods, err := tableExists("foods")
	if err != nil || !hasFoods {
		return err
	}
	hasShared, err := tableHasColumn("foods", "is_shared")
	if err != nil || hasShared {
		return err
	}

	tx, err := sqlDB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`PRAGMA foreign_keys = OFF`); err != nil {
		return err
	}

	renames := []string{
		`ALTER TABLE foods RENAME TO foods_legacy`,
		`ALTER TABLE meals RENAME TO meals_legacy`,
		`ALTER TABLE meal_items RENAME TO meal_items_legacy`,
		`ALTER TABLE sessions RENAME TO sessions_legacy`,
	}
	for _, stmt := range renames {
		if _, err := tx.Exec(stmt); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(schema); err != nil {
		return err
	}

	copyStmts := []string{
		`INSERT INTO foods (id, name, protein_per_gram, fat_per_gram, carb_per_gram, fiber_per_gram, grams, creator_user_id, barcode, is_shared)
		SELECT id, name, protein_per_gram, fat_per_gram, carb_per_gram, fiber_per_gram, grams, creator_user_id, barcode, 0
		FROM foods_legacy f
		WHERE EXISTS (SELECT 1 FROM users u WHERE u.id = f.creator_user_id)
		  AND grams > 0`,
		`INSERT INTO meals (id, user_id, name, meal_date, is_preset)
		SELECT id, user_id, name, meal_date, is_preset
		FROM meals_legacy m
		WHERE EXISTS (SELECT 1 FROM users u WHERE u.id = m.user_id)`,
		`INSERT INTO meal_items (id, meal_id, food_id, grams)
		SELECT mi.id, mi.meal_id, mi.food_id, mi.grams
		FROM meal_items_legacy mi
		WHERE mi.grams > 0
		  AND EXISTS (SELECT 1 FROM meals m WHERE m.id = mi.meal_id)
		  AND EXISTS (SELECT 1 FROM foods f WHERE f.id = mi.food_id)`,
		`INSERT INTO sessions (session_id, user_id, token, expires_at)
		SELECT session_id, user_id, token, expires_at
		FROM sessions_legacy s
		WHERE EXISTS (SELECT 1 FROM users u WHERE u.id = s.user_id)`,
	}
	for _, stmt := range copyStmts {
		if _, err := tx.Exec(stmt); err != nil {
			return err
		}
	}

	drops := []string{
		`DROP TABLE foods_legacy`,
		`DROP TABLE meals_legacy`,
		`DROP TABLE meal_items_legacy`,
		`DROP TABLE sessions_legacy`,
	}
	for _, stmt := range drops {
		if _, err := tx.Exec(stmt); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		return err
	}
	return tx.Commit()
}

func tableExists(name string) (bool, error) {
	var count int
	err := sqlDB.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?`, name).Scan(&count)
	return count > 0, err
}

func tableHasColumn(table string, column string) (bool, error) {
	rows, err := sqlDB.Query(`PRAGMA table_info(` + table + `)`)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name string
		var colType string
		var notNull int
		var defaultVal sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &colType, &notNull, &defaultVal, &pk); err != nil {
			return false, err
		}
		if name == column {
			return true, nil
		}
	}
	return false, rows.Err()
}

func makeMPG(ppg, fpg, cpg, fibpg float64) MacroPerGram {
	return MacroPerGram{
		ProteinPerGram: float32(ppg),
		FatPerGram:     float32(fpg),
		CarbPerGram:    float32(cpg),
		FiberPerGram:   float32(fibpg),
	}
}

func scanMealSummaryRows(rows *sql.Rows) []MealSummary {
	var results []MealSummary
	for rows.Next() {
		var grams, ppg, fpg, cpg, fibpg float64
		var name string
		var id int
		if err := rows.Scan(&grams, &name, &id, &ppg, &fpg, &cpg, &fibpg); err != nil {
			continue
		}
		results = append(results, MealSummary{
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

func CaloriesFromGrams(fat, carb, protein float64) float32 {
	return float32(fat*FatKcalPerGram + carb*CarbKcalPerGram + protein*ProteinKcalPerGram)
}

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

type MealSummary struct {
	Macros Macro
	Name   string
	ID     int
}

type Meal struct {
	Name  string
	ID    int
	Items []MealItem
}

type Food struct {
	Macros Macro
	Name   string
	ID     int
	Grams  float32
}

type MealItem struct {
	Macros Macro
	Name   string
	ItemID int
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

type MealItemRecord struct {
	ID     int
	MealID int
	FoodID int
	Grams  float32
}

const mealSummaryQuery = `
	SELECT COALESCE(j.grams, 0), m.name, m.id,
	       COALESCE(f.protein_per_gram, 0), COALESCE(f.fat_per_gram, 0),
	       COALESCE(f.carb_per_gram, 0), COALESCE(f.fiber_per_gram, 0)
	FROM meals m
	LEFT JOIN meal_items j ON j.meal_id = m.id
	LEFT JOIN foods f ON f.id = j.food_id`

func queryMealSummaries(where string, args ...any) []MealSummary {
	rows, err := sqlDB.Query(mealSummaryQuery+" "+where, args...)
	if err != nil {
		log.Printf("queryMealSummaries: %v", err)
		return nil
	}
	defer rows.Close()
	return scanMealSummaryRows(rows)
}

func GetMealItemsByDate(userID int, dateTime time.Time) []MealSummary {
	date := dateTime.Format("2006-01-02")
	return queryMealSummaries("WHERE m.user_id = ? AND m.meal_date = ? AND m.is_preset = 0", userID, date)
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

func SumMacros(macros []MealSummary) Macro {
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

func SumMealItemMacros(items []MealItem) Macro {
	var m Macro
	for _, item := range items {
		m.Calories += item.Macros.Calories
		m.Protein += item.Macros.Protein
		m.Fat += item.Macros.Fat
		m.Carb += item.Macros.Carb
		m.Fiber += item.Macros.Fiber
	}
	return m
}

func SumMacrosByID(macros []MealSummary) []MealSummary {
	dict := map[int][]MealSummary{}
	var order []int
	for _, m := range macros {
		if _, seen := dict[m.ID]; !seen {
			order = append(order, m.ID)
		}
		dict[m.ID] = append(dict[m.ID], m)
	}

	var newMacros []MealSummary
	for _, id := range order {
		m := dict[id]
		newMacros = append(newMacros, MealSummary{
			Macros: SumMacros(m),
			ID:     id,
			Name:   m[0].Name,
		})
	}
	return newMacros
}
