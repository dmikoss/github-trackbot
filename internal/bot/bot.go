package bot

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"golang.org/x/time/rate"

	"github.com/dmikoss/github-trackbot/internal/config"
	"github.com/go-co-op/gocron"
)

type Bot struct {
}

// New returns a new Bot object
func NewBot() *Bot {
	return &Bot{}
}

func (b *Bot) Run() error {
	config := config.New()
	if config.TelegramBotToken == "" {
		log.Fatalf("Error! You must provide valid ENV variable TELEGRAM_BOT_TOKEN")
	}
	log.Println("Using token: " + config.TelegramBotToken)

	ctx, cancel := context.WithCancel(context.Background())
	// cancel context on system signal
	go func() {
		exit := make(chan os.Signal, 1)
		signal.Notify(exit, syscall.SIGINT, syscall.SIGTERM)
		oscall := <-exit
		cancel()
		log.Printf("Reciever call:%+v. Stopping...Press Ctrl+C to force stop.", oscall)
		<-exit
		os.Exit(1)
	}()

	// schelude periodic fetch from github trending pages
	s := gocron.NewScheduler(time.UTC)
	s.Every(config.GithubFetchEvery).SingletonMode().Do(func() {
		b.fetchGithubJob(ctx, config)
	})
	s.StartAsync()

	httpclient := &http.Client{
		Timeout: time.Second * time.Duration(config.TelegramHttpTimeout),
	}
	tgClient := NewTelegramClient(config.TelegramApiHost, config.TelegramBotToken, httpclient)
	chwait := make(chan struct{}, 1)
	go tgClient.RunRecvMsgLoop(ctx, chwait, 100, 1)

	<-chwait     // wait tgClient.RunRecvMessages
	<-ctx.Done() // wait context cancel
	s.Stop()     // wait cron jobs
	return nil
}

func (b *Bot) fetchGithubJob(ctx context.Context, config *config.Config) error {
	httpclient := &http.Client{
		Timeout: time.Second * time.Duration(config.GithubFetchTimeout),
	}
	fetcher := NewFetcher(ctx, httpclient)
	// first download and parse language list
	languages, err := fetcher.FetchLanguagesList()
	if err != nil {
		return err
	}
	// github limits requests rate, so we need limit it
	RPS := config.GithubFetchRate // per second
	limiter := rate.NewLimiter(rate.Limit(RPS), 1)

fetchLoop:
	// download all lang/period permutations
	for period := TimeDaily; period < TimeMonth; period++ {
		for _, lang := range languages {
			limiter.Wait(ctx)
			if err != nil {
				return err
			}

			_, err := fetcher.FetchRepos(period, lang)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println("Fetch " + lang.Name + " " + strconv.Itoa(int(period)))

			select { // early exit on graceful shutdown
			case <-ctx.Done():
				break fetchLoop
			default:
			}
		}
	}
	return nil
}
