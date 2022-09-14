package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func Signup(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Signing up...")
	creds := &Credentials{}
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		log.Fatal(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), 8)
	if err != nil {
		log.Fatal(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	creds.Password = string(hashedPassword)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	insertUserSQL := `INSERT INTO user(username, password) VALUES (?, ?)`
	statement, err := dao.DB.PrepareContext(ctx, insertUserSQL) // Prepare statement.
	// This is good to avoid SQL injections
	if err != nil {
		log.Fatal(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, err = statement.Exec(creds.Username, creds.Password)
	if err != nil {
		log.Fatal(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Println(fmt.Sprintf("Success! %s is registered.", creds.Username))
	return
}

func Signin(w http.ResponseWriter, r *http.Request) {
	creds := &Credentials{}
	storedCreds := &Credentials{}
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else {
		storedSessionCookie, err := r.Cookie("session_token")
		if err != nil {
			fmt.Println(fmt.Sprintf("No active session found for %s", creds.Username))
		} else {
			storedSessionToken := storedSessionCookie.Value
			cachedSessionUsername, err := redis.String(cache.Do("GET", storedSessionToken))
			if err != nil {
				fmt.Println(fmt.Sprintf("No active session found for %s", creds.Username))
			} else if cachedSessionUsername == creds.Username {
				fmt.Println(fmt.Sprintf("Success! %s is authorized via session token.", cachedSessionUsername))
				w.WriteHeader(http.StatusOK)
				return
			}
		}
		// refactor mongo code to sql code
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		getUserSQL := `SELECT password FROM user WHERE username = ?`
		statement, err := dao.DB.PrepareContext(ctx, getUserSQL) // Prepare statement.
		// This is good to avoid SQL injections
		if err != nil {
			log.Fatal(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = statement.QueryRowContext(ctx, creds.Username).Scan(&storedCreds.Password)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			fmt.Println(fmt.Sprintf("no matching username: %s", creds.Username))
			return
		}
		if err = bcrypt.CompareHashAndPassword([]byte(storedCreds.Password), []byte(creds.Password)); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Println(fmt.Sprintf("%s unauthorized", creds.Username))
			return
		} else {
			sessionToken := uuid.NewString()
			_, err = cache.Do("SETEX", sessionToken, 1200, creds.Username)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			http.SetCookie(w, &http.Cookie{
				Name:    "session_token",
				Value:   fmt.Sprint(sessionToken),
				Expires: time.Now().Add(1200 * time.Second),
			})
			w.WriteHeader(http.StatusOK)
			fmt.Println(fmt.Sprintf("Session token: %s", sessionToken))
			fmt.Println(fmt.Sprintf("Success! %s is authorized.", creds.Username))
			return
		}
	}

}
