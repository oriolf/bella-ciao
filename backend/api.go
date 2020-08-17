package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/dgrijalva/jwt-go"
	"github.com/oriolf/bella-ciao/params"
)

func Uninitialized(w http.ResponseWriter, db *sql.DB, token *jwt.Token, claims *Claims, p par.Values) error {
	if initialized := getInitialized(); initialized {
		return errors.New("already initialized")
	}

	return nil // only returns OK if not initialized
}

func Initialize(w http.ResponseWriter, db *sql.DB, token *jwt.Token, claims *Claims, p par.Values) error {
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
		CountType:     e.String("count_type"),
		MinCandidates: e.Int("min_candidates"),
		MaxCandidates: e.Int("max_candidates"),
	}
	if err := createElection(db, election); err != nil {
		return fmt.Errorf("could not create election: %w", err)
	}

	initialized.value = true

	return nil
}

func Register(w http.ResponseWriter, db *sql.DB, token *jwt.Token, claims *Claims, p par.Values) error {
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

func Login(w http.ResponseWriter, db *sql.DB, token *jwt.Token, claims *Claims, p par.Values) error {
	user, err := getUserFromUniqueID(db, p.String("unique_id"))
	if err != nil {
		return fmt.Errorf("could not get user: %w", err)
	}

	if err := ValidatePassword(p.String("password"), user.Password, user.Salt); err != nil {
		return fmt.Errorf("invalid password: %w", err)
	}

	user.Files, user.Messages, err = getUserFilesAndMessages(db, user.ID)
	if err != nil {
		return fmt.Errorf("could not get files and messages: %w", err)
	}

	tokenString, err := GenerateToken(user)
	if err != nil {
		return fmt.Errorf("could not generate token: %w", err)
	}

	if err := WriteResult(w, tokenString); err != nil {
		return fmt.Errorf("could not write response: %w", err)
	}

	return nil
}

func GetOwnFiles(w http.ResponseWriter, db *sql.DB, token *jwt.Token, claims *Claims, p par.Values) error {
	files, err := getUserFiles(db, claims.User.ID)
	if err != nil {
		return err
	}

	return WriteResult(w, files)
}

func DeleteFile(w http.ResponseWriter, db *sql.DB, token *jwt.Token, claims *Claims, p par.Values) error {
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

func DownloadFile(w http.ResponseWriter, db *sql.DB, token *jwt.Token, claims *Claims, p par.Values) error {
	filename, err := getFilename(db, p.Int("id"))
	if err != nil {
		return fmt.Errorf("could not get file name: %w", err)
	}

	http.ServeFile(w, &http.Request{URL: &url.URL{}}, filepath.Join(UPLOADS_FOLDER, filename))
	return nil
}

func UploadFile(w http.ResponseWriter, db *sql.DB, token *jwt.Token, claims *Claims, p par.Values) error {
	content, filename := p.File("file")
	f, filename, err := safeCreateFile(UPLOADS_FOLDER, filename)
	if err != nil {
		return fmt.Errorf("could not create file: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(content); err != nil {
		return fmt.Errorf("could not write to file: %w", err)
	}

	if err := insertFile(db, UserFile{UserID: claims.User.ID, Name: filename, Description: p.String("description")}); err != nil {
		return fmt.Errorf("could not insert file: %w", err)
	}

	return nil
}

func GetUnvalidatedUsers(w http.ResponseWriter, db *sql.DB, token *jwt.Token, claims *Claims, p par.Values) error {
	return GetUsers(w, db, token, claims, p, "users.role == 'none'")
}

func GetValidatedUsers(w http.ResponseWriter, db *sql.DB, token *jwt.Token, claims *Claims, p par.Values) error {
	return GetUsers(w, db, token, claims, p, "users.role != 'none'")
}

func GetUsers(w http.ResponseWriter, db *sql.DB, token *jwt.Token, claims *Claims, p par.Values, where string) error {
	users, err := getUsers(db, where)
	if err != nil {
		return fmt.Errorf("could not get users from db: %w", err)
	}

	return WriteResult(w, users)
}

func AddMessage(w http.ResponseWriter, db *sql.DB, token *jwt.Token, claims *Claims, p par.Values) error {
	userID, content := p.Int("user_id"), p.String("content")
	if err := addMessage(db, UserMessage{UserID: userID, Content: content}); err != nil {
		return fmt.Errorf("could not add message to db: %w", err)
	}

	return nil
}

func GetOwnMessages(w http.ResponseWriter, db *sql.DB, token *jwt.Token, claims *Claims, p par.Values) error {
	messages, err := getUserMessages(db, claims.User.ID)
	if err != nil {
		return err
	}

	return WriteResult(w, messages)
}

func SolveMessage(w http.ResponseWriter, db *sql.DB, token *jwt.Token, claims *Claims, p par.Values) error {
	if err := solveMessage(db, p.Int("id")); err != nil {
		return fmt.Errorf("could not solve message: %w", err)
	}

	return nil
}

func ValidateUser(w http.ResponseWriter, db *sql.DB, token *jwt.Token, claims *Claims, p par.Values) error {
	if err := validateUser(db, p.Int("id")); err != nil {
		return fmt.Errorf("could not validate user: %w", err)
	}

	return nil
}

func GetCandidates(w http.ResponseWriter, db *sql.DB, token *jwt.Token, claims *Claims, params par.Values) error {
	candidates, err := getCandidates(db, 1)
	if err != nil {
		return fmt.Errorf("could not get candidates: %w", err)
	}

	if err := WriteResult(w, candidates); err != nil {
		return fmt.Errorf("could not write response: %w", err)
	}

	return nil
}

func AddCandidate(w http.ResponseWriter, db *sql.DB, token *jwt.Token, claims *Claims, p par.Values) error {
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

func DeleteCandidate(w http.ResponseWriter, db *sql.DB, token *jwt.Token, claims *Claims, p par.Values) error {
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

func GetElections(w http.ResponseWriter, db *sql.DB, token *jwt.Token, claims *Claims, params par.Values) error {
	elections, err := getElections(db, !IsAdmin(claims)) // all non-admin get only public elections
	if err != nil {
		return fmt.Errorf("could not get elections: %w", err)
	}

	if err := WriteResult(w, elections); err != nil {
		return fmt.Errorf("could not write response: %w", err)
	}

	return nil
}

func PublishElection(w http.ResponseWriter, db *sql.DB, token *jwt.Token, claims *Claims, p par.Values) error {
	if err := publishElection(db, p.Int("id")); err != nil {
		return fmt.Errorf("could not delete candidate: %w", err)
	}

	return nil
}
