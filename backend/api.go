package main

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

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
	Password string `json:"password"`
}

func NoToken(r *http.Request) (*jwt.Token, *Claims, error) {
	token, claims, err := UserToken(r)
	if err != nil {
		return &jwt.Token{}, nil, nil
	}

	return token, claims, nil
}

func UserToken(r *http.Request) (token *jwt.Token, claims *Claims, err error) {
	auth := r.Header.Get("Authorization")
	parts := strings.Split(auth, " ")
	if len(parts) < 2 {
		return token, claims, errors.New("")
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

func ValidatedToken(r *http.Request) (*jwt.Token, *Claims, error) {
	token, claims, err := UserToken(r)
	if err != nil {
		return token, claims, errors.Wrapf(err, "error getting token")
	}

	if claims.Role == ROLE_NONE {
		return token, claims, errors.New("none role")
	}

	return token, claims, nil
}

func AdminToken(r *http.Request) (*jwt.Token, *Claims, error) {
	token, claims, err := UserToken(r)
	if err != nil {
		return token, claims, errors.Wrapf(err, "error getting token")
	}

	if claims.Role != ROLE_ADMIN {
		return token, claims, errors.New("non admin role")
	}

	return token, claims, nil
}

func Initialized(w http.ResponseWriter, db *sql.DB, token *jwt.Token, claims *Claims, p interface{}) error {
	count, err := countElections(db)
	if err != nil {
		return err
	}
	if err != nil || count > 0 {
		return errors.New("already initialized")
	}

	return nil // only returns OK if not initialized
}

func GetInitializeParams(r *http.Request, token *jwt.Token) (interface{}, error) {
	var params initializeParams
	if err := GetParams(r, &params); err != nil {
		return nil, err
	}

	if invalidRegisterParams(params.Admin) || invalidElectionParams(params.Election) {
		return nil, errors.New("needed data missing")
	}

	return params, nil
}

func Initialize(w http.ResponseWriter, db *sql.DB, token *jwt.Token, claims *Claims, p interface{}) error {
	// TODO if a user admin or an election already exist, return error
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
		return errors.Wrap(err, "could not get salt or hash password")
	}

	user := User{Name: admin.Name, UniqueID: admin.UniqueID, Password: password, Salt: salt}
	if err := RegisterUserAdmin(db, user); err != nil {
		return errors.Wrap(err, "could not register user in db")
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
		return errors.Wrap(err, "could not create election")
	}

	return nil
}

func GetRegisterParams(r *http.Request, token *jwt.Token) (interface{}, error) {
	var params registerParams
	if err := GetParams(r, &params); err != nil {
		return nil, err
	}

	if invalidRegisterParams(params) {
		return nil, errors.New("needed data missing")
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
		return errors.Wrap(err, "could not get salt or hash password")
	}

	user := User{Name: params.Name, UniqueID: params.UniqueID, Password: password, Salt: salt}
	if err := RegisterUser(db, user); err != nil {
		return errors.Wrap(err, "could not register user in db")
	}

	return nil
}

func GetLoginParams(r *http.Request, token *jwt.Token) (interface{}, error) {
	var params registerParams
	if err := GetParams(r, &params); err != nil {
		return nil, err
	}

	if params.UniqueID == "" || params.Password == "" {
		return nil, errors.New("needed data missing")
	}

	return params, nil
}

func Login(w http.ResponseWriter, db *sql.DB, token *jwt.Token, claims *Claims, p interface{}) error {
	params, ok := p.(registerParams)
	if !ok {
		return errors.New("wrong params model")
	}

	user, err := GetUserFromUniqueID(db, params.UniqueID)
	if err != nil {
		return errors.Wrap(err, "could not get user")
	}

	if err := ValidatePassword(params.Password, user.Password, user.Salt); err != nil {
		return errors.Wrap(err, "invalid password")
	}

	tokenString, err := GenerateToken(user)
	if err != nil {
		return errors.Wrap(err, "could not generate token")
	}

	if err := WriteResult(w, tokenString); err != nil {
		return errors.Wrap(err, "could not write response")
	}

	return nil
}

func Refresh(w http.ResponseWriter, db *sql.DB, token *jwt.Token, claims *Claims, p interface{}) error {
	tokenString, err := GenerateToken(claims.User)
	if err != nil {
		return errors.Wrap(err, "could not generate token")
	}

	if err := WriteResult(w, tokenString); err != nil {
		return errors.Wrap(err, "could not write response")
	}

	return nil
}

func GetElectionsHandler(w http.ResponseWriter, db *sql.DB, token *jwt.Token, claims *Claims, params interface{}) error {
	elections, err := GetElections(db, !IsAdmin(claims)) // all non-admin get only public elections
	if err != nil {
		return errors.Wrap(err, "could not get elections")
	}

	if err := WriteResult(w, elections); err != nil {
		return errors.Wrap(err, "could not write response")
	}

	return nil
}

func GetCandidatesHandler(w http.ResponseWriter, db *sql.DB, token *jwt.Token, claims *Claims, params interface{}) error {
	candidates, err := GetCandidates(db, 1)
	if err != nil {
		return errors.Wrap(err, "could not get candidates")
	}

	if err := WriteResult(w, candidates); err != nil {
		return errors.Wrap(err, "could not write response")
	}

	return nil
}

type candidateParams struct {
	Name         string
	Presentation string
	Image        string
}

func GetCandidateParams(r *http.Request, token *jwt.Token) (interface{}, error) {
	var params candidateParams
	if err := GetParams(r, &params); err != nil {
		return nil, err
	}

	if params.Name == "" || params.Presentation == "" {
		return nil, errors.New("needed data missing")
	}

	return params, nil
}

func AddCandidateHandler(w http.ResponseWriter, db *sql.DB, token *jwt.Token, claims *Claims, p interface{}) error {
	params, ok := p.(candidateParams)
	if !ok {
		return errors.New("wrong params model")
	}

	err := AddCandidate(db, Candidate{Name: params.Name, Presentation: params.Presentation, Image: params.Image})
	if err != nil {
		return errors.Wrap(err, "could not add candidate")
	}

	return nil
}
