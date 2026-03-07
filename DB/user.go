package database

import (
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func validateHashPassword(hashedPassword string, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

var defaultTargets = Macro{Calories: 1751.6, Fat: 44.8, Carb: 247.1, Fiber: 32.0, Protein: 90.0}

func CreateUser(userName string, pass string) (string, error) {
	hashedPassword, err := hashPassword(pass)
	if err != nil {
		return "", err
	}

	res, err := sqlDB.Exec(
		`INSERT INTO users (username, hashed_password, target_calories, target_fat, target_carb, target_fiber, target_protein) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		userName, hashedPassword,
		defaultTargets.Calories, defaultTargets.Fat, defaultTargets.Carb, defaultTargets.Fiber, defaultTargets.Protein,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return "", fmt.Errorf("username already taken")
		}
		return "", err
	}
	id, _ := res.LastInsertId()
	fmt.Println("Created User:", userName)
	return strconv.Itoa(int(id)), nil
}

func GetUserTargets(userID int) Macro {
	var cal, fat, carb, fiber, protein float64
	err := sqlDB.QueryRow(
		`SELECT target_calories, target_fat, target_carb, target_fiber, target_protein FROM users WHERE id = ?`,
		userID,
	).Scan(&cal, &fat, &carb, &fiber, &protein)
	if err != nil {
		return defaultTargets
	}
	return Macro{
		Calories: float32(cal),
		Fat:      float32(fat),
		Carb:     float32(carb),
		Fiber:    float32(fiber),
		Protein:  float32(protein),
	}
}

func UpdateUserTargets(userID int, targets Macro) {
	sqlDB.Exec(
		`UPDATE users SET target_calories=?, target_fat=?, target_carb=?, target_fiber=?, target_protein=? WHERE id=?`,
		targets.Calories, targets.Fat, targets.Carb, targets.Fiber, targets.Protein, userID,
	)
}

func ValidateUser(userName string, pass string) (string, error) {
	var id int
	var hashedPassword string
	err := sqlDB.QueryRow(
		`SELECT id, hashed_password FROM users WHERE username = ?`,
		userName,
	).Scan(&id, &hashedPassword)
	if err != nil {
		return "", fmt.Errorf("Invalid username or password")
	}
	if !validateHashPassword(hashedPassword, pass) {
		return "", fmt.Errorf("Invalid username or password")
	}
	return strconv.Itoa(id), nil
}
