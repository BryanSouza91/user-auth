package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

const (
	COLLECTION = "users"
)

var (
	d *DAO
)

// Connection to database
func (d *DAO) Connection() {
	fmt.Printf("Connecting to SQLite DB at %s", d.DatabaseSourceName)

	fsInfo, err := os.Stat(d.DatabaseSourceName)
	if errors.Is(err, os.ErrNotExist) {
		// Create SQLite file since it does not exist
		file, err := os.Create(
			d.DatabaseSourceName)
		if err != nil {
			log.Fatal(err)
		}
		file.Close()

	} else {
		fmt.Println(err.Error())
		log.Fatal(err)
	}
	fmt.Printf("Fileinfo: %v", fsInfo)

	// Open the created SQLite File
	d.DB, err = sql.Open(
		"sqlite3",
		d.DatabaseSourceName,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer d.DB.Close() // Defer Closing the database

	// Check the connection
	err = d.DB.Ping()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to SQLite DB!")
	fmt.Println("Creating User Table...")

	// SQL Statement for Create Table
	createUserTableSQL := `CREATE TABLE user (
		"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,		
		"username" TEXT,
		"password" TEXT		
	  );`
	// Prepare SQL Statement
	statement, err := d.DB.Prepare(createUserTableSQL)
	if err != nil {
		log.Fatal(err.Error())
	}
	// Execute SQL Statements
	statement.Exec()
	log.Println("User table created successfully!")
}
