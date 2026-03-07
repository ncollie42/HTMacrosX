package main

import (
	"fmt"
	"myapp/auth"
	db "myapp/DB"
	"os"
)

func init() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./app.db"
	}
	db.Open(dbPath)
	fmt.Println("Starting server with SQLite:", dbPath)
	auth.InitSession()

	seedUser("All", "all")
	seedUser("Nico", "123")
	seedUser("Alejandro", "123")
	seedUser("foo", "123")
}

func seedUser(username, password string) {
	db.CreateUser(username, password)
}
