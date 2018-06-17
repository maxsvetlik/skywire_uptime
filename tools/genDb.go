package main

import (
	"fmt"
	"skywire_uptime/database"
)

func main() {
	dbName := "./database/testing.db"
	fmt.Println("Generating new database")
	generateNewDatabase(dbName)
}

// Generate a database file
func generateNewDatabase(name string) {
	db := database.ConnectDB(name)
	defer db.Close()
	db.SetupDb()
}
