package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/OlayiwolaSherrifSalawu/theTicketing.git/pkg/model/theticket"
)

type apiError struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"error"`
}

func (e *apiError) Error() string { return e.Message }

func fieldError(field, message string) *apiError {
	return &apiError{Field: field, Message: message}
}

func (a *Application) clientErrorJSON(w http.ResponseWriter, status int, field, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(apiError{Field: field, Message: message})
}

type ConstantError string

func (e ConstantError) Error() string {
	return string(e)
}

const (
	OPENING_POOL_FAILED       = ConstantError("Failure in creating database pool")
	MIGRATION_FAILED          = ConstantError("Failed migrations")
	UNEXPECTED_SIGNING_METHOD = ConstantError("Unexpected Signing Method")
	FAILED_TO_GETID           = ConstantError("Failed to get id from claims")
)

func (a *Application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

// serverError logs the cause with a stack trace and sends a generic 500.
// It deliberately stays plain text: SignupFormHandler shows JSON error
// messages to the user, and internals must never reach them.
func (a *Application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s \n %s", err.Error(), debug.Stack())
	a.ErroMessage.Output(2, trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// storeError maps the store's sentinel errors onto JSON client errors and
// treats anything else as a server error.
func (a *Application) storeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, theticket.ErrNotFound):
		a.clientErrorJSON(w, http.StatusNotFound, "", "Not found.")
	case errors.Is(err, theticket.ErrForbidden):
		a.clientErrorJSON(w, http.StatusForbidden, "", "You don't have permission to change this.")
	case errors.Is(err, theticket.ErrDuplicateEmail):
		a.clientErrorJSON(w, http.StatusConflict, "emailAddress", "An account with this email already exists.")
	case errors.Is(err, theticket.ErrDuplicateUsername):
		a.clientErrorJSON(w, http.StatusConflict, "userName", "This username is taken.")
	case errors.Is(err, theticket.ErrSoldOut):
		a.clientErrorJSON(w, http.StatusConflict, "quantity", "Not enough tickets left.")
	case errors.Is(err, theticket.ErrEventStarted):
		a.clientErrorJSON(w, http.StatusUnprocessableEntity, "", "This event has already started.")
	case errors.Is(err, theticket.ErrCapacityTooLow):
		a.clientErrorJSON(w, http.StatusConflict, "totalCapacity", "Capacity can't be lower than the tickets already sold.")
	case errors.Is(err, theticket.ErrAlreadyCancelled):
		a.clientErrorJSON(w, http.StatusConflict, "", "This ticket is already cancelled.")
	default:
		a.serverError(w, err)
	}
}
