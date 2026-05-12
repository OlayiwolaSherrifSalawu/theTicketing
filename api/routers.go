package api

import "net/http"

func (a *Application) routers() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /createUser", a.CreateUserHandler)
	return a.RecoverPanic(a.XssProtection(a.LogMessage(mux)))
}
