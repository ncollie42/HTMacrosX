package main

import (
	_ "embed"
	"log"
	db "myapp/DB"
	"myapp/usda"
)

const usdaSystemUsername = "__system_usda__"

//go:embed FoodData_Central_foundation_food_json_2025-12-18.json
var usdaFoundationFoodsJSON []byte

func syncUSDAFoundationFoods() error {
	systemUserID, err := db.EnsureSystemUser(usdaSystemUsername)
	if err != nil {
		return err
	}

	foods, err := usda.ParseFoundationFoods(usdaFoundationFoodsJSON)
	if err != nil {
		return err
	}
	for _, food := range foods {
		if err := db.UpsertSharedFood(food.Name, food.Fat, food.Carb, food.Fiber, food.Protein, 100, systemUserID, db.USDAFoundationSource, food.FDCID); err != nil {
			return err
		}
	}
	return nil
}

func syncUSDAFoundationFoodsOnStartup() {
	if err := syncUSDAFoundationFoods(); err != nil {
		log.Printf("sync USDA foundation foods: %v", err)
	}
}
