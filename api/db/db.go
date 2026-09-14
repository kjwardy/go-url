package db

import (
	"log"

	"github.com/kjwardy/go-url/api/config"

	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
)

var database *pg.DB

// URL ...
type URL struct {
	Key   string `sql:",pk"`
	URL   string
	Alias []string
	Views int `sql:"default:0"`
}

// Init sets up DB connection
func Init() {
	appConfig := config.GetConfig()

	db := pg.Connect(&pg.Options{
		Addr:     appConfig.Database.Addr,
		User:     appConfig.Database.User,
		Password: appConfig.Database.Pass,
		Database: appConfig.Database.Database,
	})

	database = db

	createSchema()

}

// GetDB will return the active database connection
func GetDB() *pg.DB {
	return database
}

// setupTables creates the required tables and sets up indexes
func createSchema() {
	url := URL{}
	err := database.CreateTable(&url, &orm.CreateTableOptions{
		IfNotExists: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	_, err = database.Exec(`
		CREATE TABLE IF NOT EXISTS url_queries (
			id bigserial PRIMARY KEY,
			url_key text NOT NULL REFERENCES urls (key) ON DELETE CASCADE,
			queried_at timestamptz NOT NULL DEFAULT now()
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = database.Exec(`
		CREATE INDEX IF NOT EXISTS url_queries_url_key_queried_at_idx
		ON url_queries (url_key, queried_at)
	`)
	if err != nil {
		log.Fatal(err)
	}
}
