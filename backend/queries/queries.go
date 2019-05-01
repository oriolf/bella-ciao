package queries

import (
	"database/sql"

	"github.com/pkg/errors"
)

func queryDB(db *sql.DB, scanFunc func(rows *sql.Rows) (interface{}, error), stmt string, args ...interface{}) ([]interface{}, error) {
	rows, err := db.Query(stmt, args...)
	if err != nil {
		return nil, errors.Wrap(err, "error during query")
	}
	defer rows.Close()

	res := []interface{}{}
	i := 0
	for rows.Next() {
		x, err := scanFunc(rows)
		if err != nil {
			return nil, errors.Wrapf(err, "error scaning row %d", i)
		}
		res = append(res, x)
		i += 1
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "final error in rows")
	}

	return res, nil
}
