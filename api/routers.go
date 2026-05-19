package api

import "net/http"

func (a *Application) routers() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /createUser", a.CreateUserHandler)
	mux.HandleFunc("POST /login", a.LoginUser)
	mux.Handle("GET /dashboard", a.RequireAuthentication(http.HandlerFunc(a.DownLoadHandler)))
	return a.RecoverPanic(a.XssProtection(a.LogMessage(mux)))
}
