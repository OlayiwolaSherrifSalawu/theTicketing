package api

import (
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
	a.InfoLogger.Println("afang1")
	r.Body = http.MaxBytesReader(w, r.Body, 1235)
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&inputDto)

	if err != nil {
		a.clientError(w, http.StatusBadRequest)
		return
	}
	a.InfoLogger.Println("afang2")

	HashPassword, err := hashPassword(inputDto.Password)
	if err != nil {
		a.clientError(w, http.StatusBadRequest)
		return
	}
	a.InfoLogger.Println("afang3")

	Id := uuid.New().String()
	Users := &model.User{
		ID:           Id,
		UserName:     inputDto.UserName,
		EmailAddress: inputDto.EmailAddress,
		TimeStamp:    time.Now(),
		HashPassword: HashPassword,
	}
	a.InfoLogger.Println("afang4")

	inputDto = dto{}
	a.InfoLogger.Println("afang4.1")

	err = a.TheTicket.Insert(Users) //this is for some reason stoping my fuction
	a.InfoLogger.Println("afang5")  //this is not loging
	if err != nil {
		a.clientError(w, 505)
		return
	}
	fmt.Fprint(w, "successfully register users\n")

}
