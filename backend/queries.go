package main

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

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

func countElections(db *sql.DB) (int, error) { return countDB(db, "SELECT COUNT(1) FROM elections;") }
func countAdminUsers(db *sql.DB) (int, error) {
	return countDB(db, "SELECT COUNT(1) FROM users WHERE role LIKE ?;", ROLE_ADMIN)
}

func countDB(db *sql.DB, query string, args ...interface{}) (int, error) {
	results, err := queryDB(db, scanCount, query, args...)
	if err != nil {
		return 0, errors.Wrap(err, "error querying count")
	}

	if len(results) != 1 {
		return 0, errors.New("unexpected number of count results")
	}

	count, ok := results[0].(int)
	if !ok {
		return 0, errors.New("unexpected count type")
	}

	return count, nil
}

func GetElections(db *sql.DB, onlyPublic bool) ([]Election, error) {
	results, err := queryDB(db, scanElection, `
		SELECT id, name, date_start, date_end, count_type, max_candidates, min_candidates, public 
		FROM elections WHERE public OR public = ? ORDER BY date_start ASC;`, onlyPublic)
	if err != nil {
		return nil, errors.Wrap(err, "error querying elections")
	}

	var electionIDs []int
	var elIDstring []string
	electionsMap := make(map[int]Election)
	for _, x := range results {
		e, _ := x.(Election)
		if e.ID == 0 {
			continue
		}
		electionsMap[e.ID] = e
		electionIDs = append(electionIDs, e.ID)
		elIDstring = append(elIDstring, strconv.Itoa(e.ID))
	}

	results, err = queryDB(db, scanCandidate, `
		SELECT id, election_id, name, presentation, image FROM candidates WHERE election_id IN (?) ORDER BY random();`,
		strings.Join(elIDstring, ","))
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

func GetCandidates(db *sql.DB, electionID int) ([]interface{}, error) {
	return queryDB(db, scanCandidate, `SELECT id, election_id, name, presentation, image 
	FROM candidates WHERE election_id = ? ORDER BY random();`, electionID)
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

func scanCount(rows *sql.Rows) (interface{}, error) {
	var c int
	err := rows.Scan(&c)
	return c, err
}

func RegisterUser(db *sql.DB, user User) error      { return registerUser(db, user, ROLE_NONE) }
func RegisterUserAdmin(db *sql.DB, user User) error { return registerUser(db, user, ROLE_ADMIN) }

func registerUser(db *sql.DB, user User, role string) error {
	query := fmt.Sprintf("INSERT INTO users (name, unique_id, password, salt, role) VALUES (?, ?, ?, ?, '%s');", role)
	_, err := db.Exec(query, user.Name, user.UniqueID, user.Password, user.Salt)
	return err
}

func GetUserFromUniqueID(db *sql.DB, uniqueID string) (user User, err error) {
	err = db.QueryRow("SELECT id, name, password, salt, role FROM users WHERE unique_id LIKE ?;", uniqueID).Scan(
		&user.ID, &user.Name, &user.Password, &user.Salt, &user.Role)
	user.UniqueID = uniqueID
	return user, err
}

func createElection(db *sql.DB, e Election) error {
	query := `INSERT INTO elections (name, date_start, date_end, public, count_type, max_candidates, min_candidates) 
              VALUES (?, ?, ?, ?, ?, ?, ?);`
	_, err := db.Exec(query, e.Name, e.Start, e.End, false, e.CountType, e.MaxCandidates, e.MinCandidates)
	return err
}
