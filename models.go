package main

import "database/sql"

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type DAO struct {
	DatabaseSourceName string `json:"database"`
	DB                 *sql.DB
}
