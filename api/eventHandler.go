package api

import (
	"encoding/json"
	"net/http"

	"github.com/OlayiwolaSherrifSalawu/theTicketing.git/pkg/model"
)

func (a *Application) CreateEvent(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1234)
	newEvent := new(model.Event)
	if err := json.NewDecoder(r.Body).Decode(newEvent); err != nil {
		a.serverError(w, err)
		return
	}
	
}
