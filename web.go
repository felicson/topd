package topd

import (
	"bytes"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/felicson/topd/internal/bot"
	"github.com/felicson/topd/internal/config"
	"github.com/felicson/topd/internal/log"
	"github.com/felicson/topd/internal/session"
	"github.com/felicson/topd/storage"
)

type historyWriter interface {
	WriteHistory(page, referrer, xGeo, session, userAgent string, ip net.IP, siteID int) error
}

type Web struct {
	siteMap        *storage.SiteAggregate
	sessionPerSite *storage.SessionsPerSite
	config         config.Config
	historyWriter  historyWriter
	bots           *bot.Checker
	logger         log.Logger
}

func (web *Web) logHandler(next http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, req *http.Request) {

		ip := req.Header.Get("X-Real-IP")
		start := time.Now()
		next(w, req)
		web.logger.Infof("%s [%s] %s %v", ip, req.Method, req.URL.String(), time.Now().Sub(start))
	}
}

func (web *Web) ErrHandler(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {

	return func(w http.ResponseWriter, req *http.Request) {

		reqSiteID := req.FormValue("id")
		siteID, _ := strconv.Atoi(reqSiteID)

		if _, ok := web.siteMap.Get(siteID); !ok {
			// return first image with zero values
			image, _ := web.siteMap.GetImage(1)
			if err := image.Draw(w, 0, 0); err != nil {
				web.logger.Error(err)
			}
			return
		}

		_, err := req.Cookie("sess")

		if fwdFlag := req.FormValue("fw"); fwdFlag == "" && err != nil {

			scheme := "https"

			u, err := url.Parse(req.Referer())
			if err == nil && u.Scheme != "" {
				scheme = u.Scheme
			}

			var buf bytes.Buffer
			buf.WriteString(scheme)
			buf.WriteString("://")
			buf.WriteString(web.config.Host)
			buf.WriteString(req.RequestURI)
			buf.WriteString("&fw=1")

			w.Header().Set("Location", buf.String())
			http.SetCookie(w, initCookie(web.config.Domain))
			w.WriteHeader(302)

			return
		}

		fn(w, req)
	}
}

//TopServer http handler
func (web *Web) TopServer(w http.ResponseWriter, req *http.Request) {

	var hosts bool //hosts increment flag for Increment function

	ip := req.Header.Get("X-Real-IP")
	sessionValue := ip

	if cookie, err := req.Cookie("sess"); err == nil {
		sessionValue = cookie.Value
	}

	reqSiteID := req.FormValue("id")
	siteID, _ := strconv.Atoi(reqSiteID)
	page := req.FormValue("p")
	ref := req.FormValue("ref")
	xGeo := req.Header.Get("X-Geo")

	if err := web.historyWriter.WriteHistory(page, ref, xGeo, sessionValue, req.UserAgent(), net.ParseIP(ip), siteID); err != nil {
		web.logger.Error(err)
	}

	val, _ := web.siteMap.Get(siteID)

	image, err := web.siteMap.GetImage(val.CounterID)
	if err != nil {
		web.logger.Error(err)
		return
	}

	if ok := web.sessionPerSite.CheckSession(siteID, sessionValue); !ok {
		hosts = true
	}
	if !web.bots.BadUserAgent(req.UserAgent()) {
		val.Increment(hosts, true)
	}

	if val.DisplayDigits() {
		if err := image.Draw(w, val.Hits, val.Hosts); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			web.logger.Error(err)
		}
		return
	}
	if err := image.Draw(w, 0, 0); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		web.logger.Error(err)
	}
}

func initCookie(domain string) *http.Cookie {

	now := time.Now()
	const maxYear int = 15
	const cookieLen int = 14

	return &http.Cookie{
		Name:    "sess",
		Value:   session.UUID(cookieLen),
		Path:    "/",
		Domain:  domain,
		MaxAge:  0,
		Expires: time.Date(now.Year()+maxYear, now.Month(), now.Day(), 23, 59, 59, 0, time.UTC),
	}
}
