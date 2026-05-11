package api

import "net/http"

type ConstantError string

func (e ConstantError) Error() string {
	return string(e)
}

const (
	OPENING_POOL_FAILED = ConstantError("Failure in creating database pool")
)

func (a *Application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)

}
