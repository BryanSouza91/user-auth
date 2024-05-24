package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
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

	err = dao.Connection()
	if err != nil {
		log.Printf("Connection Error: %q\n", err)
	}
	initCache()
}

// Define HTTP request routes
func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/signup", Signup)
	mux.HandleFunc("/signin", Signin)
	fmt.Printf("Listening on port %s\n", strconv.Itoa(*port))
	if err = http.ListenAndServe(":"+strconv.Itoa(*port), mux); err != nil {
		log.Fatal(err)
	}
	defer dao.DB.Close() // Defer Closing the database
}
