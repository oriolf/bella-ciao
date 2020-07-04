package main

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func InitDB(db *sql.DB) error {
	types := []DBType{
		User{},
		UserFile{},
		UserMessage{},
		Election{},
		Candidate{},
		Vote{},
	}
	for i, table := range types {
		if _, err := db.Exec(table.CreateTableQuery()); err != nil {
			return fmt.Errorf("error executing init query %d: %w", i, err)
		}
	}

	return nil
}

func queryDB(db *sql.DB, scanFunc func(rows *sql.Rows) (interface{}, error), stmt string, args ...interface{}) ([]interface{}, error) {
	rows, err := db.Query(stmt, args...)
	if err != nil {
		return nil, fmt.Errorf("error during query: %w", err)
	}
	defer rows.Close()

	res := []interface{}{}
	i := 0
	for rows.Next() {
		x, err := scanFunc(rows)
		if err != nil {
			return nil, fmt.Errorf("error scaning row %d: %w", i, err)
		}
		res = append(res, x)
		i += 1
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("final error in rows: %w", err)
	}

	return res, nil
}

func countElections(db *sql.DB) (int, error) { return countDB(db, "SELECT COUNT(1) FROM elections;") }
func countAdminUsers(db *sql.DB) (int, error) {
	return countDB(db, "SELECT COUNT(1) FROM users WHERE role LIKE ?;", ROLE_ADMIN)
}

func checkFileOwnedByUser(db *sql.DB, fileID, userID int) error {
	return resourceOwnedByUser(db, "files", fileID, userID)
}

func checkMessageOwnedByUser(db *sql.DB, fileID, userID int) error {
	return resourceOwnedByUser(db, "messages", fileID, userID)
}

func resourceOwnedByUser(db *sql.DB, resourceName string, resourceID, userID int) error {
	query := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE id = ? AND user_id = ?;", resourceName)
	n, err := countDB(db, query, resourceID, userID)
	if err != nil {
		return fmt.Errorf("error counting: %w", err)
	}
	if n == 0 {
		return errors.New("resource not owned")
	}
	return nil
}

func countDB(db *sql.DB, query string, args ...interface{}) (int, error) {
	results, err := queryDB(db, scanCount, query, args...)
	if err != nil {
		return 0, fmt.Errorf("error querying count: %w", err)
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
		return nil, fmt.Errorf("error querying elections: %w", err)
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
		return nil, fmt.Errorf("error querying candidates: %w", err)
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

func AddCandidate(db *sql.DB, c Candidate) error {
	query := "INSERT INTO candidates (election_id, name, presentation, image) VALUES (1, ?, ?, ?);"
	_, err := db.Exec(query, c.Name, c.Presentation, c.Image)
	return err
}

func solveMessage(db *sql.DB, messageID int) error {
	_, err := db.Exec("UPDATE messages SET solved=1 WHERE id=?;", messageID)
	return err
}

func deleteFile(db *sql.DB, fileID int) error {
	_, err := db.Exec("DELETE FROM files WHERE id=?;", fileID)
	return err
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

func scanUnvalidatedUser(rows *sql.Rows) (interface{}, error) {
	var u unvalidatedUser
	err := rows.Scan(&u.ID, &u.UniqueID, &u.Name, &u.FileID, &u.FileDescription)
	return u, err
}

func scanCount(rows *sql.Rows) (interface{}, error) {
	var c int
	err := rows.Scan(&c)
	return c, err
}

func scanUserFile(rows *sql.Rows) (interface{}, error) {
	var f UserFile
	err := rows.Scan(&f.ID, &f.Name, &f.Description)
	return f, err
}

func scanUserMessage(rows *sql.Rows) (interface{}, error) {
	var m UserMessage
	err := rows.Scan(&m.ID, &m.Content, &m.Solved)
	return m, err
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

func getUserFilesAndMessages(db *sql.DB, id int) (files []UserFile, messages []UserMessage, err error) {
	files, err = getUserFiles(db, id)
	if err != nil {
		return nil, nil, fmt.Errorf("could not get files: %w", err)
	}

	messages, err = getUserMessages(db, id)
	if err != nil {
		return nil, nil, fmt.Errorf("could not get messages: %w", err)
	}

	return files, messages, nil
}

func getUserFiles(db *sql.DB, id int) (files []UserFile, err error) {
	fs, err := queryDB(db, scanUserFile, "SELECT id, name, description FROM files WHERE user_id = ?;", id)
	if err != nil {
		return nil, fmt.Errorf("could not select: %w", err)
	}
	for _, x := range fs {
		files = append(files, x.(UserFile))
	}
	return files, nil
}

func getUserMessages(db *sql.DB, id int) (messages []UserMessage, err error) {
	ms, err := queryDB(db, scanUserMessage, "SELECT id, content, solved FROM messages WHERE user_id = ?;", id)
	if err != nil {
		return nil, fmt.Errorf("could not get messages: %w", err)
	}
	for _, x := range ms {
		messages = append(messages, x.(UserMessage))
	}
	return messages, nil
}

func createElection(db *sql.DB, e Election) error {
	query := `INSERT INTO elections (name, date_start, date_end, public, count_type, max_candidates, min_candidates) 
              VALUES (?, ?, ?, ?, ?, ?, ?);`
	_, err := db.Exec(query, e.Name, e.Start, e.End, false, e.CountType, e.MaxCandidates, e.MinCandidates)
	return err
}

type unvalidatedUser struct {
	ID              int
	UniqueID        string
	Name            string
	FileID          *int
	FileDescription *string
}

// TODO get also user messages
func getUnvalidatedUsers(db *sql.DB) (users []User, err error) {
	query := `SELECT users.id, users.unique_id, users.name, files.id, files.description 
	FROM users LEFT JOIN files ON users.id=files.user_id 
	WHERE users.role == 'none';`

	res, err := queryDB(db, scanUnvalidatedUser, query)
	if err != nil {
		return nil, fmt.Errorf("could not query db: %w", err)
	}

	m := make(map[int]User)
	for _, x := range res {
		y, ok := x.(unvalidatedUser)
		if !ok {
			continue
		}

		u, ok := m[y.ID]
		if !ok {
			u = User{ID: y.ID, UniqueID: y.UniqueID, Name: y.Name}
		}
		if y.FileID != nil && y.FileDescription != nil {
			u.Files = append(u.Files, UserFile{ID: *y.FileID, Description: *y.FileDescription})
		}
		m[y.ID] = u
	}

	for _, x := range m {
		users = append(users, x)
	}

	return users, nil
}

func getFilename(db *sql.DB, id int) (name string, err error) {
	err = db.QueryRow("SELECT name FROM files WHERE id=?;", id).Scan(&name)
	return name, err
}
