package database

import (
	"fmt"
	"strconv"
)

func CreateTemplateJoin(templateID string, foodID string, grams string) {
	mu.Lock()
	defer mu.Unlock()

	tid, _ := strconv.Atoi(templateID)
	fid, _ := strconv.Atoi(foodID)
	g, _ := strconv.ParseFloat(grams, 32)

	id := nextTemplateJoinID
	nextTemplateJoinID++
	templateJoins[id] = &TemplateJoinRecord{
		ID:         id,
		TemplateID: tid,
		FoodID:     fid,
		Grams:      float32(g),
	}
	fmt.Println("Created Template Join: ", id)
}

func DeleteTemplateJoin(joinID string) {
	mu.Lock()
	defer mu.Unlock()

	jid, _ := strconv.Atoi(joinID)
	delete(templateJoins, jid)
}

func UpdateTemplateJoin(joinID string, gramStr string) Join {
	mu.Lock()
	defer mu.Unlock()

	jid, _ := strconv.Atoi(joinID)
	g, _ := strconv.ParseFloat(gramStr, 32)

	tj, ok := templateJoins[jid]
	if !ok {
		return Join{}
	}
	tj.Grams = float32(g)

	food, ok := foods[tj.FoodID]
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
		JoinID: tj.ID,
		Grams:  tj.Grams,
		Macros: macrosByGrams(mpg, tj.Grams),
	}
}
