package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"

	"github.com/dgrijalva/jwt-go"
)

const dbfile = "db.db"

func main() {
	initDB()

	http.HandleFunc("/initialized", handler(NoToken, noParams, Initialized))
	http.HandleFunc("/initialize", handler(NoToken, GetInitializeParams, Initialize))

	http.HandleFunc("/auth/register", handler(NoToken, GetRegisterParams, Register))
	http.HandleFunc("/auth/login", handler(NoToken, GetLoginParams, Login))
	http.HandleFunc("/auth/refresh", handler(UserToken, noParams, Refresh))
	// http.HandleFunc("/auth/logout", handler(userToken, noParams, login)) // TODO optional blacklist token

	http.HandleFunc("/elections/get", handler(NoToken, noParams, GetElectionsHandler))
	// http.HandleFunc("/elections/create", handler(adminToken, electionParams, createElection))
	// http.HandleFunc("/elections/update", handler(adminToken, electionParams, updateElection))
	// http.HandleFunc("/elections/delete", handler(adminToken, electionParams, deleteElection))
	// http.HandleFunc("/elections/vote", handler(validatedToken, voteParams, vote))

	http.HandleFunc("/candidates/get", handler(NoToken, noParams, GetCandidatesHandler))
	http.HandleFunc("/candidates/add", handler(AdminToken, GetCandidateParams, AddCandidateHandler))

	http.HandleFunc("/users/unvalidated/get", handler(AdminToken, noParams, GetUnvalidatedUsersHandler))
	http.HandleFunc("/users/files/download", handler(OwnerOrAdminToken, IDParams, DownloadFile))
	// TODO upload and download files

	log.Println("Start listening...")
	log.Fatalln(http.ListenAndServe(":9876", nil))
}

func initDB() {
	db, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		log.Fatalln("Error during database connection:", err)
	}
	defer db.Close()

	if err := InitDB(db); err != nil {
		log.Fatalln("Error during database initialization:", err)
	}
}

func handler(
	tokenFunc func(*http.Request) (*jwt.Token, *Claims, error),
	paramsFunc func(*http.Request, *jwt.Token) (interface{}, error),
	handleFunc func(http.ResponseWriter, *sql.DB, *jwt.Token, *Claims, interface{}) error,
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}

		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST")
		w.Header().Set("Access-Control-Allow-Headers", "X-Requested-With,content-type,Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			return
		}

		log.Println("Received petition to", r.URL.Path)
		token, claims, err := tokenFunc(r)
		if err != nil {
			log.Println("Error validating token:", err)
			http.Error(w, "", http.StatusUnauthorized)
			return
		}

		params, err := paramsFunc(r, token)
		if err != nil {
			log.Println("Error validating parameters:", err)
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		db, err := sql.Open("sqlite3", dbfile)
		if err != nil {
			log.Fatalln("Error during database connection in handler:", err)
		}
		defer db.Close()

		if err := handleFunc(w, db, token, claims, params); err != nil {
			log.Println("Error handling request:", err)
			http.Error(w, "", http.StatusInternalServerError)
		}
	}
}

func noParams(*http.Request, *jwt.Token) (interface{}, error) {
	return nil, nil
}
