package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/mail"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/oriolf/bella-ciao/params"
	"golang.org/x/crypto/scrypt"
)

var (
	// TODO move to environment?
	JWTKey          = []byte("my_secret_key")
	fileUploadMutex sync.Mutex
)

func NoToken(db *sql.DB, r *http.Request, params par.Values) (*jwt.Token, *Claims, error) {
	token, claims, err := UserToken(db, r, params)
	if err != nil {
		return &jwt.Token{}, nil, nil
	}

	return token, claims, nil
}

func UserToken(db *sql.DB, r *http.Request, params par.Values) (token *jwt.Token, claims *Claims, err error) {
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

func ValidatedToken(db *sql.DB, r *http.Request, params par.Values) (*jwt.Token, *Claims, error) {
	token, claims, err := UserToken(db, r, params)
	if err != nil {
		return token, claims, fmt.Errorf("error getting token: %w", err)
	}

	if claims.Role == ROLE_NONE {
		return token, claims, errors.New("none role")
	}

	return token, claims, nil
}

func AdminToken(db *sql.DB, r *http.Request, params par.Values) (*jwt.Token, *Claims, error) {
	token, claims, err := UserToken(db, r, params)
	if err != nil {
		return token, claims, fmt.Errorf("error getting token: %w", err)
	}

	if claims.Role != ROLE_ADMIN {
		return token, claims, errors.New("non admin role")
	}

	return token, claims, nil
}

func FileOwnerOrAdminToken(db *sql.DB, r *http.Request, vals par.Values) (*jwt.Token, *Claims, error) {
	token, claims, err := UserToken(db, r, vals)
	if err != nil {
		return token, claims, fmt.Errorf("error getting token: %w", err)
	}

	if claims.Role != ROLE_ADMIN {
		if err := checkFileOwnedByUser(db, vals.Int("id"), claims.User.ID); err != nil {
			return token, claims, fmt.Errorf("not admin and file not owned: %w", err)
		}
	}

	return token, claims, nil
}

func MessageOwnerOrAdminToken(db *sql.DB, r *http.Request, vals par.Values) (*jwt.Token, *Claims, error) {
	token, claims, err := UserToken(db, r, vals)
	if err != nil {
		return token, claims, fmt.Errorf("error getting token: %w", err)
	}

	if claims.Role != ROLE_ADMIN {
		if err := checkMessageOwnedByUser(db, vals.Int("id"), claims.User.ID); err != nil {
			return token, claims, fmt.Errorf("not admin and message not owned: %w", err)
		}
	}

	return token, claims, nil
}

func IsAdmin(claims *Claims) bool {
	return claims != nil && claims.User.Role == "admin"
}

func GetParams(r *http.Request, model interface{}) error {
	if r.Body == nil {
		return errors.New("empty body")
	}

	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(model); err != nil {
		return err
	}

	return nil
}

func GetSaltAndHashPassword(pass string) (string, string, error) {
	salt, err := SafeID()
	if err != nil {
		return "", "", fmt.Errorf("could not generate salt: %w", err)
	}

	password, err := HashPassword(pass, salt)
	if err != nil {
		return "", "", fmt.Errorf("could not hash password: %w", err)
	}

	return password, salt, nil
}

func SafeID() (string, error) {
	b := make([]byte, 32)
	n, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("can't read from crypto/rand: %w", err)
	}
	if n != len(b) {
		return "", errors.New("wrong length read from crypto/rand")
	}

	return hex.EncodeToString(b), nil
}

func HashPassword(pass, salt string) (string, error) {
	bsalt, err := hex.DecodeString(salt)
	if err != nil {
		return "", errors.New("invalid salt")
	}
	bdk, err := scrypt.Key([]byte(pass), bsalt, 32768, 8, 1, 32)
	if err != nil {
		return "", errors.New("scrypt error")
	}
	return hex.EncodeToString(bdk), nil
}

func ValidatePassword(pass, dbpass, salt string) error {
	hashed, err := HashPassword(pass, salt)
	if err != nil {
		return fmt.Errorf("error hashing password: %w", err)
	}

	if hashed != dbpass {
		return errors.New("invalid password")
	}

	return nil
}

func WriteResult(w http.ResponseWriter, result interface{}) error {
	js, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("could not marshal result: %w", err)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if _, err := w.Write(js); err != nil {
		return fmt.Errorf("could not write result: %w", err)
	}

	return nil
}

func GenerateToken(user User) (string, error) {
	expirationTime := time.Now().Add(5 * time.Minute)
	claims := &Claims{
		User:           user,
		StandardClaims: jwt.StandardClaims{ExpiresAt: expirationTime.Unix()},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(JWTKey)
}

// TODO take into account user preferences
func invalidCountType(countType string) bool {
	return countType != COUNT_BORDA && countType != COUNT_DOWDALL
}

func invalidRegisterParams(params registerParamsT) (registerParamsT, bool) {
	address, err := mail.ParseAddress(params.Email)
	if err == nil {
		params.Email = address.Address
	}
	// TODO unique ID validates one of the allowed types
	return params, params.Name == "" || params.UniqueID == "" || len(params.Password) < MIN_PASSWORD_LENGTH || err != nil
}

func invalidElectionParams(params electionParams) bool {
	return params.Name == "" ||
		params.Start.IsZero() || params.End.IsZero() ||
		params.Start.After(params.End) || params.End.Before(params.Start) ||
		invalidCountType(params.CountType) ||
		params.MaxCandidates == 0 || params.MinCandidates > params.MaxCandidates
}

func safeCreateFile(folder, filename string) (*os.File, string, error) {
	fileUploadMutex.Lock()
	defer fileUploadMutex.Unlock()
	fullPath := filepath.Join(folder, filename)
	_, err := os.Stat(fullPath)
	if err == nil {
		alternativeName, err := getAlternativeFilename(folder, filename)
		if err != nil {
			return nil, "", fmt.Errorf("could not get alternative filename: %w", err)
		}
		f, err := os.Create(filepath.Join(folder, alternativeName))
		return f, alternativeName, err
	} else if os.IsNotExist(err) {
		f, err := os.Create(fullPath)
		return f, filename, err
	}

	return nil, "", fmt.Errorf("could not check if file already existed: %w", err)
}

func getAlternativeFilename(folder, filename string) (string, error) {
	files, err := ioutil.ReadDir(folder)
	if err != nil {
		return "", fmt.Errorf("could not read dir: %w", err)
	}

	filenames := make(map[string]struct{})
	for _, file := range files {
		filenames[file.Name()] = struct{}{}
	}

	base, extension := getNameAndExtension(filename)
	for i := 1; ; i++ {
		alternativeName := base + fmt.Sprintf("_%d", i) + extension
		if _, ok := filenames[alternativeName]; !ok {
			return alternativeName, nil
		}
	}

	return "", errors.New("impossible to find alternative")
}

func getNameAndExtension(filename string) (string, string) {
	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '.' {
			return filename[:i], filename[i:]
		}
	}
	return filename, ""
}

func missingFile(fileID int, files []UserFile) bool {
	for _, x := range files {
		if x.ID == fileID {
			return false
		}
	}
	return true
}

func missingMessage(messageID int, messages []UserMessage) bool {
	for _, x := range messages {
		if x.ID == messageID {
			return false
		}
	}
	return true
}
