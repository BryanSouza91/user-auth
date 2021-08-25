package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

var (
	dao  = DAO{}
	port *int
	err  error
)

// Parse the configuration file 'conf.json', and establish a connection to DB
func init() {
	port = flag.Int("port", 3000, "specified port")
	flag.Parse()
	file, err := os.Open("conf.json")
	if err != nil {
		log.Fatal("error:", err)
	}
	decoder := json.NewDecoder(file)
	defer file.Close()
	err = decoder.Decode(&dao)
	if err != nil {
		log.Fatal("error:", err)
	}

	dao.Connection()
}

// Define HTTP request routes
func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/signup", Signup)
	mux.HandleFunc("/signin", Signin)
	// mux.HandleFunc("/users", AllUsersEndpoint)
	// mux.HandleFunc("/users/new", CreateUserEndpoint)
	// mux.Handle("/users/update/", makeHandler(UpdateUserEndpoint))
	// mux.Handle("/users/delete/", makeHandler(DeleteUserEndpoint))
	// mux.Handle("/users/find/", makeHandler(FindUserEndpoint))
	fmt.Println(fmt.Sprintf("Listening on port %s", strconv.Itoa(*port)))
	if err = http.ListenAndServe(":"+strconv.Itoa(*port), mux); err != nil {
		log.Fatal(err)
	}
}

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

// Insert into database
func (d *DAO) Insert(creds Credentials) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	_, err = db.Collection(COLLECTION).InsertOne(ctx, &creds)
	if err != nil {
		log.Fatal(err)
	}
	return err
}

// FindByUsername will find Credentials by username
func (d *DAO) FindByUsername(username string) (creds Credentials, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	if err != nil {
		log.Fatal(err)
	}
	err = db.Collection(COLLECTION).FindOne(ctx, bson.D{{Key: "username", Value: username}}).Decode(&creds)
	defer cancel()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Println(fmt.Sprintf("no matching username: %s", username))
			return
		}
		log.Fatal(err)
	}
	return creds, err
}

func Signup(w http.ResponseWriter, r *http.Request) {
	creds := &Credentials{}
	err := json.NewDecoder(r.Body).Decode(creds)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), 8)
	if err != nil {
		log.Fatal(err)
	}
	creds.Password = string(hashedPassword)

	if err = d.Insert(*creds); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func Signin(w http.ResponseWriter, r *http.Request) {
	creds := &Credentials{}
	err := json.NewDecoder(r.Body).Decode(creds)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	result, err := d.FindByUsername(creds.Username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err = bcrypt.CompareHashAndPassword([]byte(result.Password), []byte(creds.Password)); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Println(fmt.Sprintf("%s unauthorized", creds.Username))
	} else {
		fmt.Println(fmt.Sprintf("Success! %s is authorized.", creds.Username))
	}
}
