package main

import (
	"flag"
	"fmt"
	db "myapp/DB"
	"myapp/auth"
	"os"
	"strconv"
)

// TODO: Convert to flag, not sure how to do w/ fly.io yet.
var PROD = (os.Getenv("IS_PROD") != "")
var PROD_TURSO = false
var PROD_REDIS = false

// go run . --clearDB --populateDB
// go run
func init() {
	populateDB := flag.Bool("populateDB", false, "a bool")
	turso := flag.Bool("turso", false, "a bool")
	clearDB := flag.Bool("clearDB", false, "a bool")
	flag.Parse()

	fmt.Println("Populating DB: ", *populateDB)

	if PROD {
		fmt.Println("Running in prod")
		PROD_TURSO = true
		PROD_REDIS = true
	} else {
		fmt.Println("Running in local")
	}

	if *turso {
		PROD_TURSO = true
	}

	fmt.Println("Starting server:")
	auth.InitSession(PROD_REDIS)
	db.CreateOrOpenDatabase(PROD_TURSO)
	if *clearDB {
		db.ClearTables()
	}
	db.CreateTables()

	if *populateDB {
		// TODO: clear DB
		all_str, _ := db.CreateUser("All", "all")
		all64, _ := strconv.ParseInt(all_str, 10, 64)
		nico_str, _ := db.CreateUser("Nico", "123")
		nico64, _ := strconv.ParseInt(nico_str, 10, 64)
		all := int(all64)
		nico := int(nico64)
		db.CreateUser("foo", "123")

		banana := db.CreateFood("Banana", .4, 27, 3, 1.3, 118, all)
		oats := db.CreateFood("Oats", 2.5, 23, 4, 6, 35, nico)
		chia := db.CreateFood("Chia", 1.7, 1.7, 1.7, 1, 5, all)
		hemp := db.CreateFood("Hemp Hearts", 5, .7, .3, 1, 10, all)
		flax := db.CreateFood("Flax", 7, 4, 3, 4, 15, all)
		raisins := db.CreateFood("Golden Raisins", 0, 12, .8, .4, 15, all)
		goji := db.CreateFood("Goji", .5, 9.6, 1.6, 1.6, 15, all)
		yogurt := db.CreateFood("Greek Yogurt 2%", 3.6, 7, 0, 16, 170, all)
		blackberry := db.CreateFood("Blackberry", .2, 3.9, 2.2, .6, 120, all)
		butter := db.CreateFood("Butter", 11, 0, 0, 0, 14, all)
		quinoa := db.CreateFood("Quinoa", 12.0, 136.0, 16.0, 20.0, 180.0, all)
		pinkLentils := db.CreateFood("Pink Lentils", 0, 63, 11, 22, 100, all)
		egg := db.CreateFood("Egg", 4, 0, 0, 6, 44, all)
		chicken := db.CreateFood("Chicken [Uncooked]", 1, 0, 0, 24, 112, all)
		olives := db.CreateFood("Olive Salad", 5, 0, 0, 0, 30, all)
		tuna := db.CreateFood("Tuna [Canned]", .5, 0, 0, 26, 85, all)
		tomato := db.CreateFood("Tomato [Canned]", 0, 5, 1, 1, 120, all)
		Huel := db.CreateFood("Huel [Meal]", 9, 59, 6, 24, 101, nico)
		Pbar := db.CreateFood("Perfect bar", 20, 24, 4, 15, 65, nico)
		coconut := db.CreateFood("Coconut water", 0, 23, 0, 0, 90, nico)
		blackLentils := db.CreateFood("Black Lentils", 0, 25, 10, 12, 50, nico)
		broccoli := db.CreateFood("Broccoli", .4, 7, 2.6, 2.8, 100, nico)
		cauliflower := db.CreateFood("Cauliflower", .3, 5, 2, 1.9, 100, nico)
		shiitake := db.CreateFood("shiitake/maitake", .5, 7, 2.5, 2.2, 100, nico)
		oliveOil := db.CreateFood("Extra Virgin Olive Oil", 9.3, 0, 0, 0, 5, nico)
		protein := db.CreateFood("Protein Poweder", 2.5, 3, 2, 20, 32.7, all)
		chocolate := db.CreateFood("Chocolate hazelnut butter bar", 15, 14, 3, 3, 30, all)
		BrazilNuts := db.CreateFood("Brazil Nuts", 19, 4, 2, 4, 30, all)
		sunflower := db.CreateFood("Sunflower Lecithin", 5, .5, 0, 0, 10, all)

		// m1ID := db.CreateMeal("m1", nico)
		// db.CreateMealJoin(fmt.Sprint(m1ID), fmt.Sprint(banana), "20.5")
		// m2ID := db.CreateMeal("m2", nico)
		// db.CreateMealJoin(fmt.Sprint(m2ID), fmt.Sprint(oats), "50")
		// m3ID := db.CreateMeal("Dinner", nico)
		// db.CreateMealJoin(fmt.Sprint(m3ID), fmt.Sprint(oats), "101.3")
		// m4ID := db.CreateMeal("Dinner", nico2)
		// db.CreateMealJoin(fmt.Sprint(m4ID), fmt.Sprint(banana), "50.6")
		// db.CreateMealJoin(fmt.Sprint(m4ID), fmt.Sprint(oats), "101.3")
		fmt.Println(banana, oats, chia, hemp, flax, raisins, goji, yogurt, blackberry,
			butter, quinoa, pinkLentils, chicken, olives, tuna, tomato, Huel, Pbar, coconut,
			blackLentils, broccoli, cauliflower, shiitake, oliveOil, protein, chocolate)
		t2 := db.CreateTemplate("Coffee", nico)
		db.CreateTemplateJoin(fmt.Sprint(t2), fmt.Sprint(butter), "10")
		// t3 := db.CreateTemplate("Breakfast Oats", nico)
		// db.CreateTemplateJoin(fmt.Sprint(t3), fmt.Sprint(banana), "120")
		// db.CreateTemplateJoin(fmt.Sprint(t3), fmt.Sprint(oats), "100")
		// db.CreateTemplateJoin(fmt.Sprint(t3), fmt.Sprint(chia), "5")
		// db.CreateTemplateJoin(fmt.Sprint(t3), fmt.Sprint(hemp), "10")
		// db.CreateTemplateJoin(fmt.Sprint(t3), fmt.Sprint(flax), "15")
		// db.CreateTemplateJoin(fmt.Sprint(t3), fmt.Sprint(raisins), "30")
		// db.CreateTemplateJoin(fmt.Sprint(t3), fmt.Sprint(goji), "15")
		// db.CreateTemplateJoin(fmt.Sprint(t3), fmt.Sprint(yogurt), "255")
		// db.CreateTemplateJoin(fmt.Sprint(t3), fmt.Sprint(blackberry), "40")
		// t4 := db.CreateTemplate("Quinoa bowl - 1", nico)
		// db.CreateTemplateJoin(fmt.Sprint(t4), fmt.Sprint(quinoa), "180")
		// db.CreateTemplateJoin(fmt.Sprint(t4), fmt.Sprint(egg), "88")
		// db.CreateTemplateJoin(fmt.Sprint(t4), fmt.Sprint(olives), "70")
		// t5 := db.CreateTemplate("Lentil bowl", nico)
		// db.CreateTemplateJoin(fmt.Sprint(t5), fmt.Sprint(egg), "88")
		// db.CreateTemplateJoin(fmt.Sprint(t5), fmt.Sprint(tuna), "128")
		// db.CreateTemplateJoin(fmt.Sprint(t5), fmt.Sprint(pinkLentils), "200")
		// db.CreateTemplateJoin(fmt.Sprint(t5), fmt.Sprint(tomato), "120")

		// t6 := db.CreateTemplate("Breakfast Oats - Lean", nico)
		// db.CreateTemplateJoin(fmt.Sprint(t6), fmt.Sprint(oats), "80")
		// db.CreateTemplateJoin(fmt.Sprint(t6), fmt.Sprint(chia), "5")
		// db.CreateTemplateJoin(fmt.Sprint(t6), fmt.Sprint(hemp), "10")
		// db.CreateTemplateJoin(fmt.Sprint(t6), fmt.Sprint(flax), "15")
		// db.CreateTemplateJoin(fmt.Sprint(t6), fmt.Sprint(raisins), "30")
		// db.CreateTemplateJoin(fmt.Sprint(t6), fmt.Sprint(goji), "15")
		// db.CreateTemplateJoin(fmt.Sprint(t6), fmt.Sprint(yogurt), "255")
		// db.CreateTemplateJoin(fmt.Sprint(t6), fmt.Sprint(blackberry), "70")

		// t7 := db.CreateTemplate("Quinoa bowl - 2", nico)
		// db.CreateTemplateJoin(fmt.Sprint(t7), fmt.Sprint(quinoa), "180")
		// db.CreateTemplateJoin(fmt.Sprint(t7), fmt.Sprint(egg), "88")
		// db.CreateTemplateJoin(fmt.Sprint(t7), fmt.Sprint(chicken), "200")

		t8 := db.CreateTemplate("Huel", nico)
		db.CreateTemplateJoin(fmt.Sprint(t8), fmt.Sprint(Huel), "101")
		db.CreateTemplateJoin(fmt.Sprint(t8), fmt.Sprint(egg), "88")
		db.CreateTemplateJoin(fmt.Sprint(t8), fmt.Sprint(egg), "88")
		db.CreateTemplateJoin(fmt.Sprint(t8), fmt.Sprint(oliveOil), "10")

		// t9 := db.CreateTemplate("bar", nico)
		// db.CreateTemplateJoin(fmt.Sprint(t9), fmt.Sprint(Pbar), "65")

		t10 := db.CreateTemplate("Super Veggie", nico)
		db.CreateTemplateJoin(fmt.Sprint(t10), fmt.Sprint(blackLentils), "45")
		db.CreateTemplateJoin(fmt.Sprint(t10), fmt.Sprint(broccoli), "250")
		db.CreateTemplateJoin(fmt.Sprint(t10), fmt.Sprint(cauliflower), "150")
		db.CreateTemplateJoin(fmt.Sprint(t10), fmt.Sprint(shiitake), "50")
		db.CreateTemplateJoin(fmt.Sprint(t10), fmt.Sprint(oliveOil), "30")

		t11 := db.CreateTemplate("Cheat meal 1", nico)
		db.CreateTemplateJoin(fmt.Sprint(t11), fmt.Sprint(banana), "110")
		db.CreateTemplateJoin(fmt.Sprint(t11), fmt.Sprint(banana), "110")
		db.CreateTemplateJoin(fmt.Sprint(t11), fmt.Sprint(protein), "32.7")
		db.CreateTemplateJoin(fmt.Sprint(t11), fmt.Sprint(chocolate), "30")
		db.CreateTemplateJoin(fmt.Sprint(t11), fmt.Sprint(coconut), "90")

		t12 := db.CreateTemplate("Breafast bowl", nico)
		db.CreateTemplateJoin(fmt.Sprint(t12), fmt.Sprint(BrazilNuts), "15")
		db.CreateTemplateJoin(fmt.Sprint(t12), fmt.Sprint(sunflower), "5")
		db.CreateTemplateJoin(fmt.Sprint(t12), fmt.Sprint(chia), "5")
		db.CreateTemplateJoin(fmt.Sprint(t12), fmt.Sprint(hemp), "10")
		db.CreateTemplateJoin(fmt.Sprint(t12), fmt.Sprint(flax), "15")
		db.CreateTemplateJoin(fmt.Sprint(t12), fmt.Sprint(raisins), "30")
		db.CreateTemplateJoin(fmt.Sprint(t12), fmt.Sprint(goji), "15")
		db.CreateTemplateJoin(fmt.Sprint(t12), fmt.Sprint(yogurt), "255")
		db.CreateTemplateJoin(fmt.Sprint(t12), fmt.Sprint(blackberry), "70")

	}
}
