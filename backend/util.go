package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

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
		return nil, wrapError(err, 31, "could not get session")
	}

	id, ok := session.Values["user_id"]
	if !ok {
		return nil, traceError{id: 1, message: "did not find user_id key"}
	}

	userID, ok := id.(int)
	if !ok {
		return nil, traceError{id: 2, message: "wrong type for user_id"}
	}

	user, err := getUser(tx, userID)
	if err != nil {
		return nil, wrapError(err, 32, "could not get user")
	}

	return &user, err
}

func noLogin(db *sql.Tx, user *User, values par.Values, err error) error {
	return nil
}

func requireLogin(db *sql.Tx, user *User, values par.Values, err error) error {
	return err
}

func validatedUser(db *sql.Tx, user *User, values par.Values, err error) error {
	if user.Role == ROLE_NONE {
		return traceError{id: 3, message: "none role"}
	}

	return nil
}

func adminUser(db *sql.Tx, user *User, values par.Values, err error) error {
	if user.Role != ROLE_ADMIN {
		return traceError{id: 4, message: "non admin role"}
	}

	return nil
}

func fileOwnerOrAdminUser(db *sql.Tx, user *User, values par.Values, err error) error {
	if user.Role != ROLE_ADMIN {
		if err := checkFileOwnedByUser(db, values.Int("id"), user.ID); err != nil {
			return wrapError(err, 33, "not admin and file not owned")
		}
	}

	return nil
}

func messageOwnerOrAdminUser(db *sql.Tx, user *User, values par.Values, err error) error {
	if user.Role != ROLE_ADMIN {
		if err := checkMessageOwnedByUser(db, values.Int("id"), user.ID); err != nil {
			return wrapError(err, 34, "not admin and message not owned")
		}
	}

	return nil
}

func electionDidNotStart(db *sql.Tx, user *User, values par.Values, err error) error {
	elections, err := getElections(db, false)
	if err != nil {
		return wrapError(err, 35, "could not get elections")
	}

	if len(elections) != 1 {
		return traceError{id: 5, message: "expected just one election"}
	}

	e := elections[0]
	if now().After(e.Start) {
		return traceError{id: 6, message: "election already started"}
	}

	return nil
}

func validIDFormats(db *sql.Tx, user *User, values par.Values, err error) error {
	uniqueID := values.String("unique_id")
	config, err := getConfig(db)
	if err != nil {
		return wrapError(err, 36, "could not get config")
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

	return traceError{id: 7, message: "unique_id did not validate any format"}
}

func IsAdmin(user *User) bool {
	return user != nil && user.Role == "admin"
}

func GetSaltAndHashPassword(pass string) (string, string, error) {
	salt, err := SafeID()
	if err != nil {
		return "", "", wrapError(err, 37, "could not generate salt")
	}

	password, err := HashPassword(pass, salt)
	if err != nil {
		return "", "", wrapError(err, 38, "could not hash password")
	}

	return password, salt, nil
}

func SafeID() (string, error) {
	b := make([]byte, 32)
	n, err := rand.Read(b)
	if err != nil {
		return "", wrapError(err, 39, "can't read from crypto/rand")
	}
	if n != len(b) {
		return "", traceError{id: 8, message: "wrong length read from crypto/rand"}
	}

	return hex.EncodeToString(b), nil
}

func HashPassword(pass, salt string) (string, error) {
	bsalt, err := hex.DecodeString(salt)
	if err != nil {
		return "", traceError{id: 9, message: "invalid salt"}
	}
	bdk, err := scrypt.Key([]byte(pass), bsalt, 32768, 8, 1, 32)
	if err != nil {
		return "", traceError{id: 10, message: "scrypt error"}
	}
	return hex.EncodeToString(bdk), nil
}

func ValidatePassword(pass, dbpass, salt string) error {
	hashed, err := HashPassword(pass, salt)
	if err != nil {
		return wrapError(err, 40, "error hashing password")
	}

	if hashed != dbpass {
		return traceError{id: 11, message: "invalid password"}
	}

	return nil
}

func WriteResult(w http.ResponseWriter, result interface{}) error {
	js, err := json.Marshal(result)
	if err != nil {
		return wrapError(err, 41, "could not marshal result")
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if _, err := w.Write(js); err != nil {
		return wrapError(err, 42, "could not write result")
	}

	return nil
}

func validateElectionParams(v par.Values) error {
	start, end, now := v.Time("start"), v.Time("end"), now()
	if start.After(end) || end.Before(start) || start.Before(now) {
		return traceError{id: 12, message: "election should end after it starts"}
	}

	min, max := v.Int("min_candidates"), v.Int("max_candidates")
	if min > max {
		return traceError{id: 13, message: "minimum number of candidates cannot be greater than maximum"}
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
			return nil, "", wrapError(err, 43, "could not get alternative filename")
		}
		f, err := os.Create(filepath.Join(folder, alternativeName))
		return f, alternativeName, err
	} else if os.IsNotExist(err) {
		f, err := os.Create(fullPath)
		return f, filename, err
	}

	return nil, "", wrapError(err, 44, "could not check if file already existed")
}

func getAlternativeFilename(folder, filename string) (string, error) {
	files, err := ioutil.ReadDir(folder)
	if err != nil {
		return "", wrapError(err, 45, "could not read dir")
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

	return "", traceError{id: 14, message: "impossible to find alternative"}
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
		return traceError{id: 15, message: "does not validate dni format"}
	}

	return controlCharacterMatches(s)
}

func controlCharacterMatches(s string) error {
	index, err := strconv.Atoi(s[:8])
	if err != nil {
		return wrapError(err, 46, "could not sum digits")
	}

	index = index % 23
	if s[8:9] != dniLetters[index:index+1] {
		return traceError{id: 16, message: "control character does not match"}
	}

	return nil
}

var nieRegex = regexp.MustCompile("^[XYZ][0-9]{7}[A-Z]$")

func validateNIE(s string) error {
	if !nieRegex.MatchString(s) {
		return traceError{id: 17, message: "does not validate nie format"}
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
		return traceError{id: 18, message: "does not validate passport format"}
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

func decodePager(p par.Values) (limit, offset int) {
	limit = p.Int("items_per_page")
	return limit, (p.Int("page") - 1) * limit
}

func now() time.Time {
	if !globalTesting {
		return time.Now()
	}

	return NOW_TEST_TIME
}

func timeTravel(d time.Duration) {
	NOW_TEST_TIME = NOW_TEST_TIME.Add(d)
}

func periodicFunc(f func(), d time.Duration) {
	c := time.Tick(5 * time.Second)
	for range c {
		f()
	}
}

// each vote is a list of candidates
// the result is a map where each candidate has its result
func countVotes(totalCandidates int, votes [][]int, countMethod string) (map[int]float64, error) {
	countFunc, ok := map[string]func(int, int) float64{
		COUNT_BORDA:   countBorda,
		COUNT_DOWDALL: countDowdall,
	}[countMethod]
	if !ok {
		return nil, traceError{id: 19, message: "unknown count method"}
	}

	points := make(map[int]float64, totalCandidates)
	for _, vote := range votes {
		for index, candidate := range vote {
			// the puntuation depends on the index inside the list and possibly on the number of candidates
			points[candidate] += countFunc(index, totalCandidates)
		}
	}

	return points, nil
}

func countBorda(index, totalCandidates int) float64 {
	return float64(totalCandidates - index) // if there are 12 candidates, 12 for the first, 11 the second... 1 for the last
}

func countDowdall(index, totalCandidates int) float64 {
	return 1.0 / float64(index+1) // 1 for the first, 0.5 for the second, 0.333... for the third, etc.
}

type traceError struct {
	id      int
	message string
	parent  error
}

func wrapError(err error, id int, message string, args ...interface{}) error {
	return traceError{parent: err, id: id, message: fmt.Sprintf(message, args...)}
}

func (e traceError) Error() string {
	if e.parent != nil {
		return fmt.Sprintf("<%03d> %s: %s", e.id, e.message, e.parent.Error())
	}
	return fmt.Sprintf("<%03d> %s", e.id, e.message)
}

func (e traceError) Unwrap() error {
	return e.parent
}

func frontendError(err error) string {
	var trace string
	for ce, ok := err.(traceError); err != nil; ce, ok = err.(traceError) {
		if ok {
			trace += fmt.Sprintf("%03d", ce.id)
		}

		u, ok := err.(interface{ Unwrap() error })
		if ok {
			err = u.Unwrap()
		} else {
			err = nil
		}
	}

	return commitHash + "-" + trace
}
