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
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/oriolf/bella-ciao/params"
	"golang.org/x/crypto/scrypt"
)

var (
	fileUploadMutex sync.Mutex
)

func getRequestUser(r *http.Request, w http.ResponseWriter, tx *sql.Tx) (*User, error) {
	session, err := store.Get(r, "bella-ciao")
	if err != nil {
		session.Save(r, w) // replace old inexistent session so error does not repeat
		return nil, fmt.Errorf("could not get session: %w", err)
	}

	id, ok := session.Values["user_id"]
	if !ok {
		return nil, errors.New("did not find user_id key")
	}

	userID, ok := id.(int)
	if !ok {
		return nil, errors.New("wrong type for user_id")
	}

	user, err := getUser(tx, userID)
	if err != nil {
		return nil, fmt.Errorf("could not get user: %w", err)
	}

	return &user, err
}

func noLogin(db *sql.Tx, user *User, values par.Values, err error) error {
	return nil
}

func requireLogin(db *sql.Tx, user *User, values par.Values, err error) error {
	return err
}

func adminUser(db *sql.Tx, user *User, values par.Values, err error) error {
	if user.Role != ROLE_ADMIN {
		return errors.New("non admin role")
	}

	return nil
}

func fileOwnerOrAdminUser(db *sql.Tx, user *User, values par.Values, err error) error {
	if user.Role != ROLE_ADMIN {
		if err := checkFileOwnedByUser(db, values.Int("id"), user.ID); err != nil {
			return fmt.Errorf("not admin and file not owned: %w", err)
		}
	}

	return nil
}

func messageOwnerOrAdminUser(db *sql.Tx, user *User, values par.Values, err error) error {
	if user.Role != ROLE_ADMIN {
		if err := checkMessageOwnedByUser(db, values.Int("id"), user.ID); err != nil {
			return fmt.Errorf("not admin and message not owned: %w", err)
		}
	}

	return nil
}

func validIDFormats(db *sql.Tx, user *User, values par.Values, err error) error {
	uniqueID := values.String("unique_id")
	config, err := getConfig(db)
	if err != nil {
		return fmt.Errorf("could not get config: %w", err)
	}

	for _, idFormat := range config.IDFormats {
		f, ok := ID_VALIDATION_FUNCS[idFormat]
		if !ok {
			continue
		}

		if err := f(uniqueID); err == nil {
			return nil
		}
	}

	return errors.New("unique_id did not validate any format")
}

func IsAdmin(user *User) bool {
	return user != nil && user.Role == "admin"
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

// from https://github.com/amnesty/drupal-nif-nie-cif-validator/blob/master/includes/nif-nie-cif.php
var dniRegex = regexp.MustCompile("^[0-9]{8}[A-Z]$")
var dniLetters = "TRWAGMYFPDXBNJZSQVHLCKE"

func validateDNI(s string) error {
	if !dniRegex.MatchString(s) {
		return errors.New("does not validate dni format")
	}

	return controlCharacterMatches(s)
}

func controlCharacterMatches(s string) error {
	index, err := strconv.Atoi(s[:8])
	if err != nil {
		return fmt.Errorf("could not sum digits: %w", err)
	}

	index = index % 23
	if s[8:9] != dniLetters[index:index+1] {
		return errors.New("control character does not match")
	}

	return nil
}

var nieRegex = regexp.MustCompile("^[XYZ][0-9]{7}[A-Z]$")

func validateNIE(s string) error {
	if !nieRegex.MatchString(s) {
		return errors.New("does not validate nie format")
	}

	start := s[:8]
	start = strings.ReplaceAll(start, "X", "0")
	start = strings.ReplaceAll(start, "Y", "1")
	start = strings.ReplaceAll(start, "Z", "2")
	control := s[8:9]
	return controlCharacterMatches(start + control)
}

// from https://es.stackoverflow.com/questions/67041/validar-pasaporte-y-dni-espa%C3%B1oles
var passportRegex = regexp.MustCompile("^[A-Z]{3}[0-9]{6}[A-Z]$")

func validatePassport(s string) error {
	if !passportRegex.MatchString(s) {
		return errors.New("does not validate passport format")
	}

	return nil
}

func stringInSlice(s string, l []string) bool {
	for _, x := range l {
		if x == s {
			return true
		}
	}
	return false
}
