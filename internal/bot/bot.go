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

	"golang.org/x/time/rate"

	"github.com/dmikoss/GithubTrackBot/internal/config"
	"github.com/robfig/cron/v3"
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
	f := func() {
		b.cronFetchTrendingGithub(ctx, config)
	}
	c := cron.New()
	c.AddJob("@every 60m", cron.NewChain(cron.SkipIfStillRunning(cron.DefaultLogger)).Then(cron.FuncJob(f)))
	c.Start()

	tgClient := NewTelegramClient(config.TelegramApiHost, config.TelegramBotToken)
	chwait := make(chan struct{}, 1)
	go tgClient.RunRecvMessages(ctx, chwait, 100, 1)

	<-chwait          // wait tgClient.RunRecvMessages
	<-ctx.Done()      // wait context cancel
	<-c.Stop().Done() // wait gocron jobs
	return nil
}

func (b *Bot) cronFetchTrendingGithub(ctx context.Context, config *config.Config) error {
	// first download language list
	fetcherClient := NewFetcher(http.DefaultClient)
	languages, err := fetcherClient.FetchLanguagesList()
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
				fmt.Println(err)
			}

			// TODO
			//_, err := fetcherClient.FetchRepos(period, lang)
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
	fmt.Println("Graceful")
	return nil
}
