package database

import (
	"fmt"
	"strconv"
)

func CreateTemplate(name string, userID int) int {
	mu.Lock()
	defer mu.Unlock()

	id := nextTemplateID
	nextTemplateID++
	templates[id] = &TemplateRecord{
		ID:     id,
		UserID: userID,
		Name:   name,
	}
	fmt.Println("Created Template: ", name)
	return id
}

func GetTemplateEntriess(userID int) []MacroOverview {
	mu.Lock()
	defer mu.Unlock()

	var results []MacroOverview
	for _, tj := range templateJoins {
		tmpl, tmplOk := templates[tj.TemplateID]
		food, foodOk := foods[tj.FoodID]
		if !tmplOk || !foodOk {
			continue
		}
		if tmpl.UserID != userID {
			continue
		}
		mpg := MacroPerGram{
			FatPerGram:     food.FatPerGram,
			ProteinPerGram: food.ProteinPerGram,
			CarbPerGram:    food.CarbPerGram,
			FiberPerGram:   food.FiberPerGram,
		}
		results = append(results, MacroOverview{
			Macros: macrosByGrams(mpg, tj.Grams),
			Name:   tmpl.Name,
			ID:     tmpl.ID,
		})
	}
	return results
}

func getTemplateFoodsByID(templateID string) []Food {
	id, _ := strconv.Atoi(templateID)
	var result []Food
	for _, tj := range templateJoins {
		if tj.TemplateID != id {
			continue
		}
		result = append(result, Food{
			ID:    tj.FoodID,
			Grams: tj.Grams,
		})
	}
	return result
}

func TemplateToMeal(templateID string, userID int) int {
	// Note: lock is handled by the called functions
	tmplFoods := getTemplateFoodsByID(templateID)
	tmplName := getTemplateName(templateID)
	mealID := CreateMeal(tmplName, userID)

	for _, f := range tmplFoods {
		CreateMealJoin(fmt.Sprint(mealID), fmt.Sprint(f.ID), fmt.Sprint(f.Grams))
	}
	return mealID
}

func GetTemplateByID(templateID string) Meal {
	mu.Lock()
	defer mu.Unlock()

	id, _ := strconv.Atoi(templateID)
	var meal Meal
	meal.ID = templateID

	tmpl, ok := templates[id]
	if !ok {
		return meal
	}
	meal.Name = tmpl.Name

	for _, tj := range templateJoins {
		if tj.TemplateID != id {
			continue
		}
		food, ok := foods[tj.FoodID]
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
			JoinID: tj.ID,
			Grams:  tj.Grams,
			Macros: macrosByGrams(mpg, tj.Grams),
		}
		meal.Foods = append(meal.Foods, j)
	}
	return meal
}

func DeleteTemplate(templateID string) {
	mu.Lock()
	defer mu.Unlock()

	id, _ := strconv.Atoi(templateID)
	for jid, tj := range templateJoins {
		if tj.TemplateID == id {
			delete(templateJoins, jid)
		}
	}
	delete(templates, id)
}

func UpdateTemplateName(templateID string, name string) {
	mu.Lock()
	defer mu.Unlock()

	id, _ := strconv.Atoi(templateID)
	if tmpl, ok := templates[id]; ok {
		tmpl.Name = name
	}
}

func getTemplateName(templateID string) string {
	mu.Lock()
	defer mu.Unlock()

	id, _ := strconv.Atoi(templateID)
	if tmpl, ok := templates[id]; ok {
		return tmpl.Name
	}
	return ""
}
