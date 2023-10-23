package database

import (
	"fmt"
)

const templateJoinTable string = `
CREATE TABLE IF NOT EXISTS "TemplateFoods" (
    "join_id"	INTEGER NOT NULL UNIQUE,
    "template_id"	INTEGER NOT NULL,
    "food_id"	INTEGER NOT NULL,
    "grams"		FLOAT NOT NULL,
    PRIMARY KEY("join_id" AUTOINCREMENT),
    FOREIGN KEY("template_id") REFERENCES "Templates"("template_id"),
    FOREIGN KEY("food_id") REFERENCES "Foods"("food_id")
);
`

func CreateTemplateJoin(templateID string, foodID string, grams string) {
	result, err := Db.Exec(
		`INSERT INTO TemplateFoods(template_id, food_id, grams)
		VALUES(?,?,?);`, templateID, foodID, grams)
	if err != nil {
		panic(err.Error())
	}
	id, err := result.LastInsertId()
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("Created Template Join: ", id)
}

func DeleteTemplateJoin(joinID string) {
	_, err := Db.Exec(`DELETE FROM TemplateFoods WHERE join_id = ?;`, joinID)
	if err != nil {
		panic(err.Error())
	}
}

func UpdateTemplateJoin(joinID string, gramStr string) Join {
	_, err := Db.Exec(`UPDATE TemplateFoods SET grams = ? WHERE join_id = ?`, gramStr, joinID)
	if err != nil {
		panic(err.Error())
	}

	result := Db.QueryRow(`
		SELECT f.food_name, f.fat_per_gram, f.protein_per_gram, f.carbs_per_gram,
		 f.fiber_per_gram, t.grams, t.join_id FROM Foods f
 		JOIN TemplateFoods t ON f.food_id = t.food_id
 		WHERE t.join_id = ?`, joinID)

	var m MacroPerGram
	var j Join
	result.Scan(&j.Name, &m.FatPerGram, &m.ProteinPerGram, &m.CarbPerGram, &m.FiberPerGram, &j.Grams, &j.JoinID)
	j.Macros = macrosByGrams(m, j.Grams)
	if err = result.Err(); err != nil {
		panic(err.Error())
	}

	return j
}
