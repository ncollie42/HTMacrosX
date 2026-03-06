package database

import (
	"fmt"
	"strconv"
	"time"
)

func CreateMeal(name string, userID int) int {
	mu.Lock()
	defer mu.Unlock()

	today := time.Now().Format("2006-01-02")
	id := nextMealID
	nextMealID++
	meals[id] = &MealRecord{
		ID:       id,
		UserID:   userID,
		Name:     name,
		MealDate: today,
	}
	return id
}

func DeleteMeal(mealID string) {
	mu.Lock()
	defer mu.Unlock()

	id, _ := strconv.Atoi(mealID)

	// Delete associated joins
	for jid, mj := range mealJoins {
		if mj.MealID == id {
			delete(mealJoins, jid)
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

	for _, mj := range mealJoins {
		if mj.MealID != id {
			continue
		}
		food, ok := foods[mj.FoodID]
		if !ok {
			continue
		}
		mpg := MacroPerGram{
			FatPerGram:     food.FatPerGram,
			ProteinPerGram: food.ProteinPerGram,
			CarbPerGram:    food.CarbPerGram,
			FiberPerGram:   food.FiberPerGram,
		}
		j := Join{
			Name:   food.Name,
			JoinID: mj.ID,
			Grams:  mj.Grams,
			Macros: macrosByGrams(mpg, mj.Grams),
		}
		model.Foods = append(model.Foods, j)
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
