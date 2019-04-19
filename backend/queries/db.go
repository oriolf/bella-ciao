package queries

import (
	"database/sql"

	"github.com/pkg/errors"
)

func InitDB(db *sql.DB) error {
	tables := []string{
		"CREATE TABLE IF NOT EXISTS users (id serial PRIMARY KEY, name TEXT, unique_id text UNIQUE, password TEXT, salt TEXT, role TEXT);",
	}

	for i, table := range tables {
		if _, err := db.Exec(table); err != nil {
			return errors.Wrapf(err, "error executing init query %d", i)
		}
	}

	return nil
}
