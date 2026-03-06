package database

import (
	"strconv"
)

func CreateMealJoin(mealID string, foodID string, grams string) {
	mu.Lock()
	defer mu.Unlock()

	mid, _ := strconv.Atoi(mealID)
	fid, _ := strconv.Atoi(foodID)
	g, _ := strconv.ParseFloat(grams, 32)

	id := nextMealJoinID
	nextMealJoinID++
	mealJoins[id] = &MealJoinRecord{
		ID:     id,
		MealID: mid,
		FoodID: fid,
		Grams:  float32(g),
	}
}

func UpdateMealJoin(joinID string, gramStr string) Join {
	mu.Lock()
	defer mu.Unlock()

	jid, _ := strconv.Atoi(joinID)
	g, _ := strconv.ParseFloat(gramStr, 32)

	mj, ok := mealJoins[jid]
	if !ok {
		return Join{}
	}
	mj.Grams = float32(g)

	food, ok := foods[mj.FoodID]
	if !ok {
		return Join{}
	}

	mpg := MacroPerGram{
		FatPerGram:     food.FatPerGram,
		ProteinPerGram: food.ProteinPerGram,
		CarbPerGram:    food.CarbPerGram,
		FiberPerGram:   food.FiberPerGram,
	}
	return Join{
		Name:   food.Name,
		JoinID: mj.ID,
		Grams:  mj.Grams,
		Macros: macrosByGrams(mpg, mj.Grams),
	}
}

func DeleteMealJoin(joinID string) {
	mu.Lock()
	defer mu.Unlock()

	jid, _ := strconv.Atoi(joinID)
	delete(mealJoins, jid)
}
