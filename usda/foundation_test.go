package usda

import "testing"

func TestParseFoundationFoodsFiltersAndMapsNutrients(t *testing.T) {
	raw := []byte(`{
		"FoundationFoods": [
			{
				"fdcId": 1001,
				"description": "Banana, raw",
				"isHistoricalReference": false,
				"foodNutrients": [
					{"amount": 1.1, "nutrient": {"number": "203"}},
					{"amount": 0.3, "nutrient": {"number": "204"}},
					{"amount": 22.8, "nutrient": {"number": "205"}},
					{"amount": 2.6, "nutrient": {"number": "291"}}
				]
			},
			{
				"fdcId": 1002,
				"description": "Broccoli, raw",
				"isHistoricalReference": false,
				"foodNutrients": [
					{"amount": 2.8, "nutrient": {"number": "203"}},
					{"amount": 0.4, "nutrient": {"number": "204"}},
					{"amount": 7.0, "nutrient": {"number": "205"}}
				]
			},
			{
				"fdcId": 1003,
				"description": "",
				"isHistoricalReference": false,
				"foodNutrients": [
					{"amount": 1.0, "nutrient": {"number": "203"}},
					{"amount": 1.0, "nutrient": {"number": "204"}},
					{"amount": 1.0, "nutrient": {"number": "205"}}
				]
			},
			{
				"fdcId": 1004,
				"description": "Missing Carb",
				"isHistoricalReference": false,
				"foodNutrients": [
					{"amount": 1.0, "nutrient": {"number": "203"}},
					{"amount": 1.0, "nutrient": {"number": "204"}}
				]
			},
			{
				"fdcId": 1005,
				"description": "Old Entry",
				"isHistoricalReference": true,
				"foodNutrients": [
					{"amount": 1.0, "nutrient": {"number": "203"}},
					{"amount": 1.0, "nutrient": {"number": "204"}},
					{"amount": 1.0, "nutrient": {"number": "205"}}
				]
			}
		]
	}`)

	foods, err := ParseFoundationFoods(raw)
	if err != nil {
		t.Fatalf("ParseFoundationFoods returned error: %v", err)
	}
	if len(foods) != 2 {
		t.Fatalf("len = %d, want 2", len(foods))
	}

	if foods[0].FDCID != "1001" || foods[0].Name != "Banana, raw" {
		t.Fatalf("first food = %+v", foods[0])
	}
	if foods[0].Protein != 1.1 || foods[0].Fat != 0.3 || foods[0].Carb != 22.8 || foods[0].Fiber != 2.6 {
		t.Fatalf("banana macros = %+v", foods[0])
	}

	if foods[1].FDCID != "1002" || foods[1].Fiber != 0 {
		t.Fatalf("second food = %+v, want fiber defaulted to 0", foods[1])
	}
}
