package main

import "database/sql"

type Credentials struct {
	Username string `bson:"username" json:"username"`
	Password string `bson:"password" json:"password"`
}

type DAO struct {
	DatabaseSourceName string `json:"database"`
	DB                 *sql.DB
}
