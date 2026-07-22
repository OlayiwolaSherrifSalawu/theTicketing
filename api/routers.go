package api

import "net/http"

func (a *Application) routers() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /createUser", a.CreateUserHandler)
	mux.HandleFunc("POST /login", a.LoginUser)
	mux.Handle("GET /dashboard", a.RequireAuthentication(http.HandlerFunc(a.DownLoadHandler)))
	mux.HandleFunc("POST /events", a.CreateEvent)
	mux.HandleFunc("GET /events", a.GetEvent)

	// Web (server-rendered) routes. Named distinctly from the JSON API
	// routes above (/signup vs /createUser) so the two can evolve
	// independently without colliding.
	mux.HandleFunc("GET /signup", a.SignupPageHandler)
	mux.HandleFunc("POST /signup", a.SignupFormHandler)

	// POST /login is already taken by the JSON API above, so the htmx
	// target for the login form gets its own path.
	mux.HandleFunc("GET /login", a.LoginPageHandler)
	mux.HandleFunc("POST /login-form", a.LoginFormHandler)

	// "{$}" matches exactly "/" (not a subtree) so it doesn't swallow
	// every unmatched path as the homepage.
	mux.HandleFunc("GET /{$}", a.HomePageHandler)
	mux.HandleFunc("GET /events-partial", a.EventsPartialHandler)

	// Static assets: web/static/css/output.css (Tailwind build) and
	// web/static/js/*.js (vendored htmx + json-enc extension — no CDN
	// dependency, works offline).
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	return a.RecoverPanic(a.XssProtection(a.LogMessage(mux)))
}
