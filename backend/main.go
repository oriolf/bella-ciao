package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync"

	_ "github.com/mattn/go-sqlite3"

	"github.com/dgrijalva/jwt-go"
	"github.com/oriolf/bella-ciao/params"
)

var (
	idParams    = par.P("query").Int("id", par.PositiveInt).End()
	appHandlers = map[string]func(http.ResponseWriter, *http.Request){
		"/uninitialized": handler(NoToken, par.None(), Uninitialized),
		"/initialize":    handler(NoToken, par.Custom(InitializeParams).End(), Initialize),

		"/auth/register": handler(NoToken, par.Custom(RegisterParams).End(), Register),
		"/auth/login":    handler(NoToken, par.Custom(LoginParams).End(), Login),

		"/users/files/own":      handler(UserToken, par.None(), GetOwnFiles),
		"/users/files/delete":   handler(FileOwnerOrAdminToken, idParams, DeleteFile),
		"/users/files/download": handler(FileOwnerOrAdminToken, idParams, DownloadFile),
		"/users/files/upload":   handler(UserToken, par.Custom(UploadFileParams).End(), UploadFile),

		"/users/unvalidated/get": handler(AdminToken, par.None(), GetUnvalidatedUsers),
		"/users/validated/get":   handler(AdminToken, par.None(), GetValidatedUsers),
		"/users/messages/add":    handler(AdminToken, par.Custom(AddMessageParams).End(), AddMessage),
		"/users/messages/own":    handler(UserToken, par.None(), GetOwnMessages),
		"/users/messages/solve":  handler(MessageOwnerOrAdminToken, idParams, SolveMessage),
		"/users/validate":        handler(AdminToken, idParams, ValidateUser),

		"/candidates/get":    handler(NoToken, par.None(), GetCandidates),
		"/candidates/add":    handler(AdminToken, par.Custom(AddCandidateParams).End(), AddCandidate),
		"/candidates/delete": handler(AdminToken, idParams, DeleteCandidate),

		"/elections/get":     handler(NoToken, par.None(), GetElections),
		"/elections/publish": handler(AdminToken, idParams, PublishElection),
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
	tokenFunc func(*sql.DB, *http.Request, par.Values) (*jwt.Token, *Claims, error),
	paramsFunc func(*http.Request) (par.Values, error),
	handleFunc func(http.ResponseWriter, *sql.DB, *jwt.Token, *Claims, par.Values) error,
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
