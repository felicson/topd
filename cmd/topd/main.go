package main

import (
	"context"
	"flag"
	"fmt"
	stdlog "log"

	"github.com/felicson/topd"
	"github.com/felicson/topd/image"
	"github.com/felicson/topd/internal/bot"
	"github.com/felicson/topd/internal/config"
	"github.com/felicson/topd/internal/log"
	"github.com/felicson/topd/storage"
	"github.com/felicson/topd/storage/memory"
	"go.uber.org/zap"
)

type webApp struct {
	config         config.Config
	siteCollection *storage.SiteAggregate
	sessionPerSite *storage.SessionsPerSite
	logger         log.Logger
	historyWriter  storage.HistoryCollector
	botChecker     *bot.Checker
}

func (wa *webApp) GetLogger() log.Logger {
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
		stdlog.Fatalf("on parse config: %v", err)
	}

	if err := run(config); err != nil {
		stdlog.Fatal(err)
	}

}

func run(config config.Config) error {

	level, err := zap.ParseAtomicLevel(config.LogLevel)
	if err != nil {
		return fmt.Errorf("on parse log level: %v", err)
	}

	zapConf := zap.NewProductionConfig()
	zapConf.Level.SetLevel(level.Level())
	zapConf.OutputPaths = []string{config.Logfile}

	zapLog, err := zapConf.Build()
	if err != nil {
		return fmt.Errorf("on build logger: %v", err)
	}
	defer zapLog.Sync()

	logger := zapLog.Sugar()

	bChecker, err := bot.NewCheckerFromFile(config.BotsList)
	if err != nil {
		return fmt.Errorf("on build bot checker: %v", err)
	}

	images, err := image.NewImages(config.ImagesPath)
	if err != nil {
		return fmt.Errorf("on build images: %v", err)
	}

	//store, err := mysql.New(config)
	store, err := memory.New(config)
	if err != nil {
		return fmt.Errorf("on make storage: %v", err)
	}
	defer store.Close()

	siteMap := storage.NewSiteAggregate(store, images)
	siteMap.Init()

	sps := storage.NewSessionPerSite()

	done := make(chan struct{}, 1)
	defer close(done)

	go siteMap.SigHandler(done)

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

	logger.Info("runing top")
	if err = topd.Run(ctx, &deps, done); err != nil {
		return fmt.Errorf("on run topd: %v", err)
	}
	return nil
}
