package main

import (
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var errDataMissing = errors.New("needed data missing")

type initializeParams struct {
	Admin    registerParams `json:"admin"`
	Election electionParams `json:"election"`
}

type electionParams struct {
	Name          string    `json:"name"`
	Start         time.Time `json:"start"`
	End           time.Time `json:"end"`
	CountType     string    `json:"count_type"`
	MaxCandidates int       `json:"max_candidates"`
	MinCandidates int       `json:"min_candidates"`
}

type registerParams struct {
	Name     string `json:"name"`
	UniqueID string `json:"unique_id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type fileUploadParams struct {
	content     []byte
	filename    string
	description string
}

type messageParams struct {
	UserID  int    `json:"user_id"`
	Content string `json:"content"`
}

func NoToken(db *sql.DB, r *http.Request, params interface{}) (*jwt.Token, *Claims, error) {
	token, claims, err := UserToken(db, r, params)
	if err != nil {
		return &jwt.Token{}, nil, nil
	}

	return token, claims, nil
}

func UserToken(db *sql.DB, r *http.Request, params interface{}) (token *jwt.Token, claims *Claims, err error) {
	auth := r.Header.Get("Authorization")
	parts := strings.Split(auth, " ")
	if len(parts) < 2 {
		return token, claims, errors.New("no token provided")
	}

	claims = &Claims{} // required, nil claims is no use
	token, err = jwt.ParseWithClaims(parts[1], claims, func(token *jwt.Token) (interface{}, error) {
		return JWTKey, nil
	})
	if !token.Valid {
		return token, claims, errors.New("invalid token")
	}

	return token, claims, nil
}

func ValidatedToken(db *sql.DB, r *http.Request, params interface{}) (*jwt.Token, *Claims, error) {
	token, claims, err := UserToken(db, r, params)
	if err != nil {
		return token, claims, fmt.Errorf("error getting token: %w", err)
	}

	if claims.Role == ROLE_NONE {
		return token, claims, errors.New("none role")
	}

	return token, claims, nil
}

func AdminToken(db *sql.DB, r *http.Request, params interface{}) (*jwt.Token, *Claims, error) {
	token, claims, err := UserToken(db, r, params)
	if err != nil {
		return token, claims, fmt.Errorf("error getting token: %w", err)
	}

	if claims.Role != ROLE_ADMIN {
		return token, claims, errors.New("non admin role")
	}

	return token, claims, nil
}

func FileOwnerOrAdminToken(db *sql.DB, r *http.Request, params interface{}) (*jwt.Token, *Claims, error) {
	token, claims, err := UserToken(db, r, params)
	if err != nil {
		return token, claims, fmt.Errorf("error getting token: %w", err)
	}

	if claims.Role != ROLE_ADMIN {
		fileID, ok := params.(int)
		if !ok {
			return token, claims, errors.New("wrong params model")
		}
		if err := checkFileOwnedByUser(db, fileID, claims.User.ID); err != nil {
			return token, claims, fmt.Errorf("not admin and file not owned: %w", err)
		}
	}

	return token, claims, nil
}

func MessageOwnerOrAdminToken(db *sql.DB, r *http.Request, params interface{}) (*jwt.Token, *Claims, error) {
	token, claims, err := UserToken(db, r, params)
	if err != nil {
		return token, claims, fmt.Errorf("error getting token: %w", err)
	}

	if claims.Role != ROLE_ADMIN {
		messageID, ok := params.(int)
		if !ok {
			return token, claims, errors.New("wrong params model")
		}
		if err := checkMessageOwnedByUser(db, messageID, claims.User.ID); err != nil {
			return token, claims, fmt.Errorf("not admin and message not owned: %w", err)
		}
	}

	return token, claims, nil
}

func Uninitialized(w http.ResponseWriter, db *sql.DB, token *jwt.Token, claims *Claims, p interface{}) error {
	if initialized := getInitialized(); initialized {
		return errors.New("already initialized")
	}

	return nil // only returns OK if not initialized
}

func GetInitializeParams(r *http.Request) (interface{}, error) {
	var params initializeParams
	if err := GetParams(r, &params); err != nil {
		return nil, err
	}

	var invalidRegister bool
	params.Admin, invalidRegister = invalidRegisterParams(params.Admin)
	if invalidRegister || invalidElectionParams(params.Election) {
		return nil, errDataMissing
	}

	return params, nil
}

func Initialize(w http.ResponseWriter, db *sql.DB, token *jwt.Token, claims *Claims, p interface{}) error {
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

	params, ok := p.(initializeParams)
	if !ok {
		return errors.New("wrong params model")
	}

	admin := params.Admin
	password, salt, err := GetSaltAndHashPassword(admin.Password)
	if err != nil {
		return fmt.Errorf("could not get salt or hash password: %w", err)
	}

	user := User{Name: admin.Name, Email: admin.Email, UniqueID: admin.UniqueID, Password: password, Salt: salt}
	if err := RegisterUserAdmin(db, user); err != nil {
		return fmt.Errorf("could not register user in db: %w", err)
	}

	e := params.Election
	election := Election{
		Name:          e.Name,
		Start:         e.Start,
		End:           e.End,
		CountType:     e.CountType,
		MinCandidates: e.MinCandidates,
		MaxCandidates: e.MaxCandidates,
	}
	if err := createElection(db, election); err != nil {
		return fmt.Errorf("could not create election: %w", err)
	}

	initialized.value = true

	return nil
}

func GetRegisterParams(r *http.Request) (interface{}, error) {
	var params registerParams
	if err := GetParams(r, &params); err != nil {
		return nil, err
	}

	params, invalid := invalidRegisterParams(params)
	if invalid {
		return nil, errDataMissing
	}

	return params, nil
}

func GetMessageParams(r *http.Request) (interface{}, error) {
	var params messageParams
	if err := GetParams(r, &params); err != nil {
		return nil, err
	}

	if params.UserID == 0 || params.Content == "" {
		return nil, errDataMissing
	}

	return params, nil
}

func Register(w http.ResponseWriter, db *sql.DB, token *jwt.Token, claims *Claims, p interface{}) error {
	params, ok := p.(registerParams)
	if !ok {
		return errors.New("wrong params model")
	}

	password, salt, err := GetSaltAndHashPassword(params.Password)
	if err != nil {
		return fmt.Errorf("could not get salt or hash password: %w", err)
	}

	user := User{Name: params.Name, UniqueID: params.UniqueID, Email: params.Email, Password: password, Salt: salt}
	if err := RegisterUser(db, user); err != nil {
		return fmt.Errorf("could not register user in db: %w", err)
	}

	return nil
}

func GetLoginParams(r *http.Request) (interface{}, error) {
	var params registerParams
	if err := GetParams(r, &params); err != nil {
		return nil, err
	}

	if params.UniqueID == "" || params.Password == "" {
		return nil, errDataMissing
	}

	return params, nil
}

func IDParams(r *http.Request) (interface{}, error) {
	value := r.URL.Query().Get("id")
	if value == "" {
		return nil, errors.New("missing parameter")
	}

	id, err := strconv.Atoi(value)
	if err != nil {
		return nil, errors.New("id is not a number")
	}

	return id, nil
}

func Login(w http.ResponseWriter, db *sql.DB, token *jwt.Token, claims *Claims, p interface{}) error {
	params, ok := p.(registerParams)
	if !ok {
		return errors.New("wrong params model")
	}

	user, err := GetUserFromUniqueID(db, params.UniqueID)
	if err != nil {
		return fmt.Errorf("could not get user: %w", err)
	}

	if err := ValidatePassword(params.Password, user.Password, user.Salt); err != nil {
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

func GetElections(w http.ResponseWriter, db *sql.DB, token *jwt.Token, claims *Claims, params interface{}) error {
	elections, err := getElections(db, !IsAdmin(claims)) // all non-admin get only public elections
	if err != nil {
		return fmt.Errorf("could not get elections: %w", err)
	}

	if err := WriteResult(w, elections); err != nil {
		return fmt.Errorf("could not write response: %w", err)
	}

	return nil
}

func PublishElection(w http.ResponseWriter, db *sql.DB, token *jwt.Token, claims *Claims, p interface{}) error {
	id, ok := p.(int)
	if !ok {
		return errors.New("wrong params model")
	}

	if err := publishElection(db, id); err != nil {
		return fmt.Errorf("could not delete candidate: %w", err)
	}

	return nil
}

func GetCandidates(w http.ResponseWriter, db *sql.DB, token *jwt.Token, claims *Claims, params interface{}) error {
	candidates, err := getCandidates(db, 1)
	if err != nil {
		return fmt.Errorf("could not get candidates: %w", err)
	}

	if err := WriteResult(w, candidates); err != nil {
		return fmt.Errorf("could not write response: %w", err)
	}

	return nil
}

type candidateParams struct {
	Name         string
	Presentation string
	Image        string
	ImageContent []byte
}

func GetCandidateParams(r *http.Request) (interface{}, error) {
	file, handler, err := r.FormFile("image")
	if err != nil {
		return nil, fmt.Errorf("could not get file from form: %w", err)
	}
	defer file.Close()

	b, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("could not read file: %w", err)
	}

	name, presentation := r.FormValue("name"), r.FormValue("presentation")
	if name == "" || presentation == "" || handler.Filename == "" || len(b) == 0 {
		return nil, errDataMissing
	}

	return candidateParams{
		Name:         name,
		Presentation: presentation,
		Image:        handler.Filename,
		ImageContent: b,
	}, nil
}

func AddCandidate(w http.ResponseWriter, db *sql.DB, token *jwt.Token, claims *Claims, p interface{}) error {
	params, ok := p.(candidateParams)
	if !ok {
		return errors.New("wrong params model")
	}

	f, filename, err := safeCreateFile(UPLOADS_FOLDER, params.Image)
	if err != nil {
		return fmt.Errorf("could not create file: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(params.ImageContent); err != nil {
		return fmt.Errorf("could not write to file: %w", err)
	}

	err = addCandidate(db, Candidate{Name: params.Name, Presentation: params.Presentation, Image: filename})
	if err != nil {
		return fmt.Errorf("could not add candidate: %w", err)
	}

	return nil
}

func DeleteCandidate(w http.ResponseWriter, db *sql.DB, token *jwt.Token, claims *Claims, p interface{}) error {
	id, ok := p.(int)
	if !ok {
		return errors.New("wrong params model")
	}

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

func GetUnvalidatedUsers(w http.ResponseWriter, db *sql.DB, token *jwt.Token, claims *Claims, p interface{}) error {
	return GetUsers(w, db, token, claims, p, "users.role == 'none'")
}

func GetValidatedUsers(w http.ResponseWriter, db *sql.DB, token *jwt.Token, claims *Claims, p interface{}) error {
	return GetUsers(w, db, token, claims, p, "users.role != 'none'")
}

func GetUsers(w http.ResponseWriter, db *sql.DB, token *jwt.Token, claims *Claims, p interface{}, where string) error {
	users, err := getUsers(db, where)
	if err != nil {
		return fmt.Errorf("could not get users from db: %w", err)
	}

	return WriteResult(w, users)
}

func GetUploadFileParams(r *http.Request) (interface{}, error) {
	file, handler, err := r.FormFile("file")
	if err != nil {
		return nil, fmt.Errorf("could not get file from form: %w", err)
	}
	defer file.Close()

	b, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("could not read file: %w", err)
	}

	description := r.FormValue("description")
	if handler.Filename == "" || len(b) == 0 || description == "" {
		return nil, errDataMissing
	}

	return fileUploadParams{
		filename:    handler.Filename,
		content:     b,
		description: description,
	}, nil
}

func UploadFile(w http.ResponseWriter, db *sql.DB, token *jwt.Token, claims *Claims, p interface{}) error {
	par, ok := p.(fileUploadParams)
	if !ok {
		return errors.New("wrong params model")
	}

	f, filename, err := safeCreateFile(UPLOADS_FOLDER, par.filename)
	if err != nil {
		return fmt.Errorf("could not create file: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(par.content); err != nil {
		return fmt.Errorf("could not write to file: %w", err)
	}

	if err := insertFile(db, UserFile{UserID: claims.User.ID, Name: filename, Description: par.description}); err != nil {
		return fmt.Errorf("could not insert file: %w", err)
	}

	return nil
}

func DownloadFile(w http.ResponseWriter, db *sql.DB, token *jwt.Token, claims *Claims, p interface{}) error {
	id, ok := p.(int)
	if !ok {
		return errors.New("wrong params model")
	}

	filename, err := getFilename(db, id)
	if err != nil {
		return fmt.Errorf("could not get file name: %w", err)
	}

	http.ServeFile(w, &http.Request{URL: &url.URL{}}, filepath.Join(UPLOADS_FOLDER, filename))
	return nil
}

func DeleteFile(w http.ResponseWriter, db *sql.DB, token *jwt.Token, claims *Claims, p interface{}) error {
	id, ok := p.(int)
	if !ok {
		return errors.New("wrong params model")
	}

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

func AddMessage(w http.ResponseWriter, db *sql.DB, token *jwt.Token, claims *Claims, p interface{}) error {
	params, _ := p.(messageParams)
	if err := addMessage(db, UserMessage{UserID: params.UserID, Content: params.Content}); err != nil {
		return fmt.Errorf("could not add message to db: %w", err)
	}

	return nil
}

func SolveMessage(w http.ResponseWriter, db *sql.DB, token *jwt.Token, claims *Claims, p interface{}) error {
	messageID, ok := p.(int)
	if !ok {
		return errors.New("wrong params model")
	}

	if err := solveMessage(db, messageID); err != nil {
		return fmt.Errorf("could not solve message: %w", err)
	}

	return nil
}

func ValidateUser(w http.ResponseWriter, db *sql.DB, token *jwt.Token, claims *Claims, p interface{}) error {
	userID, _ := p.(int)
	if err := validateUser(db, userID); err != nil {
		return fmt.Errorf("could not validate user: %w", err)
	}

	return nil
}

func GetOwnFiles(w http.ResponseWriter, db *sql.DB, token *jwt.Token, claims *Claims, p interface{}) error {
	files, err := getUserFiles(db, claims.User.ID)
	if err != nil {
		return err
	}

	return WriteResult(w, files)
}

func GetOwnMessages(w http.ResponseWriter, db *sql.DB, token *jwt.Token, claims *Claims, p interface{}) error {
	messages, err := getUserMessages(db, claims.User.ID)
	if err != nil {
		return err
	}

	return WriteResult(w, messages)
}
