package main

import (
	"encoding/json"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type DBType interface {
	CreateTableQuery() string
}

type Claims struct {
	User
	jwt.StandardClaims
}

type config struct {
	allowedIDTypes []string
}

type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	UniqueID string `json:"unique_id"`
	Password string `json:"-"`
	Salt     string `json:"-"`
	Role     string `json:"role"`

	Files    []UserFile    `json:"files"`
	Messages []UserMessage `json:"messages"`
}

func (u User) CreateTableQuery() string {
	return `CREATE TABLE IF NOT EXISTS users (
		id integer NOT NULL PRIMARY KEY,
		name TEXT NOT NULL,
		unique_id text UNIQUE NOT NULL,
		password TEXT NOT NULL,
		salt TEXT NOT NULL,
		role TEXT NOT NULL
	);`
}

type UserFile struct {
	ID          int    `json:"id"`
	UserID      int    `json:"-"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (f UserFile) CreateTableQuery() string {
	return `CREATE TABLE IF NOT EXISTS files (
		id integer NOT NULL PRIMARY KEY,
		user_id integer NOT NULL REFERENCES users(id),
		name TEXT NOT NULL,
		description text NOT NULL
	);`
}

type UserMessage struct {
	ID      int    `json:"id"`
	UserID  int    `json:"-"`
	Content string `json:"content"`
	Solved  bool   `json:"solved"`
}

func (m UserMessage) CreateTableQuery() string {
	return `CREATE TABLE IF NOT EXISTS messages (
		id integer NOT NULL PRIMARY KEY,
		user_id integer NOT NULL REFERENCES users(id),
		content TEXT NOT NULL,
		solved BOOLEAN NOT NULL
	);`
}

type Election struct {
	ID     int       `json:"id"`
	Name   string    `json:"name"`
	Start  time.Time `json:"start"`
	End    time.Time `json:"end"`
	Public bool      `json:"public"`

	// TODO add a type, not a simple string
	CountType     string `json:"count_type"`
	MaxCandidates int    `json:"max_candidates"`
	MinCandidates int    `json:"min_candidates"`

	Candidates []Candidate `json:"candidates"`
}

func (e Election) CreateTableQuery() string {
	return `CREATE TABLE IF NOT EXISTS elections (
		id integer NOT NULL PRIMARY KEY,
		name TEXT NOT NULL,
		date_start TIMESTAMP WITH TIME ZONE NOT NULL,
		date_end TIMESTAMP WITH TIME ZONE NOT NULL,
		public BOOLEAN NOT NULL,
		count_type TEXT NOT NULL,
		max_candidates INTEGER NOT NULL CHECK (max_candidates > 0),
		min_candidates INTEGER NOT NULL,
		CHECK (max_candidates >= min_candidates)
	);`
}

type Candidate struct {
	ID           int    `json:"id"`
	ElectionID   int    `json:"election_id"`
	Name         string `json:"name"`
	Presentation string `json:"presentation"`
	Image        string `json:"image"`
}

func (c Candidate) CreateTableQuery() string {
	return `CREATE TABLE IF NOT EXISTS candidates (
		id integer NOT NULL PRIMARY KEY,
		election_id INTEGER NOT NULL REFERENCES elections(id),
		name TEXT NOT NULL,
		presentation TEXT NOT NULL,
		image TEXT NOT NULL
	);`
}

type Vote struct {
	ID         int    `json:"id"`
	Hash       string `json:"hash"`
	Candidates []int  `json:"candidates"`

	CandidatesString json.RawMessage `json:"-"`
}

func (v Vote) CreateTableQuery() string {
	return `CREATE TABLE IF NOT EXISTS votes (
		id integer NOT NULL PRIMARY KEY,
		hash TEXT UNIQUE NOT NULL,
		candidates json NOT NULL
	);`
}
