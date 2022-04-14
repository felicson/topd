package main

import (
	"github.com/felicson/topd/internal/keeper"
	"github.com/felicson/topd/internal/log"
	"github.com/felicson/topd/storage"
)

type keeperD struct {
	siteCollection *storage.SiteAggregate
	sessionPerSite *storage.SessionsPerSite
	logger         log.Logger
	topData        *storage.TopDataCollection
	storage        storage.Storage
}

func (k *keeperD) GetSessionCleaner() keeper.SessionCleaner {
	return k.sessionPerSite
}

func (k *keeperD) GetLogger() log.Logger {
	return k.logger
}

func (k *keeperD) GetSites() keeper.SiteCollector {
	return k.siteCollection
}

func (k *keeperD) GetStorage() keeper.Saver {
	return k.storage
}

func (k *keeperD) GetTopData() *storage.TopDataCollection {
	return k.topData
}
