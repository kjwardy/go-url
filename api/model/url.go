package model

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-pg/pg"
	"github.com/kjwardy/go-url/api/db"
	"github.com/labstack/gommon/log"
)

// URL model
type URL struct {
	Key   string   `json:"key" sql:",pk"`
	URL   string   `json:"url"`
	Alias []string `json:"alias"`
	Views int      `json:"views" sql:"default:0"`
}

// URLQuery records when a URL key is queried
type URLQuery struct {
	ID        int64     `json:"id" sql:",pk"`
	URLKey    string    `json:"url_key"`
	QueriedAt time.Time `json:"queried_at"`
}

// Find returns matching URL
func (u *URL) Find(key string) (*URL, error) {
	url := new(URL)
	err := db.GetDB().Model(url).Where("key = ?", key).First()
	if err == pg.ErrNoRows {
		return nil, nil
	}
	return url, err
}

// Update sets the new url and alias fields in the db
func (u *URL) Update() error {
	url := URL{
		Key:   u.Key,
		URL:   u.URL,
		Alias: u.Alias,
	}
	_, err := db.GetDB().Model(&url).Column("url", "alias").WherePK().Update()
	return err
}

// Save adds a new url to the db
func (u *URL) Save() error {
	err := db.GetDB().Insert(u)
	return err
}

// Delete removes a url from the db
func (u *URL) Delete() error {
	_, err := db.GetDB().Model(u).WherePK().Delete()
	return err
}

// IncrementViewCount increments view counts and records timestamped queries
func (u *URL) IncrementViewCount(keys []string) error {
	queries := newURLQueries(keys, time.Now().UTC())
	queryKeys := make([]string, len(queries))
	for i, query := range queries {
		queryKeys[i] = query.URLKey
	}

	err := db.GetDB().RunInTransaction(func(tx *pg.Tx) error {
		if _, err := tx.Model(&URL{}).WhereIn("key IN (?)", pg.In(queryKeys)).Set("views = views + 1").Update(); err != nil {
			return err
		}
		return tx.Insert(&queries)
	})
	if err != nil {
		log.Error("Error while recording URL queries")
		log.Error(err)
	}
	return err
}

func newURLQueries(keys []string, queriedAt time.Time) []URLQuery {
	queries := make([]URLQuery, 0, len(keys))
	seen := make(map[string]bool)
	for _, key := range keys {
		key = strings.ToLower(strings.Split(key, "/")[0])
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		queries = append(queries, URLQuery{URLKey: key, QueriedAt: queriedAt})
	}
	return queries
}

// GetUrlsFromKeys returns all the db records that match the keys
func (u *URL) GetUrlsFromKeys(keys []string) ([]*URL, error) {
	urls := []*URL{}
	params := make(map[string][]string)
	actualKeys := make([]string, len(keys))
	for _, key := range keys {
		split := strings.Split(key, "/")
		actualKey, remaining := strings.ToLower(split[0]), split[1:]
		params[actualKey] = remaining
		actualKeys = append(actualKeys, actualKey)
	}
	err := db.GetDB().Model(&urls).WhereIn("key in (?)", pg.In(actualKeys)).Select()
	if err != nil {
		return nil, err
	}
	for _, url := range urls {
		if len(params[url.Key]) == 0 {
			continue
		}
		if url.URL != "" {
			for i, param := range params[url.Key] {
				url.URL = strings.ReplaceAll(url.URL, fmt.Sprintf("{{$%d}}", i+1), param)
			}
		} else {
			for i, alias := range url.Alias {
				url.Alias[i] = fmt.Sprintf("%s/%s", alias, strings.Join(params[url.Key], "/"))
			}
		}
	}
	return urls, nil
}

// Search returns all the db records that match the keys
func (u *URL) Search(query string, limit int) ([]*URL, error) {
	urls := []*URL{}
	err := db.GetDB().Model(&urls).Where("key LIKE ?", fmt.Sprintf("%%%s%%", query)).WhereOr("url LIKE ?", fmt.Sprintf("%%%s%%", query)).Limit(limit).Select()
	return urls, err
}

// GetMostPopular gets the urls sorted by views
func (u *URL) GetMostPopular(limit int) ([]*URL, error) {
	urls := []*URL{}
	err := db.GetDB().Model(&urls).Order("views DESC").Limit(limit).Select()
	return urls, err
}
