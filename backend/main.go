package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"sync/atomic"

	_ "github.com/mattn/go-sqlite3"

	"github.com/gorilla/sessions"
	"github.com/oriolf/bella-ciao/params"
)

var (
	// TODO use this to manage secret https://diogomonica.com/2017/03/27/why-you-shouldnt-use-env-variables-for-secret-data/
	store         = sessions.NewFilesystemStore(SESSIONS_FOLDER, []byte("my_secret"))
	queryCount    uint64
	globalTesting bool

	idParams = par.P("query").Int("id", par.PositiveInt).End()
	noParams = par.None()

	registerParamsAux = par.P("json").
				String("name", par.NonEmpty).
				String("unique_id", par.NonEmpty, par.UpperCase, par.StringValidates(ID_VALIDATION_FUNCS)).
				Email("email").
				String("password", par.MinLength(MIN_PASSWORD_LENGTH))
	registerParams = registerParamsAux.End()

	electionParamsAux = par.P("json").
				String("name", par.NonEmpty).
				Time("start", par.NonZeroTime).
				Time("end", par.NonZeroTime).
				String("count_method", par.StringIn(COUNT_METHODS)).
				Int("min_candidates", par.PositiveInt).
				Int("max_candidates", par.PositiveInt).
				ValidateFunc(validateElectionParams)

	globalConfigParamsAux = par.P("json").
				StringList("id_formats", par.ListMinLength(1), par.StringsIn(ID_FORMATS))

	initializeParams = par.P("json").
				JSON("admin", registerParamsAux.EndJSON()).
				JSON("election", electionParamsAux.EndJSON()).
				JSON("config", globalConfigParamsAux.EndJSON()).End()

	loginParams = par.P("json").
			String("unique_id", par.NonEmpty).
			String("password", par.NonEmpty).End()

	uploadFileParams = par.P("form").
				File("file").
				String("description", par.NonEmpty).End()

	userListParams = par.P("query").
			Int("page", par.PositiveInt).
			Int("items_per_page", par.PositiveInt).
			String("query").End()

	addMessageParams = par.P("json").
				Int("user_id", par.PositiveInt).
				String("content", par.NonEmpty).End()

	addCandidateParams = par.P("form").
				File("image").
				String("name", par.NonEmpty).
				String("presentation", par.NonEmpty).End()

	voteParams = par.P("json").
			IntList("candidates").End()

	appHandlers = map[string]func(http.ResponseWriter, *http.Request){
		"/uninitialized": handler(noParams, noLogin, Uninitialized),
		"/initialize":    handler(initializeParams, noLogin, Initialize),
		"/config/update": handler(globalConfigParamsAux.End(), authFuncs(requireLogin, adminUser), UpdateConfig),

		"/auth/register": handler(registerParams, authFuncs(noLogin, validIDFormats), Register),
		"/auth/login":    handler(loginParams, noLogin, Login),
		"/auth/logout":   handler(noParams, noLogin, Logout), // TODO test

		"/users/whoami":         handler(noParams, requireLogin, GetSelf), // TODO test
		"/users/files/own":      handler(noParams, requireLogin, GetOwnFiles),
		"/users/files/delete":   handler(idParams, authFuncs(requireLogin, fileOwnerOrAdminUser), DeleteFile),
		"/users/files/download": handler(idParams, authFuncs(requireLogin, fileOwnerOrAdminUser), DownloadFile),
		"/users/files/upload":   handler(uploadFileParams, requireLogin, UploadFile),

		"/users/unvalidated/get": handler(userListParams, authFuncs(requireLogin, adminUser), GetUnvalidatedUsers),
		"/users/validated/get":   handler(userListParams, authFuncs(requireLogin, adminUser), GetValidatedUsers),
		"/users/messages/add":    handler(addMessageParams, authFuncs(requireLogin, adminUser), AddMessage),
		"/users/messages/own":    handler(noParams, requireLogin, GetOwnMessages),
		"/users/messages/solve":  handler(idParams, authFuncs(requireLogin, messageOwnerOrAdminUser), SolveMessage),
		"/users/validate":        handler(idParams, authFuncs(requireLogin, adminUser), ValidateUser),

		"/candidates/get":    handler(noParams, noLogin, GetCandidates),
		"/candidates/image":  handler(idParams, noLogin, GetCandidateImage),
		"/candidates/add":    handler(addCandidateParams, authFuncs(requireLogin, adminUser, electionDidNotStart), AddCandidate),
		"/candidates/delete": handler(idParams, authFuncs(requireLogin, adminUser, electionDidNotStart), DeleteCandidate),

		"/elections/get":     handler(noParams, noLogin, GetElections),
		"/elections/publish": handler(idParams, authFuncs(requireLogin, adminUser), PublishElection),
		"/elections/vote":    handler(voteParams, authFuncs(requireLogin, validatedUser), CastVote),
		// TODO providing the hash you should see which candidates were voted and in which order
		//"/elections/vote/check": handler(checkVoteParams, noLogin, CheckVote),
		// TODO every minute it should be checked if an election has to be counted its results
		// TODO if the election has its results counted, then return those results
		//"/elections/result": handler(idParams, authFuncs(requireLogin, validatedUser), ElectionResults),
		// TODO implement /elections/update, test only valid params are accepted
	}

	initialized struct {
		value bool
		mutex sync.Mutex
	}
)

func main() {
	if err := bootstrap(); err != nil {
		log.Fatalln("Could not bootstrap:", err)
	}

	log.Println("Completed bootstrap, start listening...")
	log.Fatalln(http.ListenAndServe(":9876", nil))
}

func bootstrap() error {
	db, err := sql.Open("sqlite3", DB_FILE)
	if err != nil {
		return fmt.Errorf("error during database connection: %w", err)
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}

	if err := InitDB(tx); err != nil {
		return fmt.Errorf("error during database initialization: %w", err)
	}

	count, err := countElections(tx)
	if err != nil {
		return fmt.Errorf("could not count elections in check initialized: %w", err)
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return fmt.Errorf("could not commit transaction: %w", err)
	}

	if count > 0 {
		initialized.value = true
	}

	for _, folder := range []string{UPLOADS_FOLDER, SESSIONS_FOLDER} {
		if _, err := os.Stat(folder); err != nil {
			if err := os.Mkdir(folder, 0755); err != nil {
				return fmt.Errorf("could not create %s folder: %w", folder, err)
			}
		}
	}

	for path, handler := range appHandlers {
		http.HandleFunc(path, handler)
	}

	http.Handle("/", http.FileServer(http.Dir("website")))

	return nil
}

func checkInitialized(db *sql.Tx) {
}

func getInitialized() bool {
	initialized.mutex.Lock()
	defer initialized.mutex.Unlock()
	return initialized.value
}

func handler(
	paramsFunc func(*http.Request) (par.Values, error),
	authFunc func(*sql.Tx, *User, par.Values, error) error,
	handleFunc func(*http.Request, http.ResponseWriter, *sql.Tx, *User, par.Values) error,
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddUint64(&queryCount, 1)
		if initialized := getInitialized(); !initialized {
			if r.URL.Path != "/initialize" && r.URL.Path != "/uninitialized" {
				log.Printf("[%d] Invalid method %q before initialization\n", n, r.URL.Path)
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

		log.Printf("[%d] Received petition to %s\n", n, r.URL.Path)
		params, err := paramsFunc(r)
		if err != nil {
			log.Printf("[%d] Error validating parameters: %s\n", n, err)
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		db, err := sql.Open("sqlite3", DB_FILE)
		if err != nil {
			log.Fatalf("[%d] Error during database connection in handler: %s\n", n, err)
		}
		defer db.Close()

		tx, err := db.Begin()
		if err != nil {
			log.Printf("[%d] Could not begin transaction: %s\n", n, err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		user, err := getRequestUser(r, w, tx)
		if err := authFunc(tx, user, params, err); err != nil { // auth func validates permissions too
			log.Printf("[%d] Error during authorization: %s\n", n, err)
			rollback(n, tx)
			http.Error(w, "", http.StatusUnauthorized)
			return
		}

		if err := handleFunc(r, w, tx, user, params); err != nil {
			log.Printf("[%d] Error handling request: %s\n", n, err)
			rollback(n, tx)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		if err := tx.Commit(); err != nil {
			log.Printf("[%d] Error commiting transaction: %s.", n, err)
			rollback(n, tx)
			http.Error(w, "", http.StatusInternalServerError)
		}
	}
}

func rollback(n uint64, tx *sql.Tx) {
	if err := tx.Rollback(); err != nil {
		log.Printf("[%d] Error during transaction rollback: %s\n", n, err)
	}
}

func authFuncs(fs ...func(*sql.Tx, *User, par.Values, error) error) func(*sql.Tx, *User, par.Values, error) error {
	return func(db *sql.Tx, user *User, values par.Values, err error) error {
		for _, f := range fs {
			err = f(db, user, values, err)
			if err != nil {
				return err
			}
		}
		return nil
	}
}
