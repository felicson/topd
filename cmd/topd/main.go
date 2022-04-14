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
	"github.com/felicson/topd/internal/keeper"
	"github.com/felicson/topd/storage"
	"github.com/felicson/topd/storage/memory"
	"go.uber.org/zap"
)

func main() {

	var (
		envFlag  string
		confFlag string
	)
	flag.StringVar(&envFlag, "env", "a", "env from conf.yml")
	flag.StringVar(&confFlag, "conf", "config.yml", "conf path")
	flag.Parse()

	conf, err := config.NewConfig(confFlag, envFlag)
	if err != nil {
		stdlog.Fatalf("on parse conf: %v", err)
	}

	if err := run(conf); err != nil {
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

	ctx := context.Background()

	siteMap := storage.NewSiteAggregate(store, images)
	siteMap.Init()

	sps := storage.NewSessionPerSite()

	done := make(chan struct{}, 1)
	defer close(done)

	var topDataCollection storage.TopDataCollection

	kd := keeperD{
		siteCollection: &siteMap,
		sessionPerSite: sps,
		logger:         logger,
		topData:        &topDataCollection,
		storage:        store,
	}

	kpr, _ := keeper.New(&kd)
	go kpr.Run(ctx, done)

	hCollector := storage.NewHistoryCollector(&topDataCollection, 10)
	hCollector.Run(ctx)

	deps := webApp{
		config:         config,
		siteCollection: &siteMap,
		sessionPerSite: sps,
		logger:         logger,
		historyWriter:  hCollector,
		botChecker:     &bChecker,
	}

	logger.Info("running topd")
	if err := topd.Run(ctx, &deps, done); err != nil {
		return fmt.Errorf("on run topd: %v", err)
	}
	return nil
}
