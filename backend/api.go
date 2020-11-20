package main

import (
	"database/sql"
	"errors"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"

	"github.com/oriolf/bella-ciao/params"
)

func Uninitialized(r *http.Request, w http.ResponseWriter, db *sql.Tx, user *User, p par.Values) error {
	if initialized := getInitialized(); initialized {
		return traceError{id: 23, message: "already initialized"}
	}

	return nil // only returns OK if not initialized
}

func Initialize(r *http.Request, w http.ResponseWriter, db *sql.Tx, u *User, p par.Values) error {
	initialized.mutex.Lock()
	defer initialized.mutex.Unlock()

	count, err := countElections(db)
	if err != nil || count > 0 {
		return traceError{id: 24, message: "an election already exists"}
	}

	count, err = countAdminUsers(db)
	if err != nil || count > 0 {
		return traceError{id: 25, message: "an admin user already exists"}
	}

	admin := p.Values("admin")
	pass, name, email, uniqueID := admin.String("password"), admin.String("name"), admin.String("email"), admin.String("unique_id")
	password, salt, err := GetSaltAndHashPassword(pass)
	if err != nil {
		return wrapError(err, 47, "could not get salt or hash password")
	}

	user := User{Name: name, Email: email, UniqueID: uniqueID, Password: password, Salt: salt}
	if err := RegisterUserAdmin(db, user); err != nil {
		return wrapError(err, 48, "could not register user in db")
	}

	e := p.Values("election")
	election := Election{
		Name:          e.String("name"),
		Start:         e.Time("start"),
		End:           e.Time("end"),
		CountMethod:   e.String("count_method"),
		MinCandidates: e.Int("min_candidates"),
		MaxCandidates: e.Int("max_candidates"),
	}
	if err := createElection(db, election); err != nil {
		return wrapError(err, 49, "could not create election")
	}

	idFormats := p.Values("config").StringList("id_formats")
	if err := createConfig(db, Config{IDFormats: idFormats}); err != nil {
		return wrapError(err, 50, "could not create config")
	}

	initialized.value = true

	return nil
}

func UpdateConfig(r *http.Request, w http.ResponseWriter, db *sql.Tx, user *User, p par.Values) error {
	c, err := getConfig(db)
	if err != nil {
		return wrapError(err, 51, "coult not get config")
	}

	newIDFormats := p.StringList("id_formats")
	for _, x := range c.IDFormats {
		if !stringInSlice(x, newIDFormats) { // cannot remove id formats, or else registered users could be invalid
			return wrapError(nil, 52, "cannot remove id format %q", x)
		}
	}

	c.IDFormats = newIDFormats
	if err := updateConfig(db, c); err != nil {
		return wrapError(err, 53, "could not update config")
	}

	return nil
}

func Register(r *http.Request, w http.ResponseWriter, db *sql.Tx, u *User, p par.Values) error {
	uniqueID, pass := p.String("unique_id"), p.String("password")
	name, email := p.String("name"), p.String("email")
	password, salt, err := GetSaltAndHashPassword(pass)
	if err != nil {
		return wrapError(err, 54, "could not get salt or hash password")
	}

	user := User{Name: name, UniqueID: uniqueID, Email: email, Password: password, Salt: salt}
	if err := RegisterUser(db, user); err != nil {
		return wrapError(err, 55, "could not register user in db")
	}

	return nil
}

func Login(r *http.Request, w http.ResponseWriter, db *sql.Tx, u *User, p par.Values) error {
	user, err := getUserFromUniqueID(db, p.String("unique_id"))
	if err != nil {
		return wrapError(err, 56, "could not get user")
	}

	if err := ValidatePassword(p.String("password"), user.Password, user.Salt); err != nil {
		return wrapError(err, 57, "invalid password")
	}

	session, err := store.Get(r, "bella-ciao")
	if err != nil {
		session.Save(r, w) // overwrite old inexistent session so the error does not repeat
		return wrapError(err, 58, "could not get session")
	}

	session.Options.SameSite = http.SameSiteStrictMode
	session.Values["user_id"] = user.ID
	if err := session.Save(r, w); err != nil {
		return wrapError(err, 59, "could not save session")
	}

	return nil
}

func Logout(r *http.Request, w http.ResponseWriter, db *sql.Tx, u *User, p par.Values) error {
	session, err := store.Get(r, "bella-ciao")
	if err != nil {
		session.Save(r, w)
		return nil
	}

	delete(session.Values, "user_id")
	session.Save(r, w)

	return nil
}

func GetSelf(r *http.Request, w http.ResponseWriter, db *sql.Tx, user *User, p par.Values) error {
	return WriteResult(w, user)
}

func GetOwnFiles(r *http.Request, w http.ResponseWriter, db *sql.Tx, user *User, p par.Values) error {
	files, err := getUserFiles(db, user.ID)
	if err != nil {
		return err
	}

	return WriteResult(w, files)
}

func DeleteFile(r *http.Request, w http.ResponseWriter, db *sql.Tx, user *User, p par.Values) error {
	id := p.Int("id")
	filename, err := getFilename(db, id)
	if err != nil {
		return wrapError(err, 60, "could not get file name")
	}

	if err := os.Remove(filepath.Join(UPLOADS_FOLDER, filename)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return wrapError(err, 61, "could not delete file")
	}

	if err := deleteFile(db, id); err != nil {
		return wrapError(err, 62, "could not delete file from db")
	}

	return nil
}

func DownloadFile(r *http.Request, w http.ResponseWriter, db *sql.Tx, user *User, p par.Values) error {
	filename, err := getFilename(db, p.Int("id"))
	if err != nil {
		return wrapError(err, 63, "could not get file name")
	}

	http.ServeFile(w, &http.Request{URL: &url.URL{}}, filepath.Join(UPLOADS_FOLDER, filename))
	return nil
}

func UploadFile(r *http.Request, w http.ResponseWriter, db *sql.Tx, user *User, p par.Values) error {
	content, filename := p.File("file")
	f, filename, err := safeCreateFile(UPLOADS_FOLDER, filename)
	if err != nil {
		return wrapError(err, 64, "could not create file")
	}
	defer f.Close()

	if _, err := f.Write(content); err != nil {
		return wrapError(err, 65, "could not write to file")
	}

	if err := insertFile(db, UserFile{UserID: user.ID, Name: filename, Description: p.String("description")}); err != nil {
		return wrapError(err, 66, "could not insert file")
	}

	return nil
}

func GetUnvalidatedUsers(r *http.Request, w http.ResponseWriter, db *sql.Tx, user *User, p par.Values) error {
	return GetUsers(r, w, db, user, p, "users.role = 'none'")
}

func GetValidatedUsers(r *http.Request, w http.ResponseWriter, db *sql.Tx, user *User, p par.Values) error {
	return GetUsers(r, w, db, user, p, "users.role != 'none'")
}

func GetUsers(r *http.Request, w http.ResponseWriter, db *sql.Tx, user *User, p par.Values, where string) error {
	limit, offset := decodePager(p)
	var query string
	if q := p.String("query"); q != "" {
		query = "%" + q + "%"
		where += " AND unique_id LIKE ?"
	}
	users, err := getUsers(db, where, query, limit, offset)
	if err != nil {
		return wrapError(err, 67, "could not get users from db")
	}

	return WriteResult(w, users)
}

func AddMessage(r *http.Request, w http.ResponseWriter, db *sql.Tx, user *User, p par.Values) error {
	userID, content := p.Int("user_id"), p.String("content")
	if err := addMessage(db, UserMessage{UserID: userID, Content: content}); err != nil {
		return wrapError(err, 68, "could not add message to db")
	}

	return nil
}

func GetOwnMessages(r *http.Request, w http.ResponseWriter, db *sql.Tx, user *User, p par.Values) error {
	messages, err := getUserMessages(db, user.ID)
	if err != nil {
		return err
	}

	return WriteResult(w, messages)
}

func SolveMessage(r *http.Request, w http.ResponseWriter, db *sql.Tx, user *User, p par.Values) error {
	if err := solveMessage(db, p.Int("id")); err != nil {
		return wrapError(err, 69, "could not solve message")
	}

	return nil
}

func ValidateUser(r *http.Request, w http.ResponseWriter, db *sql.Tx, user *User, p par.Values) error {
	if err := validateUser(db, p.Int("id")); err != nil {
		return wrapError(err, 70, "could not validate user")
	}

	return nil
}

func GetCandidates(r *http.Request, w http.ResponseWriter, db *sql.Tx, user *User, params par.Values) error {
	candidates, err := getCandidates(db, 1)
	if err != nil {
		return wrapError(err, 71, "could not get candidates")
	}

	if err := WriteResult(w, candidates); err != nil {
		return wrapError(err, 72, "could not write response")
	}

	return nil
}

func GetCandidateImage(r *http.Request, w http.ResponseWriter, db *sql.Tx, user *User, p par.Values) error {
	candidate, err := getCandidate(db, p.Int("id"))
	if err != nil {
		return wrapError(err, 73, "could not get candidate")
	}

	http.ServeFile(w, &http.Request{URL: &url.URL{}}, filepath.Join(UPLOADS_FOLDER, candidate.Image))
	return nil
}

func AddCandidate(r *http.Request, w http.ResponseWriter, db *sql.Tx, user *User, p par.Values) error {
	image, filename := p.File("image")
	f, filename, err := safeCreateFile(UPLOADS_FOLDER, filename)
	if err != nil {
		return wrapError(err, 74, "could not create file")
	}
	defer f.Close()

	if _, err := f.Write(image); err != nil {
		return wrapError(err, 75, "could not write to file")
	}

	err = addCandidate(db, Candidate{Name: p.String("name"), Presentation: p.String("presentation"), Image: filename})
	if err != nil {
		return wrapError(err, 76, "could not add candidate")
	}

	return nil
}

func DeleteCandidate(r *http.Request, w http.ResponseWriter, db *sql.Tx, user *User, p par.Values) error {
	id := p.Int("id")
	c, err := getCandidate(db, id)
	if err != nil {
		return wrapError(err, 77, "could not get candidate")
	}

	if err := deleteCandidate(db, id); err != nil {
		return wrapError(err, 78, "could not delete candidate")
	}

	if err := os.Remove(filepath.Join(UPLOADS_FOLDER, c.Image)); err != nil {
		return wrapError(err, 79, "could not delete candidate image")
	}

	return nil
}

func GetElections(r *http.Request, w http.ResponseWriter, db *sql.Tx, user *User, params par.Values) error {
	elections, err := getElections(db, !IsAdmin(user)) // all non-admin get only public elections
	if err != nil {
		return wrapError(err, 80, "could not get elections")
	}

	if err := WriteResult(w, elections); err != nil {
		return wrapError(err, 81, "could not write response")
	}

	return nil
}

func CheckElections(r *http.Request, w http.ResponseWriter, db *sql.Tx, user *User, params par.Values) error {
	checkElectionsCount()
	return nil
}

func PublishElection(r *http.Request, w http.ResponseWriter, db *sql.Tx, user *User, p par.Values) error {
	if err := publishElection(db, p.Int("id")); err != nil {
		return wrapError(err, 82, "could not delete candidate")
	}

	return nil
}

func CastVote(r *http.Request, w http.ResponseWriter, db *sql.Tx, user *User, p par.Values) error {
	elections, err := getElections(db, true)
	if err != nil {
		return wrapError(err, 83, "could not get elections")
	}

	if len(elections) != 1 {
		return traceError{id: 26, message: "expected just one election"}
	}

	e := elections[0]
	if now().Before(e.Start) || now().After(e.End) {
		return traceError{id: 27, message: "out of election vote time"}
	}

	if user.HasVoted {
		return traceError{id: 28, message: "user has already voted"}
	}

	candidates := p.IntList("candidates")
	if len(candidates) < e.MinCandidates || len(candidates) > e.MaxCandidates {
		return traceError{id: 29, message: "less than min or more than max candidates"}
	}

	availableCandidates, err := getAvailableCandidates(db, e.ID)
	if err != nil {
		return wrapError(err, 84, "could not get available candidates")
	}

	for _, c := range candidates {
		if _, ok := availableCandidates[c]; !ok {
			return traceError{id: 30, message: "trying to vote unexistent candidate"}
		}
	}

	voteHash, err := SafeID()
	if err != nil {
		return wrapError(err, 85, "could not generate vote hash")
	}

	if err := setUserVoted(db, user.ID); err != nil {
		return wrapError(err, 86, "could not set user voted")
	}

	if err := insertVote(db, e.ID, candidates, voteHash); err != nil {
		return wrapError(err, 87, "could not insert vote")
	}

	if err := WriteResult(w, voteHash); err != nil {
		return wrapError(err, 88, "could not write response")
	}

	return nil
}

func CheckVote(r *http.Request, w http.ResponseWriter, db *sql.Tx, user *User, p par.Values) error {
	vote, err := getVoteFromHash(db, p.String("token"))
	if err != nil {
		return wrapError(err, 89, "could not get vote")
	}

	candidates, err := getCandidatesFromIDs(db, vote.Candidates)
	if err != nil {
		return wrapError(err, 90, "could not get candidates")
	}

	if len(candidates) != len(vote.Candidates) {
		return wrapError(nil, 91, "expected %d candidates, but got %d", len(vote.Candidates), len(candidates))
	}

	// sort the candidates the same as they where originally voted
	candidatesRank := make(map[int]int, len(candidates))
	for i, cID := range vote.Candidates {
		candidatesRank[cID] = i
	}

	sort.Slice(candidates, func(i, j int) bool {
		iRank, jRank := candidatesRank[candidates[i].ID], candidatesRank[candidates[j].ID]
		return iRank < jRank
	})

	if err := WriteResult(w, candidates); err != nil {
		return wrapError(err, 92, "could not write response")
	}

	return nil
}
