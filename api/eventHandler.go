package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/OlayiwolaSherrifSalawu/theTicketing.git/pkg/model"
)

func (a *Application) CreateEvent(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1234)
	newEvent := new(model.Event)
	err := json.NewDecoder(r.Body).Decode(newEvent)
	if err != nil {
		a.clientError(w, http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*3)
	defer cancel()
	err = a.TheTicket.InsertT(ctx, newEvent)
	if err != nil {
		a.serverError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	err = json.NewEncoder(w).Encode(newEvent)
	if err != nil {
		a.serverError(w, err)
		return
	}
}
