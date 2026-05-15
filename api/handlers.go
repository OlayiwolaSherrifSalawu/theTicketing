package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/OlayiwolaSherrifSalawu/theTicketing.git/pkg/model"
	"github.com/google/uuid"
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

	Id := uuid.New().String()
	Users := &model.User{
		ID:           Id,
		UserName:     inputDto.UserName,
		EmailAddress: inputDto.EmailAddress,
		TimeStamp:    time.Now(),
		HashPassword: HashPassword,
	}
	inputDto = dto{}
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	err = a.TheTicket.Insert(ctx, Users)
	if err != nil {
		a.serverError(w, err)
		return
	}
	fmt.Fprint(w, "successfully register users\n")

}
