package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	COLLECTION = "users"
)

var (
	d  *DAO
	db *mongo.Database
)

type Credentials struct {
	Username string `bson:"username" json:"username"`
	Password string `bson:"password" json:"password"`
}

type DAO struct {
	Server   string
	Database string
}

// Connection to database
func (d *DAO) Connection() {
	fmt.Println("Connecting to MongoDB...")
	clientOptions := options.Client().ApplyURI(d.Server)
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	err = client.Connect(ctx)
	defer cancel()
	if err != nil {
		log.Fatal(err)
	}
	// Check the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to MongoDB!")
	db = client.Database(d.Database)
}
