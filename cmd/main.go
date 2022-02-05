package main

import (
	"log"

	"github.com/dmikoss/GithubTrackBot/bot"
)

func main() {
	bot := bot.NewBot()
	err := bot.Run()
	if err != nil {
		log.Fatalf(err.Error())
	}
}
