package main

import (
	"errors"
	db "myapp/DB"
	"myapp/foodsource"
	"net/http"
	"strconv"
	"time"
)

var httpClient = &http.Client{Timeout: 10 * time.Second}

var externalFoodProvider foodsource.Provider = &foodsource.OpenFoodFactsProvider{Client: httpClient}

func previewExternalFood(code string, userID int) (db.Food, error) {
	if f := db.FindFoodByBarcode(code, userID); f != nil {
		return foodFromRecord(f), nil
	}
	candidate, err := externalFoodProvider.LookupBarcode(code)
	if err != nil {
		return db.Food{}, err
	}
	return foodFromCandidate(candidate), nil
}

func addExternalFoodByTarget(userID int, targetType string, targetID string, code string, mealDate time.Time) (int, error) {
	if existing := db.FindFoodByBarcode(code, userID); existing != nil {
		return addFoodByTarget(userID, targetType, targetID, existing.ID, mealDate)
	}
	candidate, err := externalFoodProvider.LookupBarcode(code)
	if err != nil {
		return 0, err
	}
	return createAndAddFoodByTargetWithBarcode(userID, targetType, targetID, candidate.Name, candidate.Fat, candidate.Carb, candidate.Fiber, candidate.Protein, 100, candidate.Code, mealDate)
}

func foodFromRecord(record *db.FoodRecord) db.Food {
	return db.Food{
		ID:     record.ID,
		Name:   record.Name,
		Grams:  100,
		Source: record.Source,
		Macros: db.Macro{
			Protein:  record.ProteinPerGram * 100,
			Fat:      record.FatPerGram * 100,
			Carb:     record.CarbPerGram * 100,
			Fiber:    record.FiberPerGram * 100,
			Calories: record.ProteinPerGram*100*db.ProteinKcalPerGram + record.FatPerGram*100*db.FatKcalPerGram + record.CarbPerGram*100*db.CarbKcalPerGram,
		},
	}
}

func foodFromCandidate(candidate foodsource.Candidate) db.Food {
	return db.Food{
		Name:  candidate.Name,
		Grams: 100,
		Macros: db.Macro{
			Calories: db.CaloriesFromGrams(candidate.Fat, candidate.Carb, candidate.Protein),
			Fat:      float32(candidate.Fat),
			Carb:     float32(candidate.Carb),
			Fiber:    float32(candidate.Fiber),
			Protein:  float32(candidate.Protein),
		},
	}
}

func createAndAddFoodByTargetWithBarcode(userID int, targetType string, targetID string, name string, fat float64, carb float64, fiber float64, protein float64, grams float64, barcode string, mealDate time.Time) (int, error) {
	var mealID int
	var err error
	if targetType == "meal" && targetID == newMealParam {
		mealID, err = db.CreateMeal(defaultMealName, userID, false, mealDate)
		if err != nil {
			return 0, err
		}
	} else {
		mealID, err = strconv.Atoi(targetID)
		if err != nil {
			return 0, errors.New("invalid target")
		}
		if err := db.ValidateMealAccess(mealID, userID, targetType == "template"); err != nil {
			return 0, err
		}
	}
	if _, err := db.CreateFoodAndMealItem(name, fat, carb, fiber, protein, grams, userID, barcode, mealID, targetType == "template", 100); err != nil {
		return 0, err
	}
	return mealID, nil
}

func defaultTargetType(mealType string) string {
	if mealType == "template" {
		return "template"
	}
	return "meal"
}

func targetIDForMeal(mealID int) string {
	if mealID == 0 {
		return newMealParam
	}
	return strconv.Itoa(mealID)
}
