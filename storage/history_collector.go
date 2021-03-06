package storage

import (
	"context"
	"net"
)

// HistoryCollector provide instance for store user activity history
type HistoryCollector struct {
	active   bool
	dataChan chan RawTopData
	topData  *TopDataCollection
}

// Run starts history collector
func (hc *HistoryCollector) Run(ctx context.Context) {
	hc.active = true
	go func() {

	LOOP:
		for {
			select {
			case item := <-hc.dataChan:
				*hc.topData = append(*hc.topData, newHistoryRow(&item))
			case <-ctx.Done():
				break LOOP
			}
		}

	}()
}

func (hc *HistoryCollector) WriteHistory(page, referrer, xGeo, session, userAgent string, ip net.IP, siteID int) error {
	if !hc.active {
		return ErrHistoryCollectorStopped
	}

	hc.dataChan <- RawTopData{page, referrer, xGeo, session, userAgent, ip, siteID}
	return nil
}

func NewHistoryCollector(dst *TopDataCollection, cap int) HistoryCollector {
	return HistoryCollector{dataChan: make(chan RawTopData, cap), topData: dst}
}
