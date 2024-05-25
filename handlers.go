package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func Signup(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Signing up...")

	// Decode request body and handle bad requests
	creds := &Credentials{}
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return // Early return for bad request
	}

	// Hash password and handle errors
	hashedPassword, err := hashPassword(creds.Password) // Refactored to separate function
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return // Early return for internal server error
	}

	creds.Password = string(hashedPassword)

	// Database interaction with context and prepared statement
	if err := createUser(r.Context(), creds); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return // Early return for internal server error
	}

	w.WriteHeader(http.StatusOK)
	fmt.Printf("Success! %s is registered.\n", creds.Username)
}

// Separate function for hashing password with error handling
func hashPassword(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), 8)
}

// Separate function for user creation with context and prepared statement
func createUser(ctx context.Context, creds *Credentials) error {
	insertUserSQL := `INSERT INTO user(username, password) VALUES (?, ?)`
	statement, err := dao.DB.PrepareContext(ctx, insertUserSQL)
	if err != nil {
		return err
	}
	defer statement.Close() // Ensure statement is closed

	_, err = statement.Exec(creds.Username, creds.Password)
	return err
}

func Signout(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Signing out...")

	// Get session token from cookie
	sessionCookie, err := r.Cookie("session_token")
	if err != nil {
		// Handle missing cookie gracefully (e.g., inform user they're not signed in)
		fmt.Println("No session cookie found.")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Delete session token from cache
	_, err = cache.Do("DEL", sessionCookie.Value)
	if err != nil {
		fmt.Printf("Error deleting session token: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Remove session cookie from client
	http.SetCookie(w, &http.Cookie{
		Name:   "session_token",
		Value:  "",
		MaxAge: -1, // Expire immediately
	})

	w.WriteHeader(http.StatusOK)
	fmt.Println("Successfully signed out.")
}

func Signin(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Signing in...")

	creds := &Credentials{}
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check session token first (early return)
	sessionToken, err := validateSessionToken(r)
	if err == nil {
		fmt.Printf("Success! %s is authorized via session token.\n", sessionToken)
		w.WriteHeader(http.StatusOK)
		return
	} else if err != ErrInvalidSessionToken {
		// Handle other errors (e.g., redis errors)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fmt.Println("Querying database...")
	// Username/Password Login
	var sqlQuery = `SELECT username, password FROM user WHERE username = ?`

	rows, err := dao.DB.Query(sqlQuery, creds.Username)
	if err != nil {
		// Handle error (consider different error types)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer rows.Close() // Ensure rows are closed

	if !rows.Next() {
		// No matching username
		w.WriteHeader(http.StatusNotFound)
		fmt.Printf("no matching username: %s", creds.Username)
		return
	}

	storedCreds := &Credentials{}
	err = rows.Scan(&storedCreds.Username, &storedCreds.Password) // Scan row data
	if err != nil {
		// Handle error (consider different error types)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fmt.Println("Hashing password...")

	if err = bcrypt.CompareHashAndPassword([]byte(storedCreds.Password), []byte(creds.Password)); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Printf("%s unauthorized.\n%s\n", creds.Username, err)
		return
	}

	fmt.Println("Creating session token...")
	// Login successful, create new session
	sessionToken = uuid.NewString()
	_, err = cache.Do("SETEX", sessionToken, 120, creds.Username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   fmt.Sprint(sessionToken),
		Expires: time.Now().Add(120 * time.Second),
	})
	w.WriteHeader(http.StatusOK)
	fmt.Printf("Session token: %s\n", sessionToken)
	fmt.Printf("Success! %s is authorized.\n", creds.Username)
}

func validateSessionToken(r *http.Request) (string, error) {
	fmt.Println("Validating Session Token...")
	storedSessionCookie, err := r.Cookie("session_token")
	if err != nil {
		return "", ErrInvalidSessionToken
	}
	storedSessionToken := storedSessionCookie.Value
	cachedSessionUsername, err := redis.String(cache.Do("GET", storedSessionToken))
	if err != nil {
		return "", err // Consider wrapping in a specific error type here
	}
	if cachedSessionUsername != "" { // Early return if valid
		return cachedSessionUsername, nil
	}
	return "", ErrInvalidSessionToken
}

var ErrInvalidSessionToken = errors.New("invalid session token") // Example custom error type
