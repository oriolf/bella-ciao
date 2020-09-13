package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/oriolf/bella-ciao/params"
)

func Uninitialized(r *http.Request, w http.ResponseWriter, db *sql.Tx, user *User, p par.Values) error {
	if initialized := getInitialized(); initialized {
		return errors.New("already initialized")
	}

	return nil // only returns OK if not initialized
}

func Initialize(r *http.Request, w http.ResponseWriter, db *sql.Tx, u *User, p par.Values) error {
	initialized.mutex.Lock()
	defer initialized.mutex.Unlock()

	count, err := countElections(db)
	if err != nil || count > 0 {
		return errors.New("an election already exists")
	}

	count, err = countAdminUsers(db)
	if err != nil || count > 0 {
		return errors.New("an admin user already exists")
	}

	admin := p.Values("admin")
	pass, name, email, uniqueID := admin.String("password"), admin.String("name"), admin.String("email"), admin.String("unique_id")
	password, salt, err := GetSaltAndHashPassword(pass)
	if err != nil {
		return fmt.Errorf("could not get salt or hash password: %w", err)
	}

	user := User{Name: name, Email: email, UniqueID: uniqueID, Password: password, Salt: salt}
	if err := RegisterUserAdmin(db, user); err != nil {
		return fmt.Errorf("could not register user in db: %w", err)
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
		return fmt.Errorf("could not create election: %w", err)
	}

	idFormats := p.Values("config").StringList("id_formats")
	if err := createConfig(db, Config{IDFormats: idFormats}); err != nil {
		return fmt.Errorf("could not create config: %w", err)
	}

	initialized.value = true

	return nil
}

func UpdateConfig(r *http.Request, w http.ResponseWriter, db *sql.Tx, user *User, p par.Values) error {
	c, err := getConfig(db)
	if err != nil {
		return fmt.Errorf("coult not get config: %w", err)
	}

	newIDFormats := p.StringList("id_formats")
	for _, x := range c.IDFormats {
		if !stringInSlice(x, newIDFormats) { // cannot remove id formats, or else registered users could be invalid
			return fmt.Errorf("cannot remove id format %q", x)
		}
	}

	c.IDFormats = newIDFormats
	if err := updateConfig(db, c); err != nil {
		return fmt.Errorf("could not update config: %w", err)
	}

	return nil
}

func Register(r *http.Request, w http.ResponseWriter, db *sql.Tx, u *User, p par.Values) error {
	uniqueID, pass := p.String("unique_id"), p.String("password")
	name, email := p.String("name"), p.String("email")
	password, salt, err := GetSaltAndHashPassword(pass)
	if err != nil {
		return fmt.Errorf("could not get salt or hash password: %w", err)
	}

	user := User{Name: name, UniqueID: uniqueID, Email: email, Password: password, Salt: salt}
	if err := RegisterUser(db, user); err != nil {
		return fmt.Errorf("could not register user in db: %w", err)
	}

	return nil
}

func Login(r *http.Request, w http.ResponseWriter, db *sql.Tx, u *User, p par.Values) error {
	user, err := getUserFromUniqueID(db, p.String("unique_id"))
	if err != nil {
		return fmt.Errorf("could not get user: %w", err)
	}

	if err := ValidatePassword(p.String("password"), user.Password, user.Salt); err != nil {
		return fmt.Errorf("invalid password: %w", err)
	}

	session, err := store.Get(r, "bella-ciao")
	if err != nil {
		session.Save(r, w) // overwrite old inexistent session so the error does not repeat
		return fmt.Errorf("could not get session: %w", err)
	}

	session.Options.SameSite = http.SameSiteStrictMode
	session.Values["user_id"] = user.ID
	if err := session.Save(r, w); err != nil {
		return fmt.Errorf("could not save session: %w", err)
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
		return fmt.Errorf("could not get file name: %w", err)
	}

	if err := os.Remove(filepath.Join(UPLOADS_FOLDER, filename)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("could not delete file: %w", err)
	}

	if err := deleteFile(db, id); err != nil {
		return fmt.Errorf("could not delete file from db: %w", err)
	}

	return nil
}

func DownloadFile(r *http.Request, w http.ResponseWriter, db *sql.Tx, user *User, p par.Values) error {
	filename, err := getFilename(db, p.Int("id"))
	if err != nil {
		return fmt.Errorf("could not get file name: %w", err)
	}

	http.ServeFile(w, &http.Request{URL: &url.URL{}}, filepath.Join(UPLOADS_FOLDER, filename))
	return nil
}

func UploadFile(r *http.Request, w http.ResponseWriter, db *sql.Tx, user *User, p par.Values) error {
	content, filename := p.File("file")
	f, filename, err := safeCreateFile(UPLOADS_FOLDER, filename)
	if err != nil {
		return fmt.Errorf("could not create file: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(content); err != nil {
		return fmt.Errorf("could not write to file: %w", err)
	}

	if err := insertFile(db, UserFile{UserID: user.ID, Name: filename, Description: p.String("description")}); err != nil {
		return fmt.Errorf("could not insert file: %w", err)
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
		return fmt.Errorf("could not get users from db: %w", err)
	}

	return WriteResult(w, users)
}

func AddMessage(r *http.Request, w http.ResponseWriter, db *sql.Tx, user *User, p par.Values) error {
	userID, content := p.Int("user_id"), p.String("content")
	if err := addMessage(db, UserMessage{UserID: userID, Content: content}); err != nil {
		return fmt.Errorf("could not add message to db: %w", err)
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
		return fmt.Errorf("could not solve message: %w", err)
	}

	return nil
}

func ValidateUser(r *http.Request, w http.ResponseWriter, db *sql.Tx, user *User, p par.Values) error {
	if err := validateUser(db, p.Int("id")); err != nil {
		return fmt.Errorf("could not validate user: %w", err)
	}

	return nil
}

func GetCandidates(r *http.Request, w http.ResponseWriter, db *sql.Tx, user *User, params par.Values) error {
	candidates, err := getCandidates(db, 1)
	if err != nil {
		return fmt.Errorf("could not get candidates: %w", err)
	}

	if err := WriteResult(w, candidates); err != nil {
		return fmt.Errorf("could not write response: %w", err)
	}

	return nil
}

func GetCandidateImage(r *http.Request, w http.ResponseWriter, db *sql.Tx, user *User, p par.Values) error {
	candidate, err := getCandidate(db, p.Int("id"))
	if err != nil {
		return fmt.Errorf("could not get candidate: %w", err)
	}

	http.ServeFile(w, &http.Request{URL: &url.URL{}}, filepath.Join(UPLOADS_FOLDER, candidate.Image))
	return nil
}

func AddCandidate(r *http.Request, w http.ResponseWriter, db *sql.Tx, user *User, p par.Values) error {
	image, filename := p.File("image")
	f, filename, err := safeCreateFile(UPLOADS_FOLDER, filename)
	if err != nil {
		return fmt.Errorf("could not create file: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(image); err != nil {
		return fmt.Errorf("could not write to file: %w", err)
	}

	err = addCandidate(db, Candidate{Name: p.String("name"), Presentation: p.String("presentation"), Image: filename})
	if err != nil {
		return fmt.Errorf("could not add candidate: %w", err)
	}

	return nil
}

func DeleteCandidate(r *http.Request, w http.ResponseWriter, db *sql.Tx, user *User, p par.Values) error {
	id := p.Int("id")
	c, err := getCandidate(db, id)
	if err != nil {
		return fmt.Errorf("could not get candidate: %w", err)
	}

	if err := deleteCandidate(db, id); err != nil {
		return fmt.Errorf("could not delete candidate: %w", err)
	}

	if err := os.Remove(filepath.Join(UPLOADS_FOLDER, c.Image)); err != nil {
		return fmt.Errorf("could not delete candidate image: %w", err)
	}

	return nil
}

func GetElections(r *http.Request, w http.ResponseWriter, db *sql.Tx, user *User, params par.Values) error {
	elections, err := getElections(db, !IsAdmin(user)) // all non-admin get only public elections
	if err != nil {
		return fmt.Errorf("could not get elections: %w", err)
	}

	if err := WriteResult(w, elections); err != nil {
		return fmt.Errorf("could not write response: %w", err)
	}

	return nil
}

func PublishElection(r *http.Request, w http.ResponseWriter, db *sql.Tx, user *User, p par.Values) error {
	if err := publishElection(db, p.Int("id")); err != nil {
		return fmt.Errorf("could not delete candidate: %w", err)
	}

	return nil
}

func CastVote(r *http.Request, w http.ResponseWriter, db *sql.Tx, user *User, p par.Values) error {
	elections, err := getElections(db, true)
	if err != nil {
		return fmt.Errorf("could not get elections: %w", err)
	}

	if len(elections) != 1 {
		return errors.New("expected just one election")
	}

	e := elections[0]
	if now().Before(e.Start) || now().After(e.End) {
		return errors.New("out of election vote time")
	}

	if user.HasVoted {
		return errors.New("user has already voted")
	}

	candidates := p.IntList("candidates")
	if len(candidates) < e.MinCandidates || len(candidates) > e.MaxCandidates {
		return errors.New("less than min or more than max candidates")
	}

	availableCandidates, err := getAvailableCandidates(db, e.ID)
	if err != nil {
		return fmt.Errorf("could not get available candidates: %w", err)
	}

	for _, c := range candidates {
		if _, ok := availableCandidates[c]; !ok {
			return errors.New("trying to vote unexistent candidate")
		}
	}

	voteHash, err := SafeID()
	if err != nil {
		return fmt.Errorf("could not generate vote hash: %w", err)
	}

	if err := setUserVoted(db, user.ID); err != nil {
		return fmt.Errorf("could not set user voted: %w", err)
	}

	time.Sleep(time.Second) // TODO remove when ensured same user can only vote once
	if err := insertVote(db, candidates, voteHash); err != nil {
		return fmt.Errorf("could not insert vote: %w", err)
	}

	if err := WriteResult(w, voteHash); err != nil {
		return fmt.Errorf("could not write response: %w", err)
	}

	return nil
}
