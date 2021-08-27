package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/x/mongo/driver/uuid"
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
}

func Signin(w http.ResponseWriter, r *http.Request) {
	creds := &Credentials{}
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	storedCreds := &Credentials{}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
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
	}

	if err = bcrypt.CompareHashAndPassword([]byte(storedCreds.Password), []byte(creds.Password)); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Println(fmt.Sprintf("%s unauthorized", creds.Username))
	} else {
		sessionToken, err := uuid.New()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, err = cache.Do("SETEX", sessionToken, "120", creds.Username)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:    "session_token",
			Value:   fmt.Sprint(sessionToken),
			Expires: time.Now().Add(120 * time.Second),
		})
		fmt.Println(fmt.Sprintf("Success! %s is authorized.", creds.Username))
	}

}
