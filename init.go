package main

import (
	"fmt"
	"myapp/auth"
	db "myapp/DB"
)

func init() {
	fmt.Println("Running with in-memory storage")
	fmt.Println("Starting server:")
	auth.InitSession()

	db.CreateUser("All", "all")
	db.CreateUser("Nico", "123")
	db.CreateUser("Alejandro", "123")
	db.CreateUser("foo", "123")
}
