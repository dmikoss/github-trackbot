package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/dmikoss/GithubTrackBot/bot"
	"github.com/dmikoss/GithubTrackBot/config"
)

func main() {
	config := config.New()
	if config.TelegramToken == "" {
		log.Fatalf("Error! You must provide env variable TELEGRAM_BOT_TOKEN")
	}
	log.Print("Using token" + config.TelegramToken)

	trends := bot.New(http.DefaultClient)
	fmt.Println(trends.FetchLanguagesList())
}
