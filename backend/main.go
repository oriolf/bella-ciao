package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"

	"github.com/dgrijalva/jwt-go"
	"github.com/oriolf/bella-ciao/backend/api"
	"github.com/oriolf/bella-ciao/backend/models"
	"github.com/oriolf/bella-ciao/backend/queries"
)

var db *sql.DB

func main() {
	db = initDB()
	defer db.Close()

	http.HandleFunc("/auth/register", handler(api.NoToken, api.GetRegisterParams, api.Register))
	http.HandleFunc("/auth/login", handler(api.NoToken, api.GetLoginParams, api.Login))
	http.HandleFunc("/auth/refresh", handler(api.UserToken, noParams, api.Refresh))
	// http.HandleFunc("/auth/logout", handler(userToken, noParams, login)) // TODO optional blacklist token

	http.HandleFunc("/elections/get", handler(api.NoToken, noParams, api.GetElections))
	// http.HandleFunc("/elections/create", handler(adminToken, electionParams, createElection))
	// http.HandleFunc("/elections/update", handler(adminToken, electionParams, updateElection))
	// http.HandleFunc("/elections/delete", handler(adminToken, electionParams, deleteElection))
	// http.HandleFunc("/elections/vote", handler(validatedToken, voteParams, vote))

	log.Fatalln(http.ListenAndServe(":9876", nil))
}

func initDB() *sql.DB {
	dbUser, dbPassword := os.Getenv("BELLACIAO_USER"), os.Getenv("BELLACIAO_PASS")
	connStr := fmt.Sprintf("host=localhost port=5432 user=%s password=%s dbname=bella_ciao sslmode=disable", dbUser, dbPassword)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalln("Error during database connection:", err)
	}

	if err := queries.InitDB(db); err != nil {
		log.Fatalln("Error during database initialization:", err)
	}

	return db
}

func handler(
	tokenFunc func(*http.Request) (*jwt.Token, *models.Claims, error),
	paramsFunc func(*http.Request, *jwt.Token) (interface{}, error),
	handleFunc func(http.ResponseWriter, *sql.DB, *jwt.Token, *models.Claims, interface{}) error,
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}

		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, PATCH, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "X-Requested-With,content-type,Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

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

		if err := handleFunc(w, db, token, claims, params); err != nil {
			log.Println("Error handling request:", err)
			http.Error(w, "", http.StatusInternalServerError)
		}
	}
}

func noParams(*http.Request, *jwt.Token) (interface{}, error) {
	return nil, nil
}
