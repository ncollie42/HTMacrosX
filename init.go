package main

import (
	"fmt"
	db "myapp/DB"
	"myapp/auth"
)

func init() {
	fmt.Println("Running with in-memory storage")
	fmt.Println("Starting server:")
	auth.InitSession()

	{
		all_str, _ := db.CreateUser("All", "all")
		nico_str, _ := db.CreateUser("Nico", "123")
		alejandro_str, _ := db.CreateUser("Alejandro", "123")
		db.CreateUser("foo", "123")

		// Parse user IDs
		all := atoiSafe(all_str)
		nico := atoiSafe(nico_str)
		alejandro := atoiSafe(alejandro_str)

		//All
		banana := db.CreateFood("Banana", .4, 27, 3, 1.3, 118, all)
		oats := db.CreateFood("Oats", 2.5, 23, 4, 6, 35, nico)
		chia := db.CreateFood("Chia", 1.7, 1.7, 1.7, 1, 5, all)
		hemp := db.CreateFood("Hemp Hearts", 5, .7, .3, 1, 10, all)
		flax := db.CreateFood("Flax", 7, 4, 3, 4, 15, all)
		raisins := db.CreateFood("Golden Raisins", 0, 12, .8, .4, 15, all)
		goji := db.CreateFood("Goji", .5, 9.6, 1.6, 1.6, 15, all)
		yogurt := db.CreateFood("Greek Yogurt 2%", 3.6, 7, 0, 16, 170, all)
		yogurtFatFree := db.CreateFood("Greek Yogurt 0%", 0, 7, 0, 18, 170, all)
		blackberry := db.CreateFood("Blackberry", .2, 3.9, 2.2, .6, 120, all)
		butter := db.CreateFood("Butter", 11, 0, 0, 0, 14, all)
		quinoa := db.CreateFood("Quinoa", 12.0, 136.0, 16.0, 20.0, 180.0, all)
		pinkLentils := db.CreateFood("Pink Lentils", 0, 63, 11, 22, 100, all)
		egg := db.CreateFood("Egg", 4, 0, 0, 6, 44, all)
		chicken := db.CreateFood("Chicken [Uncooked]", 1, 0, 0, 24, 112, all)
		olives := db.CreateFood("Olive Salad", 5, 0, 0, 0, 30, all)
		tuna := db.CreateFood("Tuna [Canned]", 1, 0, 0, 29, 113, all)
		tomato := db.CreateFood("Tomato [Canned]", 0, 5, 1, 1, 120, all)
		Huel := db.CreateFood("Huel [Meal]", 9, 59, 6, 24, 101, nico)
		Pbar := db.CreateFood("Perfect bar", 20, 24, 4, 15, 65, nico)
		coconut := db.CreateFood("Coconut water", 0, 23, 0, 0, 90, nico)
		blackLentils := db.CreateFood("Black Lentils", 0, 25, 10, 12, 50, nico)
		broccoli := db.CreateFood("Broccoli", .4, 7, 2.6, 2.8, 100, nico)
		cauliflower := db.CreateFood("Cauliflower", .3, 5, 2, 1.9, 100, nico)
		shiitake := db.CreateFood("shiitake/maitake", .5, 7, 2.5, 2.2, 100, nico)
		oliveOil := db.CreateFood("Extra Virgin Olive Oil", 14, 0, 0, 0, 15, nico)
		protein := db.CreateFood("Protein Poweder", 2.5, 3, 2, 20, 32.7, all)
		chocolate := db.CreateFood("Chocolate hazelnut butter bar", 15, 14, 3, 3, 30, all)
		BrazilNuts := db.CreateFood("Brazil Nuts", 19, 4, 2, 4, 30, all)
		Walnuts := db.CreateFood("Walnuts", 18, 4, 2, 4, 28, all)
		sunflower := db.CreateFood("Sunflower Lecithin", 5, .5, 0, 0, 10, all)
		kefir := db.CreateFood("Kefir", 2, 8, 0, 8, 240, all)
		//Alejandro
		pasta := db.CreateFood("Pasta", .5, 40, 2, 7, 56, all)
		BBSauce := db.CreateFood("BB Sauce", 0, 5, 0, 0, 32, all)
		granollaV := db.CreateFood("Granolla V", 6, 23, 2, 3, 35, all)
		granollaC := db.CreateFood("Granolla C", 6, 18, 2, 2, 30, all)
		honey := db.CreateFood("Honey", 0, 18, 0, 0, 21, all)

		fmt.Println(banana, oats, chia, hemp, flax, raisins, goji, yogurt, blackberry,
			butter, quinoa, pinkLentils, chicken, olives, tuna, tomato, Huel, Pbar, coconut,
			blackLentils, broccoli, cauliflower, shiitake, oliveOil, protein, chocolate, kefir)

		t2 := db.CreateTemplate("Coffee", nico)
		db.CreateTemplateJoin(fmt.Sprint(t2), fmt.Sprint(butter), "10")

		t8 := db.CreateTemplate("Huel", nico)
		db.CreateTemplateJoin(fmt.Sprint(t8), fmt.Sprint(Huel), "101")
		db.CreateTemplateJoin(fmt.Sprint(t8), fmt.Sprint(egg), "88")
		db.CreateTemplateJoin(fmt.Sprint(t8), fmt.Sprint(oliveOil), "30")

		t10 := db.CreateTemplate("Super Veggie", nico)
		db.CreateTemplateJoin(fmt.Sprint(t10), fmt.Sprint(blackLentils), "50")
		db.CreateTemplateJoin(fmt.Sprint(t10), fmt.Sprint(broccoli), "250")
		db.CreateTemplateJoin(fmt.Sprint(t10), fmt.Sprint(cauliflower), "150")
		db.CreateTemplateJoin(fmt.Sprint(t10), fmt.Sprint(shiitake), "50")
		db.CreateTemplateJoin(fmt.Sprint(t10), fmt.Sprint(oliveOil), "30")
		db.CreateTemplateJoin(fmt.Sprint(t10), fmt.Sprint(tuna), "113")

		t12 := db.CreateTemplate("Breafast bowl", nico)
		db.CreateTemplateJoin(fmt.Sprint(t12), fmt.Sprint(BrazilNuts), "15")
		db.CreateTemplateJoin(fmt.Sprint(t12), fmt.Sprint(Walnuts), "15")
		db.CreateTemplateJoin(fmt.Sprint(t12), fmt.Sprint(sunflower), "5")
		db.CreateTemplateJoin(fmt.Sprint(t12), fmt.Sprint(chia), "5")
		db.CreateTemplateJoin(fmt.Sprint(t12), fmt.Sprint(hemp), "10")
		db.CreateTemplateJoin(fmt.Sprint(t12), fmt.Sprint(flax), "15")
		db.CreateTemplateJoin(fmt.Sprint(t12), fmt.Sprint(raisins), "30")
		db.CreateTemplateJoin(fmt.Sprint(t12), fmt.Sprint(goji), "15")
		db.CreateTemplateJoin(fmt.Sprint(t12), fmt.Sprint(yogurtFatFree), "255")
		db.CreateTemplateJoin(fmt.Sprint(t12), fmt.Sprint(blackberry), "70")

		a1 := db.CreateTemplate("pasta", alejandro)
		db.CreateTemplateJoin(fmt.Sprint(a1), fmt.Sprint(pasta), "156")
		db.CreateTemplateJoin(fmt.Sprint(a1), fmt.Sprint(BBSauce), "106")
		db.CreateTemplateJoin(fmt.Sprint(a1), fmt.Sprint(butter), "23.3")
		db.CreateTemplateJoin(fmt.Sprint(a1), fmt.Sprint(chicken), "378")

		a2 := db.CreateTemplate("Yogurt Bowl", alejandro)
		db.CreateTemplateJoin(fmt.Sprint(a2), fmt.Sprint(yogurtFatFree), "300")
		db.CreateTemplateJoin(fmt.Sprint(a2), fmt.Sprint(granollaV), "100")
		db.CreateTemplateJoin(fmt.Sprint(a2), fmt.Sprint(granollaC), "50")
		db.CreateTemplateJoin(fmt.Sprint(a2), fmt.Sprint(honey), "10")
	}
}

func atoiSafe(s string) int {
	var n int
	fmt.Sscanf(s, "%d", &n)
	return n
}
