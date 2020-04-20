// TODO maybe some consts could be defined by environment variables
package main

const (
	// ID_ represent types of identification documents used for a user's unique ID
	ID_DNI      = "dni"      // spanish DNI
	ID_NIE      = "nie"      // spanish NIE
	ID_PASSPORT = "passport" // international passport

	// ROLE_ represent the possible roles that a used has in the system
	ROLE_NONE      = "none"      // the user can log in and see public information
	ROLE_VALIDATED = "validated" // the user can vote in all elections
	ROLE_ADMIN     = "admin"     // the user can see and edit everythin

	// COUNT_ represent the available count methods for elections
	COUNT_BORDA   = "borda"   // https://en.wikipedia.org/wiki/Borda_count
	COUNT_DOWDALL = "dowdall" // https://en.wikipedia.org/wiki/Borda_count

	MIN_PASSWORD_LENGTH = 6
)
