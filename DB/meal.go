package database

import (
	"fmt"
	"strconv"
	"time"
)

func CreateMeal(name string, userID int, isPreset bool) int {
	mu.Lock()
	defer mu.Unlock()

	today := ""
	if !isPreset {
		today = time.Now().Format("2006-01-02")
	}
	id := nextMealID
	nextMealID++
	meals[id] = &MealRecord{
		ID:       id,
		UserID:   userID,
		Name:     name,
		MealDate: today,
		IsPreset: isPreset,
	}
	return id
}

func DeleteMeal(mealID string) {
	mu.Lock()
	defer mu.Unlock()

	id, _ := strconv.Atoi(mealID)

	// Delete associated joins
	for jid, j := range joins {
		if j.MealID == id {
			delete(joins, jid)
		}
	}
	delete(meals, id)
}

func GetMealByID(mealID string) Meal {
	mu.Lock()
	defer mu.Unlock()

	id, _ := strconv.Atoi(mealID)
	var model Meal
	model.ID = mealID

	meal, ok := meals[id]
	if !ok {
		return model
	}
	model.Name = meal.Name

	for _, j := range joins {
		if j.MealID != id {
			continue
		}
		food, ok := foods[j.FoodID]
		if !ok {
			continue
		}
		mpg := MacroPerGram{
			FatPerGram:     food.FatPerGram,
			ProteinPerGram: food.ProteinPerGram,
			CarbPerGram:    food.CarbPerGram,
			FiberPerGram:   food.FiberPerGram,
		}
		jn := Join{
			Name:   food.Name,
			JoinID: j.ID,
			Grams:  j.Grams,
			Macros: macrosByGrams(mpg, j.Grams),
		}
		model.Foods = append(model.Foods, jn)
	}
	return model
}

func UpdateMealName(mealID string, name string) {
	mu.Lock()
	defer mu.Unlock()

	id, _ := strconv.Atoi(mealID)
	if meal, ok := meals[id]; ok {
		meal.Name = name
	}
	fmt.Println("Updated Meal Name:", mealID, name)
}

func GetPresetsEntries(userID int) []MacroOverview {
	mu.Lock()
	defer mu.Unlock()

	var results []MacroOverview
	for _, j := range joins {
		meal, mealOk := meals[j.MealID]
		food, foodOk := foods[j.FoodID]
		if !mealOk || !foodOk || !meal.IsPreset || meal.UserID != userID {
			continue
		}
		mpg := MacroPerGram{
			FatPerGram:     food.FatPerGram,
			ProteinPerGram: food.ProteinPerGram,
			CarbPerGram:    food.CarbPerGram,
			FiberPerGram:   food.FiberPerGram,
		}
		results = append(results, MacroOverview{
			Macros: macrosByGrams(mpg, j.Grams),
			Name:   meal.Name,
			ID:     meal.ID,
		})
	}
	return results
}

func TemplateToMeal(templateID string, userID int) int {
	id, _ := strconv.Atoi(templateID)

	mu.Lock()
	var presetName string
	type item struct {
		FoodID int
		Grams  float32
	}
	var items []item
	if p, ok := meals[id]; ok {
		presetName = p.Name
		for _, j := range joins {
			if j.MealID == id {
				items = append(items, item{j.FoodID, j.Grams})
			}
		}
	}
	mu.Unlock()

	mealID := CreateMeal(presetName, userID, false)
	for _, f := range items {
		CreateMealJoin(strconv.Itoa(mealID), strconv.Itoa(f.FoodID), fmt.Sprint(f.Grams))
	}
	return mealID
}
