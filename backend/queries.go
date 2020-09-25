package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
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
	HasVoted        bool
	FileID          *int
	FileDescription *string
	FileName        *string
	MessageID       *int
	MessageContent  *string
	MessageSolved   *bool
}

func InitDB(db *sql.Tx) error {
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
	err := rows.Scan(&e.ID, &e.Name, &start, &end, &e.CountMethod, &e.MaxCandidates, &e.MinCandidates, &e.Public, &e.Counted)
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

func scanVote(rows *sql.Rows) (interface{}, error) {
	var v Vote
	err := rows.Scan(&v.ID, &v.ElectionID, &v.Hash, &v.CandidatesString)
	if err != nil {
		return nil, fmt.Errorf("could not scan: %w", err)
	}

	if err := json.Unmarshal([]byte(v.CandidatesString), &v.Candidates); err != nil {
		return nil, fmt.Errorf("could not unmarshal candidates: %w", err)
	}

	v.CandidatesString = ""
	return v, nil
}

func scanCandidate(rows *sql.Rows) (interface{}, error) {
	var c Candidate
	err := rows.Scan(&c.ID, &c.ElectionID, &c.Name, &c.Presentation, &c.Image, &c.Points)
	return c, err
}

func scanQueriedUser(rows *sql.Rows) (interface{}, error) {
	var u queriedUser
	err := rows.Scan(&u.ID, &u.UniqueID, &u.Name, &u.Email, &u.Role, &u.HasVoted, &u.FileID, &u.FileDescription, &u.FileName, &u.MessageID, &u.MessageContent, &u.MessageSolved)
	return u, err
}

func scanCount(rows *sql.Rows) (interface{}, error) {
	var c int
	err := rows.Scan(&c)
	return c, err
}

func scanUser(rows *sql.Rows) (interface{}, error) {
	var u User
	err := rows.Scan(&u.ID, &u.UniqueID, &u.Name, &u.Email, &u.Role, &u.HasVoted)
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

func scanID(rows *sql.Rows) (interface{}, error) {
	var id int
	err := rows.Scan(&id)
	return id, err
}

// helpers

func queryDB(db *sql.Tx, scanFunc func(rows *sql.Rows) (interface{}, error), stmt string, args ...interface{}) ([]interface{}, error) {
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

func countDB(db *sql.Tx, query string, args ...interface{}) (int, error) {
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

func countElections(db *sql.Tx) (int, error) {
	return countDB(db, "SELECT COUNT(1) FROM elections;")
}

func countAdminUsers(db *sql.Tx) (int, error) {
	return countDB(db, "SELECT COUNT(1) FROM users WHERE role LIKE ?;", ROLE_ADMIN)
}

func RegisterUser(db *sql.Tx, user User) error {
	return registerUser(db, user, ROLE_NONE)
}

func RegisterUserAdmin(db *sql.Tx, user User) error {
	return registerUser(db, user, ROLE_ADMIN)
}

func registerUser(db *sql.Tx, user User, role string) error {
	query := fmt.Sprintf("INSERT INTO users (name, unique_id, email, password, salt, role) VALUES (?, ?, ?, ?, ?, '%s');", role)
	_, err := db.Exec(query, user.Name, user.UniqueID, user.Email, user.Password, user.Salt)
	return err
}

func createElection(db *sql.Tx, e Election) error {
	query := `INSERT INTO elections (name, date_start, date_end, public, count_method, max_candidates, min_candidates) 
              VALUES (?, ?, ?, ?, ?, ?, ?);`
	_, err := db.Exec(query, e.Name, e.Start, e.End, false, e.CountMethod, e.MaxCandidates, e.MinCandidates)
	return err
}

func createConfig(db *sql.Tx, c Config) error {
	return execConfig(db, c, `INSERT INTO config (id_formats) VALUES (?);`, "create")
}

func updateConfig(db *sql.Tx, c Config) error {
	return execConfig(db, c, `UPDATE config SET id_formats=? WHERE id=1;`, "update")
}

func execConfig(db *sql.Tx, c Config, query, action string) error {
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

func getConfig(db *sql.Tx) (c Config, err error) {
	err = db.QueryRow("SELECT id_formats FROM config WHERE id=1;").Scan(&c.IDFormatsString)
	if err != nil {
		return c, fmt.Errorf("could not query row: %w", err)
	}
	if err := json.Unmarshal([]byte(c.IDFormatsString), &c.IDFormats); err != nil {
		return c, fmt.Errorf("could not unmarshal id formats: %w", err)
	}
	return c, nil
}

func getUser(db *sql.Tx, userID int) (user User, err error) {
	err = db.QueryRow("SELECT unique_id, name, email, password, salt, role, has_voted FROM users WHERE id=?;", userID).Scan(
		&user.UniqueID, &user.Name, &user.Email, &user.Password, &user.Salt, &user.Role, &user.HasVoted)
	user.ID = userID
	return user, err
}

func getUserFromUniqueID(db *sql.Tx, uniqueID string) (user User, err error) {
	err = db.QueryRow("SELECT id, name, email, password, salt, role, has_voted FROM users WHERE unique_id LIKE ?;", uniqueID).Scan(
		&user.ID, &user.Name, &user.Email, &user.Password, &user.Salt, &user.Role, &user.HasVoted)
	user.UniqueID = uniqueID
	return user, err
}

func getUserFilesAndMessages(db *sql.Tx, id int) (files []UserFile, messages []UserMessage, err error) {
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

func getUserFiles(db *sql.Tx, id int) (files []UserFile, err error) {
	fs, err := queryDB(db, scanUserFile, "SELECT id, name, description FROM files WHERE user_id = ?;", id)
	if err != nil {
		return nil, fmt.Errorf("could not select: %w", err)
	}
	files = make([]UserFile, 0)
	for _, x := range fs {
		files = append(files, x.(UserFile))
	}
	return files, nil
}

func getFilename(db *sql.Tx, id int) (name string, err error) {
	err = db.QueryRow("SELECT name FROM files WHERE id=?;", id).Scan(&name)
	return name, err
}

func deleteFile(db *sql.Tx, fileID int) error {
	_, err := db.Exec("DELETE FROM files WHERE id=?;", fileID)
	return err
}

func insertFile(db *sql.Tx, file UserFile) error {
	_, err := db.Exec("INSERT INTO files (user_id, name, description) VALUES (?, ?, ?);",
		file.UserID, file.Name, file.Description)
	return err
}

type getUsersResponse struct {
	Users []User `json:"users"`
	Total int    `json:"total"`
}

func getUsers(db *sql.Tx, where, query string, limit, offset int) (response getUsersResponse, err error) {
	countSQL := fmt.Sprintf("SELECT COUNT(1) FROM users WHERE %s;", where)
	var total int
	if query != "" {
		total, err = countDB(db, countSQL, query)
	} else {
		total, err = countDB(db, countSQL)
	}
	if err != nil {
		return getUsersResponse{}, fmt.Errorf("could not count users: %w", err)
	}

	sql := fmt.Sprintf(`SELECT users.id, users.unique_id, users.name, users.email, users.role, users.has_voted,
	files.id, files.description, files.name, 
	messages.id, messages.content, messages.solved
	FROM (SELECT * FROM users WHERE %s ORDER BY unique_id ASC LIMIT %d OFFSET %d) AS users 
	LEFT JOIN files ON users.id=files.user_id 
	LEFT JOIN messages ON users.id=messages.user_id;`, where, limit, offset)

	var res []interface{}
	if query != "" {
		res, err = queryDB(db, scanQueriedUser, sql, query)
	} else {
		res, err = queryDB(db, scanQueriedUser, sql)
	}
	if err != nil {
		return getUsersResponse{}, fmt.Errorf("could not query db: %w", err)
	}

	m := make(map[int]User)
	for _, x := range res {
		y, ok := x.(queriedUser)
		if !ok {
			continue
		}

		u, ok := m[y.ID]
		if !ok {
			u = User{ID: y.ID, UniqueID: y.UniqueID, Name: y.Name, Email: y.Email, HasVoted: y.HasVoted}
		}
		if y.FileID != nil && y.FileDescription != nil && y.FileName != nil {
			if missingFile(*y.FileID, u.Files) {
				u.Files = append(u.Files, UserFile{ID: *y.FileID, Description: *y.FileDescription, Name: *y.FileName})
			}
		}
		if y.MessageID != nil && y.MessageContent != nil && y.MessageSolved != nil {
			if missingMessage(*y.MessageID, u.Messages) {
				u.Messages = append(u.Messages, UserMessage{ID: *y.MessageID, Content: *y.MessageContent, Solved: *y.MessageSolved})
			}
		}
		m[y.ID] = u
	}

	var users []User
	for _, x := range m {
		users = append(users, x)
	}

	sort.Slice(users, func(i, j int) bool { return users[i].UniqueID < users[j].UniqueID })
	return getUsersResponse{Users: users, Total: total}, nil
}

func addMessage(db *sql.Tx, m UserMessage) error {
	query := "INSERT INTO messages (user_id, content, solved) VALUES (?, ?, 0);"
	_, err := db.Exec(query, m.UserID, m.Content)
	return err
}

func getUserMessages(db *sql.Tx, id int) (messages []UserMessage, err error) {
	ms, err := queryDB(db, scanUserMessage, "SELECT id, content, solved FROM messages WHERE user_id = ?;", id)
	if err != nil {
		return nil, fmt.Errorf("could not get messages: %w", err)
	}
	messages = make([]UserMessage, 0)
	for _, x := range ms {
		messages = append(messages, x.(UserMessage))
	}
	return messages, nil
}

func solveMessage(db *sql.Tx, messageID int) error {
	_, err := db.Exec("UPDATE messages SET solved=1 WHERE id=?;", messageID)
	return err
}

func validateUser(db *sql.Tx, userID int) error {
	return updateOneRecord(db, "UPDATE users SET role=? WHERE role=? AND id=?;", ROLE_VALIDATED, ROLE_NONE, userID)
}

func getCandidates(db *sql.Tx, electionID int) ([]interface{}, error) {
	return queryDB(db, scanCandidate, `SELECT id, election_id, name, presentation, image, points
	FROM candidates WHERE election_id = ? ORDER BY random();`, electionID)
}

func getCandidatesFromIDs(db *sql.Tx, ids []int) ([]Candidate, error) {
	candidateIDs := make([]string, 0, len(ids))
	for _, x := range ids {
		candidateIDs = append(candidateIDs, strconv.Itoa(x))
	}

	results, err := queryDB(db, scanCandidate, fmt.Sprintf(`
		SELECT id, election_id, name, presentation, image, points FROM candidates WHERE id IN (%s);`,
		strings.Join(candidateIDs, ",")))
	if err != nil {
		return nil, fmt.Errorf("could not get candidates: %w", err)
	}

	candidates := make([]Candidate, 0, len(results))
	for _, x := range results {
		candidates = append(candidates, x.(Candidate))
	}

	return candidates, nil
}

func updateCandidatePoints(db *sql.Tx, candidateID int, points float64) error {
	return updateOneRecord(db, "UPDATE candidates SET points=? WHERE id=?;", points, candidateID)
}

func getAvailableCandidates(db *sql.Tx, electionID int) (map[int]struct{}, error) {
	res, err := queryDB(db, scanID, `SELECT id FROM candidates WHERE election_id = ?;`, electionID)
	if err != nil {
		return nil, fmt.Errorf("could not select candidate's ids: %w", err)
	}
	m := make(map[int]struct{}, len(res))
	for _, x := range res {
		m[x.(int)] = struct{}{}
	}

	return m, nil
}

func getCandidate(db *sql.Tx, candidateID int) (Candidate, error) {
	var c Candidate
	err := db.QueryRow(`SELECT id, election_id, name, presentation, image 
	FROM candidates WHERE id = ?;`, candidateID).Scan(
		&c.ID, &c.ElectionID, &c.Name, &c.Presentation, &c.Image)
	return c, err
}

func addCandidate(db *sql.Tx, c Candidate) error {
	query := "INSERT INTO candidates (election_id, name, presentation, image) VALUES (1, ?, ?, ?);"
	_, err := db.Exec(query, c.Name, c.Presentation, c.Image)
	return err
}

func deleteCandidate(db *sql.Tx, id int) error {
	_, err := db.Exec("DELETE FROM candidates WHERE id=?;", id)
	return err
}

func getElections(db *sql.Tx, onlyPublic bool) ([]Election, error) {
	results, err := queryDB(db, scanElection, `
		SELECT id, name, date_start, date_end, count_method, max_candidates, min_candidates, public, counted
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

	results, err = queryDB(db, scanCandidate, fmt.Sprintf(`
		SELECT id, election_id, name, presentation, image, points FROM candidates WHERE election_id IN (%s) ORDER BY random();`,
		strings.Join(elIDstring, ",")))
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

func publishElection(db *sql.Tx, electionID int) error {
	return updateOneRecord(db, "UPDATE elections SET public=TRUE WHERE id=?;", electionID)
}

func setElectionCounted(db *sql.Tx, electionID int) error {
	return updateOneRecord(db, "UPDATE elections SET counted=TRUE WHERE id=?;", electionID)
}

func updateOneRecord(db *sql.Tx, query string, args ...interface{}) error {
	res, err := db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("could not execute update: %w", err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("could not count rows affected: %w", err)
	}

	if n != 1 {
		return fmt.Errorf("updated %d rows", n)
	}

	return nil
}

func setUserVoted(db *sql.Tx, userID int) error {
	return updateOneRecord(db, "UPDATE users SET has_voted=1 WHERE has_voted=0 AND id=?;", userID)
}

func insertVote(db *sql.Tx, electionID int, candidates []int, hash string) error {
	b, err := json.Marshal(candidates)
	if err != nil {
		return fmt.Errorf("could not marshal candidates: %w", err)
	}

	_, err = db.Exec("INSERT INTO votes (election_id, hash, candidates) VALUES (?, ?, ?);", electionID, hash, string(b))
	if err != nil {
		return fmt.Errorf("could not insert vote: %w", err)
	}

	return nil
}

func getVotes(db *sql.Tx, electionID int) ([]Vote, error) {
	results, err := queryDB(db, scanVote, "SELECT id, election_id, hash, candidates FROM votes WHERE election_id=?;", electionID)
	if err != nil {
		return nil, err
	}

	votes := make([]Vote, 0, len(results))
	for _, x := range results {
		votes = append(votes, x.(Vote))
	}

	return votes, nil
}

func getVoteFromHash(db *sql.Tx, hash string) (Vote, error) {
	results, err := queryDB(db, scanVote, "SELECT id, election_id, hash, candidates FROM votes WHERE hash=?;", hash)
	if err != nil {
		return Vote{}, fmt.Errorf("could not get vote: %w", err)
	}

	if len(results) != 1 {
		return Vote{}, fmt.Errorf("expected 1 vote, got %d", len(results))
	}

	return results[0].(Vote), nil
}

// params check queries

func checkFileOwnedByUser(db *sql.Tx, fileID, userID int) error {
	return resourceOwnedByUser(db, "files", fileID, userID)
}

func checkMessageOwnedByUser(db *sql.Tx, fileID, userID int) error {
	return resourceOwnedByUser(db, "messages", fileID, userID)
}

func resourceOwnedByUser(db *sql.Tx, resourceName string, resourceID, userID int) error {
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

func getAllUsers(db *sql.Tx) (users []User, err error) {
	query := "SELECT id, unique_id, name, email, role, has_voted FROM users;"
	res, err := queryDB(db, scanUser, query)
	if err != nil {
		return nil, fmt.Errorf("could not query db: %w", err)
	}

	for _, x := range res {
		users = append(users, x.(User))
	}

	return users, nil
}
