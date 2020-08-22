// TODO maybe some consts could be defined by environment variables; that would require that dart constants also come from the environment...
// TODO add function that generates constants in frontend format, so we can build backend -> generate frontend constants -> build frontend -> build docker image
package main

import (
	"strings"
	"time"
)

const (
	// ROLE_ represent the possible roles that a used has in the system
	ROLE_NONE      = "none"      // the user can log in and see public information
	ROLE_VALIDATED = "validated" // the user can vote in all elections
	ROLE_ADMIN     = "admin"     // the user can see and edit everythin

	// ID_ represent types of identification documents used for a user's unique ID
	ID_DNI      = "dni"      // spanish DNI
	ID_NIE      = "nie"      // spanish NIE
	ID_PASSPORT = "passport" // international passport

	// COUNT_ represent the available count methods for elections
	COUNT_BORDA   = "borda"   // https://en.wikipedia.org/wiki/Borda_count
	COUNT_DOWDALL = "dowdall" // https://en.wikipedia.org/wiki/Borda_count

	MIN_PASSWORD_LENGTH = 8

	UPLOADS_FOLDER  = "uploads"
	SESSIONS_FOLDER = "sessions"
	DB_FILE         = "db.db"
)

var (
	SQLITE_TIME_FORMAT = ""

	COUNT_METHODS       = []string{COUNT_BORDA, COUNT_DOWDALL}
	ID_VALIDATION_FUNCS = map[string]func(string) error{
		ID_DNI:      validateDNI,
		ID_NIE:      validateNIE,
		ID_PASSPORT: validatePassport,
	}
	ID_FORMATS = []string{ID_DNI, ID_NIE, ID_PASSPORT}
)

func init() {
	SQLITE_TIME_FORMAT = strings.Replace(time.RFC3339Nano, "T", " ", 1)
}
