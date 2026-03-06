package database

import (
	"fmt"
	"sort"
	"strings"
)

func CreateFood(name string, fat float64, carb float64, fiber float64, protein float64, grams float64, userID int) int {
	return CreateFoodWithBarcode(name, fat, carb, fiber, protein, grams, userID, "")
}

func CreateFoodWithBarcode(name string, fat float64, carb float64, fiber float64, protein float64, grams float64, userID int, barcode string) int {
	mu.Lock()
	defer mu.Unlock()

	id := nextFoodID
	nextFoodID++
	foods[id] = &FoodRecord{
		ID:             id,
		Name:           name,
		ProteinPerGram: float32(protein / grams),
		FatPerGram:     float32(fat / grams),
		CarbPerGram:    float32(carb / grams),
		FiberPerGram:   float32(fiber / grams),
		Grams:          float32(grams),
		CreatorUserID:  userID,
		Barcode:        barcode,
	}
	fmt.Println("Created Food: ", name)
	return id
}

func FindFoodByBarcode(barcode string) *FoodRecord {
	mu.Lock()
	defer mu.Unlock()

	for _, f := range foods {
		if f.Barcode == barcode {
			return f
		}
	}
	return nil
}

func FoodSearch(name string, userID int) []Food {
	mu.Lock()
	defer mu.Unlock()

	search := strings.ToLower(name)
	var result []Food
	for _, f := range foods {
		if f.CreatorUserID != 1 && f.CreatorUserID != userID {
			continue
		}
		if search != "" && !strings.Contains(strings.ToLower(f.Name), search) {
			continue
		}
		mpg := MacroPerGram{
			ProteinPerGram: f.ProteinPerGram,
			FatPerGram:     f.FatPerGram,
			CarbPerGram:    f.CarbPerGram,
			FiberPerGram:   f.FiberPerGram,
		}
		result = append(result, Food{
			ID:     f.ID,
			Name:   f.Name,
			Grams:  f.Grams,
			Macros: macrosByGrams(mpg, 100),
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}
