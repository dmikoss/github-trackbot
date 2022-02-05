package bot

import (
	"fmt"
	"log"

	"github.com/dmikoss/GithubTrackBot/config"
)

type Bot struct {
	offset int
}

// New returns a new Bot object
func NewBot() *Bot {
	return &Bot{}
}

func (b *Bot) Run() error {
	config := config.New()
	if config.TelegramToken == "" {
		log.Fatalf("Error! You must provide valid ENV variable TELEGRAM_BOT_TOKEN")
	}
	log.Println("Using token " + config.TelegramToken)

	tgClient := NewTelegramClient(config.TelegramHost, config.TelegramToken)

	for {
		// delay 1 sec, max 100 updates in one batch
		updates, err := tgClient.Updates(b.offset, 1)
		if err != nil {
			log.Println("Get bot telegram updates failed:" + err.Error())
		}

		for _, update := range updates {
			if b.offset < update.ID+1 {
				b.offset = update.ID + 1
			}
		}
		fmt.Println(updates)
	}
	return nil
}
