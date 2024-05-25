package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

var (
	dao        = DAO{}
	port       int
	dbFilename string
	err        error
)

// Parse the configuration file 'conf.json', and establish a connection to DB
func init() {
	port = *flag.Int("port", 3000, "specified port")
	dbFilename = *flag.String("dbFilename", "users.db", "filename for SQLite database")
	flag.Parse()
	dao.DatabaseSourceName = dbFilename

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
	mux.HandleFunc("/signout", Signout)

	fmt.Printf("Listening on port %s\n", strconv.Itoa(port))
	if err = http.ListenAndServe(":"+strconv.Itoa(port), mux); err != nil {
		log.Fatal(err)
	}
	defer dao.DB.Close() // Defer Closing the database
}
