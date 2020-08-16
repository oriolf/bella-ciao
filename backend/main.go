package main

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"

	_ "github.com/mattn/go-sqlite3"

	"github.com/dgrijalva/jwt-go"
)

var (
	appHandlers = map[string]func(http.ResponseWriter, *http.Request){
		"/uninitialized": handler(NoToken, noParams, Uninitialized),
		"/initialize":    handler(NoToken, InitializeParams, Initialize),

		"/auth/register": handler(NoToken, RegisterParams, Register),
		"/auth/login":    handler(NoToken, LoginParams, Login),

		"/users/files/own":      handler(UserToken, noParams, GetOwnFiles),
		"/users/files/delete":   handler(FileOwnerOrAdminToken, IDParams, DeleteFile),
		"/users/files/download": handler(FileOwnerOrAdminToken, IDParams, DownloadFile),
		"/users/files/upload":   handler(UserToken, UploadFileParams, UploadFile),

		"/users/unvalidated/get": handler(AdminToken, noParams, GetUnvalidatedUsers),
		"/users/validated/get":   handler(AdminToken, noParams, GetValidatedUsers),
		"/users/messages/add":    handler(AdminToken, AddMessageParams, AddMessage),
		"/users/messages/own":    handler(UserToken, noParams, GetOwnMessages),
		"/users/messages/solve":  handler(MessageOwnerOrAdminToken, IDParams, SolveMessage),
		"/users/validate":        handler(AdminToken, IDParams, ValidateUser),

		"/candidates/get":    handler(NoToken, noParams, GetCandidates),
		"/candidates/add":    handler(AdminToken, AddCandidateParams, AddCandidate),
		"/candidates/delete": handler(AdminToken, IDParams, DeleteCandidate),

		"/elections/get":     handler(NoToken, noParams, GetElections),
		"/elections/publish": handler(AdminToken, IDParams, PublishElection),
		// TODO store allowed identification types and count methods as part of the initialization
		// TODO validate (and test) that unique IDs and count methods are one of the allowed
		// TODO implement and test /config/update (for global options), /elections/update, /elections/vote, etc.
	}

	initialized struct {
		value bool
		mutex sync.Mutex
	}
)

func main() {
	bootstrap()

	log.Println("Completed bootstrap, start listening...")
	log.Fatalln(http.ListenAndServe(":9876", nil))
}

func bootstrap() {
	db, err := sql.Open("sqlite3", DB_FILE)
	if err != nil {
		log.Fatalln("Error during database connection:", err)
	}
	defer db.Close()

	if err := InitDB(db); err != nil {
		log.Fatalln("Error during database initialization:", err)
	}

	checkInitialized(db)

	if _, err := os.Stat(UPLOADS_FOLDER); err != nil {
		if err := os.Mkdir(UPLOADS_FOLDER, 0755); err != nil {
			log.Fatalln("Could not create uploads folder:", err)
		}
	}

	for path, handler := range appHandlers {
		http.HandleFunc(path, handler)
	}
}

func checkInitialized(db *sql.DB) {
	count, err := countElections(db)
	if err != nil {
		log.Fatalln("Could not count elections in check initialized:", err)
	}

	if count > 0 {
		initialized.value = true
	}
}

func getInitialized() bool {
	initialized.mutex.Lock()
	defer initialized.mutex.Unlock()
	return initialized.value
}

func handler(
	tokenFunc func(*sql.DB, *http.Request, interface{}) (*jwt.Token, *Claims, error),
	paramsFunc func(*http.Request) (interface{}, error),
	handleFunc func(http.ResponseWriter, *sql.DB, *jwt.Token, *Claims, interface{}) error,
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if initialized := getInitialized(); !initialized {
			if r.URL.Path != "/initialize" && r.URL.Path != "/uninitialized" {
				log.Printf("Invalid method %q before initialization\n", r.URL.Path)
				http.Error(w, "", http.StatusUnauthorized)
				return
			}
		}

		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}

		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "X-Requested-With,content-type,Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			return
		}

		log.Println("Received petition to", r.URL.Path)
		params, err := paramsFunc(r)
		if err != nil {
			log.Println("Error validating parameters:", err)
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		db, err := sql.Open("sqlite3", DB_FILE)
		if err != nil {
			log.Fatalln("Error during database connection in handler:", err)
		}
		defer db.Close()

		token, claims, err := tokenFunc(db, r, params) // token func validates permissions too
		if err != nil {
			log.Println("Error validating token:", err)
			http.Error(w, "", http.StatusUnauthorized)
			return
		}

		if err := handleFunc(w, db, token, claims, params); err != nil {
			log.Println("Error handling request:", err)
			http.Error(w, "", http.StatusInternalServerError)
		}
	}
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

func noParams(*http.Request) (interface{}, error) {
	return nil, nil
}
