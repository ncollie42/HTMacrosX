package usda

import (
	"encoding/json"
	"strconv"
	"strings"
)

type Food struct {
	FDCID   string
	Name    string
	Protein float64
	Fat     float64
	Carb    float64
	Fiber   float64
}

func ParseFoundationFoods(raw []byte) ([]Food, error) {
	var payload struct {
		FoundationFoods []struct {
			FDCID                 int    `json:"fdcId"`
			Description           string `json:"description"`
			IsHistoricalReference bool   `json:"isHistoricalReference"`
			FoodNutrients         []struct {
				Amount   float64 `json:"amount"`
				Nutrient struct {
					Number string `json:"number"`
				} `json:"nutrient"`
			} `json:"foodNutrients"`
		} `json:"FoundationFoods"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}

	foods := make([]Food, 0, len(payload.FoundationFoods))
	for _, item := range payload.FoundationFoods {
		if item.IsHistoricalReference {
			continue
		}
		name := strings.TrimSpace(item.Description)
		if name == "" {
			continue
		}

		var (
			protein float64
			fat     float64
			carb    float64
			fiber   float64
			hasP    bool
			hasF    bool
			hasC    bool
		)
		for _, nutrient := range item.FoodNutrients {
			switch nutrient.Nutrient.Number {
			case "203":
				protein = nutrient.Amount
				hasP = true
			case "204":
				fat = nutrient.Amount
				hasF = true
			case "205":
				carb = nutrient.Amount
				hasC = true
			case "291":
				fiber = nutrient.Amount
			}
		}
		if !hasP || !hasF || !hasC {
			continue
		}

		foods = append(foods, Food{
			FDCID:   strconv.Itoa(item.FDCID),
			Name:    name,
			Protein: protein,
			Fat:     fat,
			Carb:    carb,
			Fiber:   fiber,
		})
	}
	return foods, nil
}
