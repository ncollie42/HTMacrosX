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

	id := nextJoinID
	nextJoinID++
	joins[id] = &JoinRecord{
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

	j, ok := joins[jid]
	if !ok {
		return Join{}
	}
	j.Grams = float32(g)

	food, ok := foods[j.FoodID]
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
		JoinID: j.ID,
		Grams:  j.Grams,
		Macros: macrosByGrams(mpg, j.Grams),
	}
}

func DeleteMealJoin(joinID string) {
	mu.Lock()
	defer mu.Unlock()

	jid, _ := strconv.Atoi(joinID)
	delete(joins, jid)
}
