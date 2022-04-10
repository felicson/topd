package storage

import (
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

type sessionH map[int]map[string]struct{}

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

type SessionsPerSite struct {
	sessions sessionH
	lock     sync.RWMutex
}

func NewSessionPerSite() *SessionsPerSite {
	sps = &SessionsPerSite{sessions: make(sessionH)}
	return sps
}

// CheckSession checking session in hash
func (sps *SessionsPerSite) CheckSession(siteID int, session string) bool {

	if site, ok := sps.sessions[siteID]; ok {
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

	if _, ok := sps.sessions[siteID]; !ok {
		sps.sessions[siteID] = map[string]struct{}{session: {}}
		return
	}
	sps.sessions[siteID][session] = struct{}{}
}
