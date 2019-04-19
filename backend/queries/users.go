package queries

import (
	"database/sql"

	"github.com/oriolf/bella-ciao/backend/models"
)

func RegisterUser(db *sql.DB, user models.User) error {
	_, err := db.Exec("INSERT INTO users (name, unique_id, password, salt, role) VALUES ($1, $2, $3, $4, 'none');",
		user.Name, user.UniqueID, user.Password, user.Salt)
	return err
}

func GetUserFromUniqueID(db *sql.DB, uniqueID string) (user models.User, err error) {
	err = db.QueryRow("SELECT id, name, password, salt, role FROM users WHERE unique_id LIKE $1;", uniqueID).Scan(
		&user.ID, &user.Name, &user.Password, &user.Salt, &user.Role)
	user.UniqueID = uniqueID
	return user, err
}
