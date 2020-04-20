package main

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

func NoToken(r *http.Request) (*jwt.Token, *Claims, error) {
	token, claims, err := UserToken(r)
	if err != nil {
		return &jwt.Token{}, nil, nil
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

type registerParams struct {
	Name     string
	UniqueID string `json:"unique_id"`
	Password string
}

func GetRegisterParams(r *http.Request, token *jwt.Token) (interface{}, error) {
	var params registerParams
	if err := GetParams(r, &params); err != nil {
		return nil, err
	}

	// TODO unique ID validates one of the allowed types
	if params.Name == "" || params.UniqueID == "" || len(params.Password) < MIN_PASSWORD_LENGTH {
		return nil, errors.New("needed data missing")
	}

	return params, nil
}

func Register(w http.ResponseWriter, db *sql.DB, token *jwt.Token, claims *Claims, p interface{}) error {
	params, ok := p.(registerParams)
	if !ok {
		return errors.New("wrong params model")
	}

	salt, err := SafeID()
	if err != nil {
		return errors.Wrap(err, "could not generate salt")
	}

	password, err := HashPassword(params.Password, salt)
	if err != nil {
		return errors.Wrap(err, "could not hash password")
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
