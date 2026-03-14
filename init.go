package main

import (
	"fmt"
	db "myapp/DB"
	"myapp/auth"
	"os"
)

func init() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./app.db"
	}
	db.Open(dbPath)
	syncUSDAFoundationFoodsOnStartup()
	fmt.Println("Starting server with SQLite:", dbPath)
	auth.InitSession()
}
