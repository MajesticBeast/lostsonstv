package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	lstv "github.com/majesticbeast/lostsonstv/types"
)

type profile struct {
	Port    string
	BaseUrl string
}

var localhost = profile{
	Port:    ":8083",
	BaseUrl: "http://localhost",
}

var remote = profile{
	Port:    ":80",
	BaseUrl: "http://thelostsons.net",
}

func main() {
	godotenv.Load()

	activeProfile := localhost

	dbConnStr := os.Getenv("DB_CONN_STR")

	muxApiAuth := lstv.MuxApiAuth{
		Id:    os.Getenv("MUX_TOKEN_ID"),
		Token: os.Getenv("MUX_TOKEN_SECRET"),
	}

	discordClient := NewDiscordClient(os.Getenv("DISCORD_CLIENT_ID"), os.Getenv("DISCORD_CLIENT_SECRET"), activeProfile.BaseUrl+"/redirect")

	store, err := NewPostgresStore(dbConnStr)
	if err != nil {
		log.Fatal(err)
	}

	if err := store.Init(); err != nil {
		log.Fatal(err)
	}

	server := NewAPIServer(activeProfile.Port, store, muxApiAuth, discordClient)
	server.Run()
}
