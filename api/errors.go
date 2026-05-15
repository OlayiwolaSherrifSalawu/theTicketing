package api

import (
	"fmt"
	"net/http"
	"runtime/debug"
)

type ConstantError string

func (e ConstantError) Error() string {
	return string(e)
}

const (
	OPENING_POOL_FAILED = ConstantError("Failure in creating database pool")
	MIGRATION_FAILED    = ConstantError("Failed migrations")
)

func (a *Application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (a *Application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s \n %s", err.Error(), debug.Stack())
	a.ErroMessage.Output(2, trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}
