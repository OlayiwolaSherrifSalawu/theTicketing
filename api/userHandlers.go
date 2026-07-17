package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/OlayiwolaSherrifSalawu/theTicketing.git/pkg/model"
	"github.com/google/uuid"
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

func (a *Application) LoginUser(w http.ResponseWriter, r *http.Request) {
	dtoInput := dto{}
	fetechedUser := &model.User{}
	r.Body = http.MaxBytesReader(w, r.Body, 1234)
	err := json.NewDecoder(r.Body).Decode(&dtoInput)

	if err != nil {
		a.clientError(w, http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	err = a.TheTicket.GetByEmail(ctx, dtoInput.EmailAddress, fetechedUser)
	a.InfoLogger.Println("my love is jesu")
	if err != nil {
		a.clientError(w, http.StatusUnauthorized)
		return
	}
	a.InfoLogger.Println("my love is jesu2")

	err = bcrypt.CompareHashAndPassword([]byte(fetechedUser.HashPassword), []byte(dtoInput.Password))
	if err != nil {
		a.clientError(w, http.StatusUnauthorized)
		return

	}
	a.InfoLogger.Println("my love is jesu3")

	dtoInput = dto{}
	tokenss, err := GenerateToken(fetechedUser.ID)
	if err != nil {
		a.serverError(w, err)
		return
	}

	response := map[string]string{
		"token": tokenss,
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		a.serverError(w, err)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "access",
		Value:    tokenss,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	// http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (a *Application) DownLoadHandler(w http.ResponseWriter, r *http.Request) {
	if userId, ok := extractUserFromContext(r.Context()); ok {
		fmt.Fprintf(w, "welcome user %s", userId)
		return
	}
	fmt.Fprint(w, "error getting your token")
	a.clientError(w, http.StatusUnauthorized)
}
