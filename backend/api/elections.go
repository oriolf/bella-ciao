package api

import (
	"database/sql"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/oriolf/bella-ciao/backend/models"
	"github.com/oriolf/bella-ciao/backend/queries"
	"github.com/oriolf/bella-ciao/backend/util"
	"github.com/pkg/errors"
)

func GetElections(w http.ResponseWriter, db *sql.DB, token *jwt.Token, claims *models.Claims, params interface{}) error {
	elections, err := queries.GetElections(db, !util.IsAdmin(claims)) // all non-admin get only public elections
	if err != nil {
		return errors.Wrap(err, "could not get elections")
	}

	if err := util.WriteResult(w, elections); err != nil {
		return errors.Wrap(err, "could not write response")
	}

	return nil
}
