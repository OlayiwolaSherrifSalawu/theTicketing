package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/OlayiwolaSherrifSalawu/theTicketing.git/pkg/model"
	"golang.org/x/crypto/bcrypt"
)

func (a *Application) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	Users := new(model.User)
	r.Body = http.MaxBytesReader(w, r.Body, 1235)
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(Users)
	if err != nil {
		a.clientError(w, http.StatusBadRequest)
		return
	}
	Users.HashPassword, err = hashPassword(Users.HashPassword)
	if err != nil {
		a.clientError(w, http.StatusBadRequest)
		return
	}
	fmt.Fprint(w, "successfully register users\n")
}

func hashPassword(s string) (string, error) {
	hashByte, err := bcrypt.GenerateFromPassword([]byte(s), 10)
	if err != nil {
		return "", err
	}
	hashedPassword := string(hashByte)
	return hashedPassword, nil
}
