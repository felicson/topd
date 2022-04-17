package keeper

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/felicson/topd/internal/log"
	"github.com/felicson/topd/storage"
)

type SiteCollector interface {
	Reset() bool
	Init()
	KeepState() error
}

type SessionCleaner interface {
	Reset() error
}

type Saver interface {
	SaveData([]storage.TopData) error
}

type Keeper struct {
	siteCollector SiteCollector
	sessCleaner   SessionCleaner
	topData       *storage.TopDataCollection
	storage       Saver
	logger        log.Logger
}

func (k *Keeper) Run(ctx context.Context, done chan<- struct{}) {

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	ticker := time.NewTicker(10 * time.Second)
	defer close(sigChan)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case sig := <-sigChan:
			t := (*k.topData)[:len(*k.topData):len(*k.topData)]
			if err := k.storage.SaveData(t); err != nil {
				k.logger.Error(err)
			}
			if err := k.siteCollector.KeepState(); err != nil {
				k.logger.Error(err)
			}

			if sig == syscall.SIGHUP {
				k.siteCollector.Reset()
				*k.topData = (*k.topData)[:0]
				if err := k.sessCleaner.Reset(); err != nil {
					k.logger.Error(err)
				}
				continue
			}
			done <- struct{}{}
			return

		case <-ticker.C:
			t := (*k.topData)[:len(*k.topData):len(*k.topData)]
			*k.topData = (*k.topData)[:0]
			if err := k.storage.SaveData(t); err != nil {
				k.logger.Error(err)
			}
			if err := k.siteCollector.KeepState(); err != nil {
				k.logger.Error(err)
			}
			k.siteCollector.Init()
		}
	}
}

func New(deps Deps) (Keeper, error) {

	return Keeper{
		siteCollector: deps.GetSites(),
		sessCleaner:   deps.GetSessionCleaner(),
		topData:       deps.GetTopData(),
		storage:       deps.GetStorage(),
		logger:        deps.GetLogger(),
	}, nil
}
