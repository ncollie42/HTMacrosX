package database

import (
	"fmt"
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

func hashPassword(pass string) string {
	// TODO: Hashpassword
	return pass
}

func CreateUser(userName string, pass string) (int, error) {
	hashedPassword := hashPassword(pass)

	result, err := Db.Exec(`INSERT INTO Users (username, password) VALUES (?,?);`, userName, hashedPassword)
	if err != nil {
		return 0, err
	}
	fmt.Println("Created User: ", userName)
	ID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(ID), nil
}

func ValidateUser(userName string, pass string) (int, error) {
	hashedPassword := hashPassword(pass)

	result := Db.QueryRow("SELECT user_id FROM Users WHERE username = ? AND password = ?", userName, hashedPassword)
	ID := 0
	err := result.Scan(&ID)

	return ID, err
}
