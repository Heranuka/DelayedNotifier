// Package postgres provides a PostgreSQL client adapter for the application.
package postgres

import (
	pgxdriver "github.com/wb-go/wbf/dbpg/pgx-driver"
	"github.com/wb-go/wbf/logger"
)

// New creates a new PostgreSQL client.
func New(dsn string, log logger.Logger) (*pgxdriver.Postgres, error) {
	return pgxdriver.New(dsn, log)
}
