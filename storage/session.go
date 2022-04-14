package storage

import "sync"

type sessionH map[int]map[string]struct{}

type SessionsPerSite struct {
	sessions sessionH
	lock     sync.RWMutex
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

func (sps *SessionsPerSite) Reset() error {
	sps.lock.Lock()
	sps.sessions = make(sessionH)
	sps.lock.Unlock()
	return nil
}

func (sps *SessionsPerSite) append(siteID int, session string) {

	if _, ok := sps.sessions[siteID]; !ok {
		sps.sessions[siteID] = map[string]struct{}{session: {}}
		return
	}
	sps.sessions[siteID][session] = struct{}{}
}

func NewSessionPerSite() *SessionsPerSite {
	return &SessionsPerSite{sessions: make(sessionH)}
}
