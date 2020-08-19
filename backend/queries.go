package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type queriedUser struct {
	ID              int
	UniqueID        string
	Name            string
	Email           string
	Role            string
	FileID          *int
	FileDescription *string
	MessageID       *int
	MessageContent  *string
	MessageSolved   *bool
}

func InitDB(db *sql.DB) error {
	types := []DBType{
		User{},
		UserFile{},
		UserMessage{},
		Election{},
		Config{},
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

// scan functions

func scanElection(rows *sql.Rows) (interface{}, error) {
	var e Election
	var start, end string
	err := rows.Scan(&e.ID, &e.Name, &start, &end, &e.CountMethod, &e.MaxCandidates, &e.MinCandidates, &e.Public)
	if err != nil {
		return nil, fmt.Errorf("could not scan: %w", err)
	}

	e.Start, err = time.Parse(SQLITE_TIME_FORMAT, start)
	if err != nil {
		return nil, fmt.Errorf("could not parse start: %w", err)
	}

	e.End, err = time.Parse(SQLITE_TIME_FORMAT, end)
	if err != nil {
		return nil, fmt.Errorf("could not parse end: %w", err)
	}

	return e, nil
}

func scanCandidate(rows *sql.Rows) (interface{}, error) {
	var c Candidate
	err := rows.Scan(&c.ID, &c.ElectionID, &c.Name, &c.Presentation, &c.Image)
	return c, err
}

func scanQueriedUser(rows *sql.Rows) (interface{}, error) {
	var u queriedUser
	err := rows.Scan(&u.ID, &u.UniqueID, &u.Name, &u.Email, &u.Role, &u.FileID, &u.FileDescription, &u.MessageID, &u.MessageContent, &u.MessageSolved)
	return u, err
}

func scanCount(rows *sql.Rows) (interface{}, error) {
	var c int
	err := rows.Scan(&c)
	return c, err
}

func scanUser(rows *sql.Rows) (interface{}, error) {
	var u User
	err := rows.Scan(&u.ID, &u.UniqueID, &u.Name, &u.Email, &u.Role)
	return u, err
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

// helpers

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

// api related queries

func countElections(db *sql.DB) (int, error) {
	return countDB(db, "SELECT COUNT(1) FROM elections;")
}

func countAdminUsers(db *sql.DB) (int, error) {
	return countDB(db, "SELECT COUNT(1) FROM users WHERE role LIKE ?;", ROLE_ADMIN)
}

func RegisterUser(db *sql.DB, user User) error {
	return registerUser(db, user, ROLE_NONE)
}

func RegisterUserAdmin(db *sql.DB, user User) error {
	return registerUser(db, user, ROLE_ADMIN)
}

func registerUser(db *sql.DB, user User, role string) error {
	query := fmt.Sprintf("INSERT INTO users (name, unique_id, email, password, salt, role) VALUES (?, ?, ?, ?, ?, '%s');", role)
	_, err := db.Exec(query, user.Name, user.UniqueID, user.Email, user.Password, user.Salt)
	return err
}

func createElection(db *sql.DB, e Election) error {
	query := `INSERT INTO elections (name, date_start, date_end, public, count_method, max_candidates, min_candidates) 
              VALUES (?, ?, ?, ?, ?, ?, ?);`
	_, err := db.Exec(query, e.Name, e.Start, e.End, false, e.CountMethod, e.MaxCandidates, e.MinCandidates)
	return err
}

func createConfig(db *sql.DB, c Config) error {
	return execConfig(db, c, `INSERT INTO config (id_formats) VALUES (?);`, "create")
}

func updateConfig(db *sql.DB, c Config) error {
	return execConfig(db, c, `UPDATE config SET id_formats=? WHERE id=1;`, "update")
}

func execConfig(db *sql.DB, c Config, query, action string) error {
	b, err := json.Marshal(c.IDFormats)
	if err != nil {
		return fmt.Errorf("could not marshal id formats: %w", err)
	}
	_, err = db.Exec(query, string(b))
	if err != nil {
		return fmt.Errorf("could not %s config: %w", action, err)
	}
	return nil
}

func getConfig(db *sql.DB) (c Config, err error) {
	err = db.QueryRow("SELECT id_formats FROM config WHERE id=1;").Scan(&c.IDFormatsString)
	if err != nil {
		return c, fmt.Errorf("could not query row: %w", err)
	}
	if err := json.Unmarshal([]byte(c.IDFormatsString), &c.IDFormats); err != nil {
		return c, fmt.Errorf("could not unmarshal id formats: %w", err)
	}
	return c, nil
}

func getUserFromUniqueID(db *sql.DB, uniqueID string) (user User, err error) {
	err = db.QueryRow("SELECT id, name, email, password, salt, role FROM users WHERE unique_id LIKE ?;", uniqueID).Scan(
		&user.ID, &user.Name, &user.Email, &user.Password, &user.Salt, &user.Role)
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

func getFilename(db *sql.DB, id int) (name string, err error) {
	err = db.QueryRow("SELECT name FROM files WHERE id=?;", id).Scan(&name)
	return name, err
}

func deleteFile(db *sql.DB, fileID int) error {
	_, err := db.Exec("DELETE FROM files WHERE id=?;", fileID)
	return err
}

func insertFile(db *sql.DB, file UserFile) error {
	_, err := db.Exec("INSERT INTO files (user_id, name, description) VALUES (?, ?, ?);",
		file.UserID, file.Name, file.Description)
	return err
}

func getUsers(db *sql.DB, where string) (users []User, err error) {
	query := fmt.Sprintf(`SELECT users.id, users.unique_id, users.name, users.email, users.role, files.id, files.description, messages.id, messages.content, messages.solved
	FROM users LEFT JOIN files ON users.id=files.user_id 
	LEFT JOIN messages ON users.id=messages.user_id
	WHERE %s;`, where)

	res, err := queryDB(db, scanQueriedUser, query)
	if err != nil {
		return nil, fmt.Errorf("could not query db: %w", err)
	}

	m := make(map[int]User)
	for _, x := range res {
		y, ok := x.(queriedUser)
		if !ok {
			continue
		}

		u, ok := m[y.ID]
		if !ok {
			u = User{ID: y.ID, UniqueID: y.UniqueID, Name: y.Name}
		}
		if y.FileID != nil && y.FileDescription != nil {
			if missingFile(*y.FileID, u.Files) {
				u.Files = append(u.Files, UserFile{ID: *y.FileID, Description: *y.FileDescription})
			}
		}
		if y.MessageID != nil && y.MessageContent != nil && y.MessageSolved != nil {
			if missingMessage(*y.MessageID, u.Messages) {
				u.Messages = append(u.Messages, UserMessage{ID: *y.MessageID, Content: *y.MessageContent, Solved: *y.MessageSolved})
			}
		}
		m[y.ID] = u
	}

	for _, x := range m {
		users = append(users, x)
	}

	return users, nil
}

func addMessage(db *sql.DB, m UserMessage) error {
	query := "INSERT INTO messages (user_id, content, solved) VALUES (?, ?, 0);"
	_, err := db.Exec(query, m.UserID, m.Content)
	return err
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

func solveMessage(db *sql.DB, messageID int) error {
	_, err := db.Exec("UPDATE messages SET solved=1 WHERE id=?;", messageID)
	return err
}

func validateUser(db *sql.DB, userID int) error {
	res, err := db.Exec("UPDATE users SET role=? WHERE role=? AND id=?;", ROLE_VALIDATED, ROLE_NONE, userID)
	if err != nil {
		return fmt.Errorf("could not execute update: %w", err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("could not count rows affected: %w", err)
	}

	if n != 1 {
		return errors.New("no user updated")
	}

	return nil
}

func getCandidates(db *sql.DB, electionID int) ([]interface{}, error) {
	return queryDB(db, scanCandidate, `SELECT id, election_id, name, presentation, image 
	FROM candidates WHERE election_id = ? ORDER BY random();`, electionID)
}

func getCandidate(db *sql.DB, candidateID int) (Candidate, error) {
	var c Candidate
	err := db.QueryRow(`SELECT id, election_id, name, presentation, image 
	FROM candidates WHERE id = ?;`, candidateID).Scan(
		&c.ID, &c.ElectionID, &c.Name, &c.Presentation, &c.Image)
	return c, err
}

func addCandidate(db *sql.DB, c Candidate) error {
	query := "INSERT INTO candidates (election_id, name, presentation, image) VALUES (1, ?, ?, ?);"
	_, err := db.Exec(query, c.Name, c.Presentation, c.Image)
	return err
}

func deleteCandidate(db *sql.DB, id int) error {
	_, err := db.Exec("DELETE FROM candidates WHERE id=?;", id)
	return err
}

func getElections(db *sql.DB, onlyPublic bool) ([]Election, error) {
	results, err := queryDB(db, scanElection, `
		SELECT id, name, date_start, date_end, count_method, max_candidates, min_candidates, public 
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

func publishElection(db *sql.DB, electionID int) error {
	res, err := db.Exec("UPDATE elections SET public=TRUE WHERE id=?;", electionID)
	if err != nil {
		return fmt.Errorf("could not execute update: %w", err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("could not count rows affected: %w", err)
	}

	if n != 1 {
		return errors.New("no election updated")
	}

	return nil
}

// params check queries

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

// test checks queries

func getAllUsers(db *sql.DB) (users []User, err error) {
	query := "SELECT id, unique_id, name, email, role FROM users;"
	res, err := queryDB(db, scanUser, query)
	if err != nil {
		return nil, fmt.Errorf("could not query db: %w", err)
	}

	for _, x := range res {
		users = append(users, x.(User))
	}

	return users, nil
}
