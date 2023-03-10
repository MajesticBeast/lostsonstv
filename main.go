package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type muxApiAuth struct {
	Id    string
	Token string
}

func main() {
	godotenv.Load()

	dbConnStr := os.Getenv("DB_CONN_STR")

	muxApiAuth := muxApiAuth{
		Id:    os.Getenv("MUX_TOKEN_ID"),
		Token: os.Getenv("MUX_TOKEN_SECRET"),
	}

	store, err := NewPostgresStore(dbConnStr)
	if err != nil {
		log.Fatal(err)
	}

	if err := store.Init(); err != nil {
		log.Fatal(err)
	}

	server := NewAPIServer(":8083", store, muxApiAuth)
	server.Run()
}
