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
	// TODO use this to manage secret https://diogomonica.com/2017/03/27/why-you-shouldnt-use-env-variables-for-secret-data/
	JWTKey          = []byte("my_secret_key")
	fileUploadMutex sync.Mutex
)

func getRequestToken(r *http.Request) (token *jwt.Token, claims *Claims, err error) {
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

func noToken(db *sql.DB, claims *Claims, values par.Values, err error) error {
	return nil
}

func requireToken(db *sql.DB, claims *Claims, values par.Values, err error) error {
	return err
}

func adminToken(db *sql.DB, claims *Claims, values par.Values, err error) error {
	if claims.Role != ROLE_ADMIN {
		return errors.New("non admin role")
	}

	return nil
}

func fileOwnerOrAdminToken(db *sql.DB, claims *Claims, values par.Values, err error) error {
	if claims.Role != ROLE_ADMIN {
		if err := checkFileOwnedByUser(db, values.Int("id"), claims.User.ID); err != nil {
			return fmt.Errorf("not admin and file not owned: %w", err)
		}
	}

	return nil
}

func messageOwnerOrAdminToken(db *sql.DB, claims *Claims, values par.Values, err error) error {
	if claims.Role != ROLE_ADMIN {
		if err := checkMessageOwnedByUser(db, values.Int("id"), claims.User.ID); err != nil {
			return fmt.Errorf("not admin and message not owned: %w", err)
		}
	}

	return nil
}

func IsAdmin(claims *Claims) bool {
	return claims != nil && claims.User.Role == "admin"
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

func validateElectionParams(v par.Values) error {
	start, end := v.Time("start"), v.Time("end")
	if start.After(end) || end.Before(start) {
		return errors.New("election should end after it starts")
	}

	min, max := v.Int("min_candidates"), v.Int("max_candidates")
	if min > max {
		return errors.New("minimum number of candidates cannot be greater than maximum")
	}

	return nil
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

// TODO
func validateDNI(s string) error { return nil }

// TODO
func validateNIE(s string) error { return nil }

// TODO
func validatePASSPORT(s string) error { return nil }
