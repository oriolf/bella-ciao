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
	idParams = par.P("query").Int("id", par.PositiveInt).End()
	noParams = par.None()

	loginParams = par.P("json").
			String("unique_id", par.NonEmpty).
			String("password", par.NonEmpty).End()

	registerParamsAux = par.P("json").
				String("name", par.NonEmpty).
				String("unique_id", par.NonEmpty). // TODO also check unique_id in list of all possible
				Email("email").
				String("password", par.MinLength(MIN_PASSWORD_LENGTH))
	registerParams = registerParamsAux.End()

	addMessageParams = par.P("json").
				Int("user_id", par.PositiveInt).
				String("content", par.NonEmpty).End()

	uploadFileParams = par.P("form").
				File("file").
				String("description", par.NonEmpty).End()

	addCandidateParams = par.P("form").
				File("image").
				String("name", par.NonEmpty).
				String("presentation", par.NonEmpty).End()

	electionParamsAux = par.P("json").
				String("name", par.NonEmpty).
				Time("start", par.NonZeroTime).
				Time("end", par.NonZeroTime).
				String("count_type", par.NonEmpty). // TODO also check count_type in list of all possible
				Int("min_candidates", par.PositiveInt).
				Int("max_candidates", par.PositiveInt).
				ValidateFunc(validateElectionParams)

	initializeParams = par.P("json").
				JSON("admin", registerParamsAux.EndJSON()).
				JSON("election", electionParamsAux.EndJSON()).End()

	appHandlers = map[string]func(http.ResponseWriter, *http.Request){
		"/uninitialized": handler(noParams, NoToken, Uninitialized),
		"/initialize":    handler(initializeParams, NoToken, Initialize),

		"/auth/register": handler(registerParams, NoToken, Register),
		"/auth/login":    handler(loginParams, NoToken, Login),

		"/users/files/own":      handler(noParams, UserToken, GetOwnFiles),
		"/users/files/delete":   handler(idParams, FileOwnerOrAdminToken, DeleteFile),
		"/users/files/download": handler(idParams, FileOwnerOrAdminToken, DownloadFile),
		"/users/files/upload":   handler(uploadFileParams, UserToken, UploadFile),

		"/users/unvalidated/get": handler(noParams, AdminToken, GetUnvalidatedUsers),
		"/users/validated/get":   handler(noParams, AdminToken, GetValidatedUsers),
		"/users/messages/add":    handler(addMessageParams, AdminToken, AddMessage),
		"/users/messages/own":    handler(noParams, UserToken, GetOwnMessages),
		"/users/messages/solve":  handler(idParams, MessageOwnerOrAdminToken, SolveMessage),
		"/users/validate":        handler(idParams, AdminToken, ValidateUser),

		"/candidates/get":    handler(noParams, NoToken, GetCandidates),
		"/candidates/add":    handler(addCandidateParams, AdminToken, AddCandidate),
		"/candidates/delete": handler(idParams, AdminToken, DeleteCandidate),

		"/elections/get":     handler(noParams, NoToken, GetElections),
		"/elections/publish": handler(idParams, AdminToken, PublishElection),
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
	paramsFunc func(*http.Request) (par.Values, error),
	tokenFunc func(*sql.DB, *http.Request, par.Values) (*jwt.Token, *Claims, error),
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
