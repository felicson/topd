package keeper

import (
	"github.com/felicson/topd/internal/log"
	"github.com/felicson/topd/storage"
)

type Deps interface {
	GetSessionCleaner() SessionCleaner
	GetLogger() log.Logger
	GetSites() SiteCollector
	GetStorage() Saver
	GetTopData() *storage.TopDataCollection
}
