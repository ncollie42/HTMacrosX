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
}
