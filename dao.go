package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Connection to database
func (d *DAO) Connection() error {
	fmt.Printf("Connecting to SQLite DB at %s\n", d.DatabaseSourceName)

	_, err := os.Stat(d.DatabaseSourceName)
	if errors.Is(err, os.ErrNotExist) {
		// Create SQLite file since it does not exist
		file, err := os.Create(
			d.DatabaseSourceName)
		if err != nil {
			log.Println("Error creating database file:", err)
			return err
		}
		file.Close()

	} else if err != nil {
		log.Println("Error checking database file:", err)
		return err
	}

	// Open the created SQLite File
	d.DB, err = sql.Open(
		"sqlite3",
		d.DatabaseSourceName,
	)
	if err != nil {
		log.Println("Error opening database connection:", err)
		return err
	}

	// Check the connection
	err = d.DB.Ping()
	if err != nil {
		log.Println("Error pinging database:", err)
		return err
	}
	fmt.Println("Connected to SQLite DB!")
	fmt.Println("Creating User Table unless it exists...")

	// SQL Statement for Create Table
	createUserTableSQL := `CREATE TABLE IF NOT EXISTS user (
		"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,        
		"username" TEXT,
		"password" TEXT     
	  );`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Prepare SQL Statement
	statement, err := d.DB.PrepareContext(ctx, createUserTableSQL)
	if err != nil {
		log.Println("Error preparing SQL statement:", err)
		return err
	}
	defer statement.Close()

	// Execute SQL Statements
	_, err = statement.Exec()
	if err != nil {
		log.Println("Error creating table:", err)
		return err
	}

	return nil
}
