package queries

import (
	"database/sql"

	"github.com/oriolf/bella-ciao/backend/models"
	"github.com/pkg/errors"
)

func InitDB(db *sql.DB) error {
	types := []models.DBType{
		models.User{},
		models.Election{},
		models.Candidate{},
		models.Vote{},
	}
	for i, table := range types {
		if _, err := db.Exec(table.CreateTableQuery()); err != nil {
			return errors.Wrapf(err, "error executing init query %d", i)
		}
	}

	return nil
}
