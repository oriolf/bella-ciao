package models

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

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
}

type Election struct {
	ID    int       `json:"id"`
	Name  string    `json:"name"`
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`

	CountType     string `json:"count_type"`
	MaxCandidates int    `json:"max_candidates"`
	MinCandidates int    `json:"min_candidates"`
}

type Candidate struct {
	ID           int    `json:"id"`
	ElectionID   int    `json:"election_id"`
	Name         string `json:"name"`
	Presentation string `json:"presentation"`
	Image        string `json:"image"`
}

type Vote struct {
	ID               int    `json:"id"`
	Hash             string `json:"hash"`
	Candidates       []int  `json:"candidates"`
	CandidatesString string `json:"-"`
}
