package database

import (
	"sync"
	"time"
)

var mu sync.Mutex

// Auto-increment counters
var nextUserID = 1
var nextFoodID = 1
var nextMealID = 1
var nextMealJoinID = 1
var nextTemplateID = 1
var nextTemplateJoinID = 1

// In-memory stores
var users = map[int]*UserRecord{}
var foods = map[int]*FoodRecord{}
var meals = map[int]*MealRecord{}
var mealJoins = map[int]*MealJoinRecord{}
var templates = map[int]*TemplateRecord{}
var templateJoins = map[int]*TemplateJoinRecord{}

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
}

type MealRecord struct {
	ID       int
	UserID   int
	Name     string
	MealDate string
}

type MealJoinRecord struct {
	ID     int
	MealID int
	FoodID int
	Grams  float32
}

type TemplateRecord struct {
	ID     int
	UserID int
	Name   string
}

type TemplateJoinRecord struct {
	ID         int
	TemplateID int
	FoodID     int
	Grams      float32
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

func GetEntriessByDate(userID int, dateTime time.Time) []MacroOverview {
	date := dateTime.Format("2006-01-02")
	mu.Lock()
	defer mu.Unlock()

	var results []MacroOverview
	for _, mj := range mealJoins {
		meal, mealOk := meals[mj.MealID]
		food, foodOk := foods[mj.FoodID]
		if !mealOk || !foodOk {
			continue
		}
		if meal.UserID != userID || meal.MealDate != date {
			continue
		}
		mpg := MacroPerGram{
			FatPerGram:     food.FatPerGram,
			ProteinPerGram: food.ProteinPerGram,
			CarbPerGram:    food.CarbPerGram,
			FiberPerGram:   food.FiberPerGram,
		}
		results = append(results, MacroOverview{
			Macros: macrosByGrams(mpg, mj.Grams),
			Name:   meal.Name,
			ID:     meal.ID,
		})
	}
	return results
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
