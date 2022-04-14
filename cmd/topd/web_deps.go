package main

import (
	"github.com/felicson/topd/internal/bot"
	"github.com/felicson/topd/internal/config"
	"github.com/felicson/topd/internal/log"
	"github.com/felicson/topd/storage"
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
