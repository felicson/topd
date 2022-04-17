package storage

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/felicson/topd/image"
)

var (
	ErrHistoryCollectorStopped = errors.New("history collector stopped")
)

type Storage interface {
	Populate(int) ([]Site, error)
	UpdateSites([]Site) error
	SaveData([]TopData) error
}

type TopDataCollection []TopData

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

// KeepState saves in the storage current hits and hosts values of the sites
func (sm *SiteAggregate) KeepState() error {

	sm.lock.Lock()
	var sites []Site
	for k := range sm.sites {
		site := sm.sites[k]
		if site.Hosts == 0 {
			continue
		}
		sites = append(sites, *site)
	}
	sm.lock.Unlock()
	if err := sm.storage.UpdateSites(sites); err != nil {
		return fmt.Errorf("on update sites: %v", err)
	}
	return nil
}

// Init populate SiteAggregate from storage.
// On first call it receiving all records from storage, on another calls only new ones.
func (sm *SiteAggregate) Init() {

	sm.lock.Lock()
	defer sm.lock.Unlock()
	sites, _ := sm.storage.Populate(sm.lastID)

	for i := range sites {
		site := &sites[i]
		if site.ID > sm.lastID {
			sm.lastID = site.ID
		}
		s := NewSite(site.ID, site.CounterID, site.Hosts, site.Hits, site.Digits)
		sm.sites[site.ID] = &s
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
