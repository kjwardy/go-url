package db

import (
	"github.com/kjwardy/go-url/api/config"
	"github.com/kjwardy/go-url/api/logging"

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
func Init(logger *logging.Logger) error {
	appConfig := config.GetConfig()

	db := pg.Connect(&pg.Options{
		Addr:     appConfig.Database.Addr,
		User:     appConfig.Database.User,
		Password: appConfig.Database.Pass,
		Database: appConfig.Database.Database,
	})

	database = db

	if err := createSchema(logger); err != nil {
		if closeErr := database.Close(); closeErr != nil {
			logger.Event("ERROR", "Database connection cleanup failed", "database.connection", "failure", map[string]interface{}{
				"database.system":    "postgresql",
				"database.operation": "close_after_schema_failure",
				"error.type":         "database_error",
			})
		}
		database = nil
		return err
	}
	logger.Event("INFO", "Database schema initialized", "database.schema", "success", map[string]interface{}{
		"database.system": "postgresql",
	})
	return nil
}

// GetDB will return the active database connection
func GetDB() *pg.DB {
	return database
}

// setupTables creates the required tables and sets up indexes
func createSchema(logger *logging.Logger) error {
	url := URL{}
	err := database.CreateTable(&url, &orm.CreateTableOptions{
		IfNotExists: true,
	})
	if err := schemaError(logger, "create_urls_table", err); err != nil {
		return err
	}

	_, err = database.Exec(`
		CREATE TABLE IF NOT EXISTS invalid_queries (
			query text PRIMARY KEY,
			views bigint NOT NULL DEFAULT 0
		)
	`)
	if err := schemaError(logger, "create_invalid_queries_table", err); err != nil {
		return err
	}

	_, err = database.Exec(`
		CREATE TABLE IF NOT EXISTS url_queries (
			id bigserial PRIMARY KEY,
			url_key text NOT NULL,
			queried_at timestamptz NOT NULL DEFAULT now(),
			successful boolean NOT NULL DEFAULT true
		)
	`)
	if err := schemaError(logger, "create_url_queries_table", err); err != nil {
		return err
	}

	_, err = database.Exec(`
		ALTER TABLE url_queries
			ADD COLUMN IF NOT EXISTS successful boolean NOT NULL DEFAULT true;
		ALTER TABLE url_queries
			DROP CONSTRAINT IF EXISTS url_queries_url_key_fkey
	`)
	if err := schemaError(logger, "migrate_url_queries_table", err); err != nil {
		return err
	}

	_, err = database.Exec(`
		CREATE INDEX IF NOT EXISTS url_queries_url_key_queried_at_idx
		ON url_queries (url_key, queried_at)
	`)
	if err := schemaError(logger, "create_url_queries_key_index", err); err != nil {
		return err
	}

	_, err = database.Exec(`
		CREATE INDEX IF NOT EXISTS url_queries_queried_at_idx
		ON url_queries (queried_at)
	`)
	if err := schemaError(logger, "create_url_queries_time_index", err); err != nil {
		return err
	}
	return nil
}

func schemaError(logger *logging.Logger, operation string, err error) error {
	if err == nil {
		return nil
	}
	logger.Event("ERROR", "Database schema initialization failed", "database.schema", "failure", map[string]interface{}{
		"database.system":    "postgresql",
		"database.operation": operation,
		"error.type":         "database_error",
	})
	return err
}
