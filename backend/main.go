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
				String("unique_id", par.NonEmpty, par.UpperCase, par.StringValidates(ID_VALIDATION_FUNCS)).
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
				String("count_method", par.StringIn(COUNT_METHODS)).
				Int("min_candidates", par.PositiveInt).
				Int("max_candidates", par.PositiveInt).
				ValidateFunc(validateElectionParams)

	initializeParams = par.P("json").
				JSON("admin", registerParamsAux.EndJSON()).
				JSON("election", electionParamsAux.EndJSON()).End()

	appHandlers = map[string]func(http.ResponseWriter, *http.Request){
		"/uninitialized": handler(noParams, noToken, Uninitialized),
		"/initialize":    handler(initializeParams, noToken, Initialize),

		"/auth/register": handler(registerParams, noToken, Register),
		"/auth/login":    handler(loginParams, noToken, Login),

		"/users/files/own":      handler(noParams, requireToken, GetOwnFiles),
		"/users/files/delete":   handler(idParams, tokenFuncs(requireToken, fileOwnerOrAdminToken), DeleteFile),
		"/users/files/download": handler(idParams, tokenFuncs(requireToken, fileOwnerOrAdminToken), DownloadFile),
		"/users/files/upload":   handler(uploadFileParams, requireToken, UploadFile),

		"/users/unvalidated/get": handler(noParams, tokenFuncs(requireToken, adminToken), GetUnvalidatedUsers),
		"/users/validated/get":   handler(noParams, tokenFuncs(requireToken, adminToken), GetValidatedUsers),
		"/users/messages/add":    handler(addMessageParams, tokenFuncs(requireToken, adminToken), AddMessage),
		"/users/messages/own":    handler(noParams, requireToken, GetOwnMessages),
		"/users/messages/solve":  handler(idParams, tokenFuncs(requireToken, messageOwnerOrAdminToken), SolveMessage),
		"/users/validate":        handler(idParams, tokenFuncs(requireToken, adminToken), ValidateUser),

		"/candidates/get":    handler(noParams, noToken, GetCandidates),
		"/candidates/add":    handler(addCandidateParams, tokenFuncs(requireToken, adminToken), AddCandidate),
		"/candidates/delete": handler(idParams, tokenFuncs(requireToken, adminToken), DeleteCandidate),

		"/elections/get":     handler(noParams, noToken, GetElections),
		"/elections/publish": handler(idParams, tokenFuncs(requireToken, adminToken), PublishElection),
		// TODO store allowed identification types as part of the initialization
		// TODO implement and test /config/update (for global options), /elections/update
		// validate and test count method in COUNT_METHODS on initialize, /elections/update
		// validate and test unique ID on register
		// /config/update can only add allowed identification formats, not remove them
		// TODO implement and test /elections/vote, etc.
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
	tokenFunc func(*sql.DB, *Claims, par.Values, error) error,
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

		// TODO wrap everything in a transaction, update the rest of the code accordingly

		token, claims, err := getRequestToken(r)
		if err := tokenFunc(db, claims, params, err); err != nil { // token func validates permissions too
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

func tokenFuncs(fs ...func(*sql.DB, *Claims, par.Values, error) error) func(*sql.DB, *Claims, par.Values, error) error {
	return func(db *sql.DB, claims *Claims, values par.Values, err error) error {
		for _, f := range fs {
			err = f(db, claims, values, err)
			if err != nil {
				return err
			}
		}
		return nil
	}
}
