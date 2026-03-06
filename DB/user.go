package database

import (
	"fmt"
	"strconv"

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

	mu.Lock()
	defer mu.Unlock()

	// Check for duplicate username
	for _, u := range users {
		if u.Username == userName {
			return "", fmt.Errorf("username already taken")
		}
	}

	id := nextUserID
	nextUserID++
	users[id] = &UserRecord{
		ID:             id,
		Username:       userName,
		HashedPassword: hashedPassword,
		Targets:        defaultTargets,
	}
	fmt.Println("Created User: ", userName)
	return strconv.Itoa(id), nil
}

func GetUserTargets(userID int) Macro {
	mu.Lock()
	defer mu.Unlock()

	if u, ok := users[userID]; ok {
		return u.Targets
	}
	return defaultTargets
}

func UpdateUserTargets(userID int, targets Macro) {
	mu.Lock()
	defer mu.Unlock()

	if u, ok := users[userID]; ok {
		u.Targets = targets
	}
}

func ValidateUser(userName string, pass string) (string, error) {
	mu.Lock()
	defer mu.Unlock()

	for _, u := range users {
		if u.Username == userName {
			if !validateHashPassword(u.HashedPassword, pass) {
				return "", fmt.Errorf("Invalid username or password")
			}
			return strconv.Itoa(u.ID), nil
		}
	}
	return "", fmt.Errorf("Invalid username or password")
}
