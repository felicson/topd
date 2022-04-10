package main

import (
	"context"
	"flag"
	"io"
	"log"
	"os"

	"github.com/felicson/topd"
	"github.com/felicson/topd/image"
	"github.com/felicson/topd/internal/bot"
	"github.com/felicson/topd/internal/config"
	"github.com/felicson/topd/storage"
	"github.com/felicson/topd/storage/mysql"
)

func main() {

	var (
		env  string
		conf string
	)
	flag.StringVar(&env, "env", "a", "env from config.yml")
	flag.StringVar(&conf, "config", "./config.yml", "config path")
	flag.Parse()

	config, err := config.NewConfig(conf, env)
	if err != nil {
		log.Fatal(err)
	}

	bChecker, err := bot.NewCheckerFromFile(config.BotsList)
	if err != nil {
		log.Fatal(err)
	}

	images, err := image.NewImages(config.ImagesPath)
	if err != nil {
		log.Fatal(err)
	}

	store, err := mysql.New(config)
	//store, err := memory.New(config)
	if err != nil {
		log.Fatalf("on make storage: %v", err)
	}
	defer store.Close()

	siteMap := storage.NewSiteAggregate(store, images)
	siteMap.Init()

	sps := storage.NewSessionPerSite()

	done := make(chan struct{}, 1)
	defer close(done)

	go siteMap.SigHandler(done)

	logger, err := os.OpenFile(config.Logfile, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		log.Fatalf("Open log file: %v", err)
	}
	defer logger.Close()

	ctx := context.Background()
	hCollector := storage.NewHistoryCollector(10)
	hCollector.Run(ctx)

	deps := webApp{
		config:         config,
		siteCollection: &siteMap,
		sessionPerSite: sps,
		logger:         logger,
		historyWriter:  hCollector,
		botChecker:     &bChecker,
	}

	if err = topd.Run(ctx, &deps, done); err != nil {
		log.Fatal(err)
	}

}

type webApp struct {
	config         config.Config
	siteCollection *storage.SiteAggregate
	sessionPerSite *storage.SessionsPerSite
	logger         io.Writer
	historyWriter  storage.HistoryCollector
	botChecker     *bot.Checker
}

func (wa *webApp) GetLogger() io.Writer {
	return wa.logger
}

func (wa *webApp) GetConfig() config.Config {
	return wa.config
}

func (wa *webApp) GetSiteCollection() *storage.SiteAggregate {
	return wa.siteCollection
}

func (wa *webApp) GetSessionPerSite() *storage.SessionsPerSite {
	return wa.sessionPerSite
}

func (wa *webApp) GetHistoryWriter() *storage.HistoryCollector {
	return &wa.historyWriter
}

func (wa *webApp) GetBotChecker() *bot.Checker {
	return wa.botChecker
}
