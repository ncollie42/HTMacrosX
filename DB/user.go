package database

import (
	"fmt"
	"strconv"

	"golang.org/x/crypto/bcrypt"
)

// ----------------  User  --------------------------
const userTable string = `
CREATE TABLE IF NOT EXISTS "Users" (
	"user_id"	INTEGER UNIQUE,
	"username"	VARCHAR(255) NOT NULL UNIQUE,
	"password"	VARCHAR(255) NOT NULL,
	"date_create"	DATETIME DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY("user_id" AUTOINCREMENT)
);
`

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
func validateHashPassword(hashedPassword string, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return false
	}
	return true
}

func CreateUser(userName string, pass string) (string, error) {
	hashedPassword, err := hashPassword(pass)

	if err != nil {
		return "", err
	}

	result, err := Db.Exec(`INSERT INTO Users (username, password) VALUES (?,?);`, userName, hashedPassword)
	if err != nil {
		return "", err
	}
	fmt.Println("Created User: ", userName)
	ID, err := result.LastInsertId()
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(ID, 10), nil
}

func ValidateUser(userName string, pass string) (string, error) {
	result := Db.QueryRow("SELECT user_id, password FROM Users WHERE username = ?", userName)

	var ID int64
	var hash string
	err := result.Scan(&ID, &hash)
	if err != nil {
		return "", err
	}

	if !validateHashPassword(hash, pass) {
		return "", fmt.Errorf("Invalid username or password")
	}
	return strconv.FormatInt(ID, 10), err
}
