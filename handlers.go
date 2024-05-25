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
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

func Signup(w http.ResponseWriter, r *http.Request) {
	creds := &Credentials{}
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), 8)
	if err != nil {
		log.Fatal(err)
	}
	creds.Password = string(hashedPassword)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	_, err = db.Collection(COLLECTION).InsertOne(ctx, &creds)
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
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		defer cancel()
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
		err = db.Collection(COLLECTION).FindOne(ctx, bson.D{{Key: "username", Value: creds.Username}}).Decode(&storedCreds)
		defer cancel()
		if err != nil {
			if err == mongo.ErrNoDocuments {
				w.WriteHeader(http.StatusNotFound)
				fmt.Println(fmt.Sprintf("no matching username: %s", creds.Username))
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			log.Fatal(err)
			return
		}
		if err = bcrypt.CompareHashAndPassword([]byte(storedCreds.Password), []byte(creds.Password)); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Println(fmt.Sprintf("%s unauthorized", creds.Username))
			return
		} else {
			sessionToken := uuid.NewString()
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
			fmt.Println(fmt.Sprintf("Session token: %s", sessionToken))
			fmt.Println(fmt.Sprintf("Success! %s is authorized.", creds.Username))
			return
		}
	}

}
