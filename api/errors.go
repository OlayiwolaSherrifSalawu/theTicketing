package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
)

type apiError struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"error"`
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
	CANT_SEND_NIL_CTX         = ConstantError("Cant serve nil context would cause panic")
)

func (a *Application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (a *Application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s \n %s", err.Error(), debug.Stack())
	a.ErroMessage.Output(2, trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}
