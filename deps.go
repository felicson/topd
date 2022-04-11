package topd

import (
	"github.com/felicson/topd/internal/bot"
	"github.com/felicson/topd/internal/config"
	"github.com/felicson/topd/internal/log"
	"github.com/felicson/topd/storage"
)

type Deps interface {
	GetConfig() config.Config
	GetLogger() log.Logger
	GetSiteCollection() *storage.SiteAggregate
	GetSessionPerSite() *storage.SessionsPerSite
	GetHistoryWriter() *storage.HistoryCollector
	GetBotChecker() *bot.Checker
}
