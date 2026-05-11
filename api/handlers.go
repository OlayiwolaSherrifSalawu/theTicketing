package api

import (
	"encoding/json"
	"net/http"

	"github.com/OlayiwolaSherrifSalawu/theTicketing.git/pkg/model"
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
}
