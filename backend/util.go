package util

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/oriolf/bella-ciao/backend/models"
	"github.com/pkg/errors"
	"golang.org/x/crypto/scrypt"
)

// TODO move to environment?
var JWTKey = []byte("my_secret_key")

func IsAdmin(claims *models.Claims) bool {
	return claims != nil && claims.User.Role == "admin"
}

func GetParams(r *http.Request, model interface{}) error {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(model); err != nil {
		return err
	}

	return nil
}

func SafeID() (string, error) {
	b := make([]byte, 32)
	n, err := rand.Read(b)
	if err != nil {
		return "", errors.Wrap(err, "can't read from crypto/rand")
	}
	if n != len(b) {
		return "", errors.New("wrong length read from crypto/rand")
	}

	return hex.EncodeToString(b), nil
}

func HashPassword(pass, salt string) (string, error) {
	bsalt, err := hex.DecodeString(salt)
	if err != nil {
		return "", errors.New("invalid salt")
	}
	bdk, err := scrypt.Key([]byte(pass), bsalt, 32768, 8, 1, 32)
	if err != nil {
		return "", errors.New("scrypt error")
	}
	return hex.EncodeToString(bdk), nil
}

func ValidatePassword(pass, dbpass, salt string) error {
	hashed, err := HashPassword(pass, salt)
	if err != nil {
		return errors.Wrap(err, "error hashing password")
	}

	if hashed != dbpass {
		return errors.New("invalid password")
	}

	return nil
}

func WriteResult(w http.ResponseWriter, result interface{}) error {
	js, err := json.Marshal(result)
	if err != nil {
		return errors.Wrap(err, "could not marshal result")
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(js); err != nil {
		return errors.Wrap(err, "could not write result")
	}

	return nil
}

func GenerateToken(user models.User) (string, error) {
	expirationTime := time.Now().Add(5 * time.Minute)
	claims := &models.Claims{
		User:           user,
		StandardClaims: jwt.StandardClaims{ExpiresAt: expirationTime.Unix()},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(JWTKey)
}
