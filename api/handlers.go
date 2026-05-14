package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/OlayiwolaSherrifSalawu/theTicketing.git/pkg/model"
	"golang.org/x/crypto/bcrypt"
)

type dto struct {
	UserName     string `json:"userName"`
	EmailAddress string `json:"emailAddress"`
	Password     string `json:"password"`
}

func (a *Application) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	inputDto := dto{}
	r.Body = http.MaxBytesReader(w, r.Body, 1235)
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&inputDto)
	if err != nil {
		a.clientError(w, http.StatusBadRequest)
		return
	}
	HashPassword, err := hashPassword(inputDto.Password)
	if err != nil {
		a.clientError(w, http.StatusBadRequest)
		return
	}
	Users := &model.User{
		ID:           inputDto.ID,
		UserName:     inputDto.UserName,
		EmailAddress: inputDto.EmailAddress,
		TimeStamp:    time.Now(),
		HashPassword: HashPassword,
	}
	inputDto = dto{}
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
