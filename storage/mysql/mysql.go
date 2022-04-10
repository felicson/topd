package mysql

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/felicson/topd/internal/config"
	"github.com/felicson/topd/storage"

	_ "github.com/go-sql-driver/mysql"
)

const createdFormat = "2006-01-02 15:04:05"

var location *time.Location

type Mysql struct {
	db *sql.DB
}

var (
	once sync.Once
	db   *sql.DB
)

func init() {
	location, _ = time.LoadLocation("Europe/Moscow")
}

//New storage constructor
func New(config config.Config) (Mysql, error) {
	var err error

	once.Do(func() {
		db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@unix(%s)/%s",
			config.DatabaseUser,
			config.DatabasePassword,
			config.DatabaseSocket,
			config.Database))
	})

	return Mysql{
		db: db,
	}, err
}

func (s Mysql) SaveData(tmpTopDataArray []storage.TopData) (err error) {

	if len(tmpTopDataArray) == 0 {
		return nil
	}

	sql := `INSERT INTO top_data (user_id, sess_id, page, refferer, date, day, ua, ip, city, country) VALUES (?,?,?,?,?,?,?,?,?,?)`
	stmt, err := s.db.Prepare(sql)
	if err != nil {
		return fmt.Errorf("on stmt prepare: %v", err)
	}
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("on begin tx: %v", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	txStmt := tx.Stmt(stmt)

	for _, row := range tmpTopDataArray {

		if _, err := txStmt.Exec(row.SiteID,
			row.Sess,
			row.Page,
			row.Referrer,
			row.Date.Format(createdFormat),
			toDays(row.Date),
			row.UA,
			row.IP.String(),
			row.City,
			row.Country,
		); err != nil {
			return fmt.Errorf("on exec tx: %v", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("on commit tx: %v", err)
	}
	return
}

func (s Mysql) Populate(lastID int) ([]storage.Site, error) {

	result, err := s.db.Query("SELECT id, counter_id, visitors, hits, show_digits FROM top_sites WHERE id > ?", lastID)
	if err != nil {
		return nil, fmt.Errorf("on populate: %v", err)
	}

	defer result.Close()

	var sites []storage.Site

	for result.Next() {

		var (
			site storage.Site
		)

		if err := result.Scan(&site.ID, &site.CounterID, &site.Hosts, &site.Hits, &site.Digits); err != nil {
			return nil, fmt.Errorf("on scan: %v", err)
		}
		sites = append(sites, site)
	}
	return sites, nil
}

func (s *Mysql) Close() error {
	return s.db.Close()
}

func toDays(t time.Time) int64 {

	//mysql analog TO_DAYS(NOW())
	t = t.In(location)
	_, offset := t.Zone()
	return ((t.Unix() + int64(offset)) / (60 * 60 * 24)) + 719528
}
