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
	ID         int64     `json:"id" sql:",pk"`
	URLKey     string    `json:"url_key"`
	QueriedAt  time.Time `json:"queried_at"`
	Successful bool      `json:"successful" sql:",notnull"`
}

// InvalidQuery tracks aggregate views for unresolved queries
type InvalidQuery struct {
	Query string `json:"query" sql:",pk"`
	Views int    `json:"views" sql:"default:0"`
}

// URLQueryMetrics contains aggregate URL query counts
type URLQueryMetrics struct {
	TotalAllTime                   int64                   `json:"total_all_time"`
	FailedAllTime                  int64                   `json:"failed_all_time"`
	SuccessPercentage              float64                 `json:"success_percentage"`
	TotalLastSevenDays             int64                   `json:"total_last_seven_days"`
	FailedLastSevenDays            int64                   `json:"failed_last_seven_days"`
	SuccessPercentageLastSevenDays float64                 `json:"success_percentage_last_seven_days"`
	DailyQueries                   []*DailyURLQueryMetrics `json:"daily_queries"`
}

// DailyURLQueryMetrics contains query counts for one calendar day
type DailyURLQueryMetrics struct {
	Date       string `json:"date"`
	Successful int64  `json:"successful"`
	Failed     int64  `json:"failed"`
	Total      int64  `json:"total"`
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
	return db.GetDB().RunInTransaction(func(tx *pg.Tx) error {
		if err := tx.Insert(u); err != nil {
			return err
		}
		_, err := tx.Model(&InvalidQuery{}).Where("query = ?", u.Key).Delete()
		return err
	})
}

// Delete removes a url from the db
func (u *URL) Delete() error {
	_, err := db.GetDB().Model(u).WherePK().Delete()
	return err
}

// IncrementViewCount increments view counts and records timestamped queries
func (u *URL) IncrementViewCount(keys []string) error {
	queries := newURLQueries(keys, time.Now().UTC(), true)
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

// GetRecent returns the most recent URL queries
func (u *URLQuery) GetRecent(limit int) ([]*URLQuery, error) {
	queries := []*URLQuery{}
	err := db.GetDB().Model(&queries).Order("queried_at DESC").Limit(limit).Select()
	return queries, err
}

// GetMetrics returns aggregate URL query counts
func (u *URLQuery) GetMetrics(timezone string) (*URLQueryMetrics, error) {
	metrics := new(URLQueryMetrics)
	_, err := db.GetDB().QueryOne(metrics, `
		WITH bounds AS (
			SELECT (
				DATE_TRUNC('day', CURRENT_TIMESTAMP AT TIME ZONE ?) - INTERVAL '6 days'
			) AT TIME ZONE ? AS start_at
		)
		SELECT
			COUNT(*) AS total_all_time,
			COUNT(*) FILTER (WHERE NOT successful) AS failed_all_time,
			COALESCE(
				ROUND(100.0 * COUNT(*) FILTER (WHERE successful) / NULLIF(COUNT(*), 0), 1),
				0
			) AS success_percentage,
			COUNT(*) FILTER (WHERE queried_at >= bounds.start_at) AS total_last_seven_days,
			COUNT(*) FILTER (
				WHERE NOT successful AND queried_at >= bounds.start_at
			) AS failed_last_seven_days,
			COALESCE(
				ROUND(
					100.0 * COUNT(*) FILTER (
						WHERE successful AND queried_at >= bounds.start_at
					) / NULLIF(
						COUNT(*) FILTER (WHERE queried_at >= bounds.start_at),
						0
					),
					1
				),
				0
			) AS success_percentage_last_seven_days
		FROM url_queries
		CROSS JOIN bounds
		GROUP BY bounds.start_at
	`, timezone, timezone)
	if err != nil {
		return metrics, err
	}
	_, err = db.GetDB().Query(&metrics.DailyQueries, `
		SELECT
			TO_CHAR(days.day, 'YYYY-MM-DD') AS date,
			COUNT(q.id) FILTER (WHERE q.successful) AS successful,
			COUNT(q.id) FILTER (WHERE NOT q.successful) AS failed,
			COUNT(q.id) AS total
		FROM GENERATE_SERIES(
			DATE_TRUNC('day', CURRENT_TIMESTAMP AT TIME ZONE ?) - INTERVAL '6 days',
			DATE_TRUNC('day', CURRENT_TIMESTAMP AT TIME ZONE ?),
			INTERVAL '1 day'
		) AS days(day)
		LEFT JOIN url_queries AS q
			ON q.queried_at >= (days.day AT TIME ZONE ?)
			AND q.queried_at < ((days.day + INTERVAL '1 day') AT TIME ZONE ?)
		GROUP BY days.day
		ORDER BY days.day
	`, timezone, timezone, timezone, timezone)
	return metrics, err
}

// IncrementInvalidViewCount tracks unresolved queries and records their history
func (u *URLQuery) IncrementInvalidViewCount(keys []string) error {
	queries := newURLQueries(keys, time.Now().UTC(), false)
	err := db.GetDB().RunInTransaction(func(tx *pg.Tx) error {
		for _, query := range queries {
			_, err := tx.Exec(`
				INSERT INTO invalid_queries (query, views) VALUES (?, 1)
				ON CONFLICT (query) DO UPDATE
				SET views = invalid_queries.views + 1
			`, query.URLKey)
			if err != nil {
				return err
			}
		}
		return tx.Insert(&queries)
	})
	if err != nil {
		log.Error("Error while recording invalid URL queries")
		log.Error(err)
	}
	return err
}

func newURLQueries(keys []string, queriedAt time.Time, successful bool) []URLQuery {
	queries := make([]URLQuery, 0, len(keys))
	seen := make(map[string]bool)
	for _, key := range keys {
		key = strings.ToLower(strings.Split(key, "/")[0])
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		queries = append(queries, URLQuery{URLKey: key, QueriedAt: queriedAt, Successful: successful})
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

// GetMostWanted gets the most frequently unresolved queries sorted by views
func (u *InvalidQuery) GetMostWanted(limit int) ([]*InvalidQuery, error) {
	queries := []*InvalidQuery{}
	_, err := db.GetDB().Query(&queries, `
		SELECT invalid.query, invalid.views
		FROM invalid_queries AS invalid
		WHERE NOT EXISTS (
			SELECT 1 FROM urls WHERE urls.key = invalid.query
		)
		ORDER BY invalid.views DESC, invalid.query
		LIMIT ?
	`, limit)
	return queries, err
}
