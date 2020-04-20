package main

import (
	"database/sql"

	"github.com/lib/pq"
	"github.com/pkg/errors"
)

func InitDB(db *sql.DB) error {
	types := []DBType{
		User{},
		Election{},
		Candidate{},
		Vote{},
	}
	for i, table := range types {
		if _, err := db.Exec(table.CreateTableQuery()); err != nil {
			return errors.Wrapf(err, "error executing init query %d", i)
		}
	}

	return nil
}

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

func GetElections(db *sql.DB, onlyPublic bool) ([]Election, error) {
	results, err := queryDB(db, scanElection, `
		SELECT id, name, date_start, date_end, count_type, max_candidates, min_candidates, public 
		FROM elections WHERE public OR public = $1 ORDER BY date_start ASC;`, onlyPublic)
	if err != nil {
		return nil, errors.Wrap(err, "error querying elections")
	}

	var electionIDs []int
	electionsMap := make(map[int]Election)
	for _, x := range results {
		e, _ := x.(Election)
		if e.ID == 0 {
			continue
		}
		electionsMap[e.ID] = e
		electionIDs = append(electionIDs, e.ID)
	}

	results, err = queryDB(db, scanCandidate, `
		SELECT id, election_id, name, presentation, image FROM candidates WHERE election_id=ANY($1) ORDER BY random();`,
		pq.Array(electionIDs))
	if err != nil {
		return nil, errors.Wrap(err, "error querying candidates")
	}

	for _, x := range results {
		c, _ := x.(Candidate)
		e := electionsMap[c.ElectionID]
		if c.ID == 0 || e.ID == 0 {
			continue
		}
		e.Candidates = append(e.Candidates, c)
		electionsMap[c.ElectionID] = e
	}

	var elections []Election
	for _, id := range electionIDs {
		elections = append(elections, electionsMap[id])
	}

	return elections, nil
}

func scanElection(rows *sql.Rows) (interface{}, error) {
	var e Election
	err := rows.Scan(&e.ID, &e.Name, &e.Start, &e.End, &e.CountType, &e.MaxCandidates, &e.MinCandidates, &e.Public)
	return e, err
}

func scanCandidate(rows *sql.Rows) (interface{}, error) {
	var c Candidate
	err := rows.Scan(&c.ID, &c.ElectionID, &c.Name, &c.Presentation, &c.Image)
	return c, err
}

func RegisterUser(db *sql.DB, user User) error {
	_, err := db.Exec("INSERT INTO users (name, unique_id, password, salt, role) VALUES ($1, $2, $3, $4, 'none');",
		user.Name, user.UniqueID, user.Password, user.Salt)
	return err
}

func GetUserFromUniqueID(db *sql.DB, uniqueID string) (user User, err error) {
	err = db.QueryRow("SELECT id, name, password, salt, role FROM users WHERE unique_id LIKE $1;", uniqueID).Scan(
		&user.ID, &user.Name, &user.Password, &user.Salt, &user.Role)
	user.UniqueID = uniqueID
	return user, err
}
