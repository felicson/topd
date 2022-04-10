package storage

import (
	"context"
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/felicson/topd/image"
)

var (
	ErrHistoryCollectorStopped = errors.New("history collector stopped")
)

type Storage interface {
	Populate(int) ([]Site, error)
	SaveData([]TopData) error
}

var TopDataArray []TopData

type TopData struct {
	Page     string
	Referrer string
	Sess     string
	City     int
	Country  string
	SiteID   int
	IP       net.IP
	UA       string
	Date     time.Time
}

type Site struct {
	Hosts     int
	Hits      int
	ID        int
	CounterID int
	Digits    bool
	l         sync.RWMutex
}

//DisplayDigits check need to show digits on counter
func (s *Site) DisplayDigits() bool {
	return s.Digits
}

func (s *Site) Increment(hosts bool, hits bool) {

	s.l.Lock()
	defer s.l.Unlock()

	if hosts {
		s.Hosts += 1
	}
	if hits {
		s.Hits += 1
	}
}

func NewSite(id, counterID, visitors, hits int, digits bool) Site {
	return Site{
		ID:        id,
		CounterID: counterID,
		Hits:      hits,
		Hosts:     visitors,
		Digits:    digits,
	}
}

type RawTopData struct {
	page      string
	referrer  string
	xGeo      string
	session   string
	userAgent string
	ip        net.IP
	id        int
}

type HistoryCollector struct {
	active   bool
	dataChan chan *RawTopData
}

// Run starts history collector
func (hc *HistoryCollector) Run(ctx context.Context) {
	hc.active = true
	go func() {

	LOOP:
		for {
			select {
			case item := <-hc.dataChan:
				TopDataArray = append(TopDataArray, newHistoryRow(item))
			case <-ctx.Done():
				break LOOP
			}
		}

	}()
}

func NewHistoryCollector(cap int) HistoryCollector {
	return HistoryCollector{dataChan: make(chan *RawTopData, cap)}
}

func (hc *HistoryCollector) WriteHistory(page, referrer, xGeo, session, userAgent string, ip net.IP, siteID int) error {
	if !hc.active {
		return ErrHistoryCollectorStopped
	}

	hc.dataChan <- &RawTopData{page, referrer, xGeo, session, userAgent, ip, siteID}
	return nil
}

func newHistoryRow(raw *RawTopData) TopData {

	cityID := 0
	country := "0"

	if city := raw.xGeo; city != "" && city != "0" {
		cityID, _ = strconv.Atoi(city[3:])
		country = city[:2]
	}

	return TopData{
		Page:     raw.page,
		Referrer: raw.referrer,
		Sess:     raw.session,
		City:     cityID,
		Country:  country,
		SiteID:   raw.id,
		IP:       raw.ip,
		UA:       raw.userAgent,
		Date:     time.Now(),
	}
}

const SIGHUP = syscall.SIGHUP

var sps *SessionsPerSite

type SessionsPerSite struct {
	data map[int]map[string]struct{}
	lock sync.RWMutex
}

func NewSessionPerSite() *SessionsPerSite {
	sps = &SessionsPerSite{data: make(map[int]map[string]struct{})}
	return sps
}

type SiteAggregate struct {
	lock    sync.RWMutex
	sites   map[int]*Site
	lastID  int
	images  image.ImageList
	storage Storage
}

func (sm *SiteAggregate) GetImage(id int) (image.Image, error) {
	return sm.images.GetImage(uint(id))
}

//Get SiteID by id
func (sm *SiteAggregate) Get(key int) (*Site, bool) {
	sm.lock.RLock()
	defer sm.lock.RUnlock()
	d, ok := sm.sites[key]
	return d, ok
}

func (sm *SiteAggregate) Set(key int, d *Site) {
	sm.lock.Lock()
	defer sm.lock.Unlock()
	sm.sites[key] = d
}

//Reset sites statistic
func (sm *SiteAggregate) Reset() bool {

	sm.lock.Lock()
	defer sm.lock.Unlock()

	for _, v := range sm.sites {

		v.Hits = 0
		v.Hosts = 0
	}
	return true
}

func (sm *SiteAggregate) Init() {

	sm.lock.Lock()
	defer sm.lock.Unlock()
	sites, _ := sm.storage.Populate(sm.lastID)

	for _, site := range sites {
		if site.ID > sm.lastID {
			sm.lastID = site.ID
		}
		s := NewSite(site.ID, site.CounterID, site.Hosts, site.Hits, site.Digits)
		sm.sites[site.ID] = &s
	}
}

//SigHandler handler of signals
func (sm *SiteAggregate) SigHandler(done chan<- struct{}) {

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	ticker := time.NewTicker(10 * time.Second)
	defer close(sigChan)
	defer ticker.Stop()

	for {
		select {
		case sig := <-sigChan:
			if sig == SIGHUP {
				sm.storage.SaveData(TopDataArray[:len(TopDataArray):len(TopDataArray)])
				sm.Reset()
				TopDataArray = make([]TopData, 0)
				sps = NewSessionPerSite()
				continue
			}
			done <- struct{}{}
			return

		case <-ticker.C:
			t := TopDataArray[:len(TopDataArray):len(TopDataArray)]
			TopDataArray = make([]TopData, 0)
			if err := sm.storage.SaveData(t); err != nil {
				log.Println(err)
			}
			sm.Init()
		}
	}
}

//NewSiteAggregate gen new struct from db
func NewSiteAggregate(storage Storage, images image.ImageList) SiteAggregate {
	return SiteAggregate{
		sites:   make(map[int]*Site),
		images:  images,
		storage: storage,
	}
}

// CheckSession checking session in hash
func (sps *SessionsPerSite) CheckSession(siteID int, session string) bool {

	if site, ok := sps.data[siteID]; ok {
		if _, ok = site[session]; ok {
			return ok
		}
	}

	sps.lock.Lock()
	sps.append(siteID, session)
	sps.lock.Unlock()

	return false
}

func (sps *SessionsPerSite) append(siteID int, session string) {

	if _, ok := sps.data[siteID]; !ok {
		sps.data[siteID] = map[string]struct{}{session: {}}
		return
	}
	sps.data[siteID][session] = struct{}{}
}
