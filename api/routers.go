package api

import "net/http"

func (a *Application) routers() http.Handler {
	mux := http.NewServeMux()
	auth := func(h http.HandlerFunc) http.Handler { return a.RequireAuthentication(h) }

	// JSON API — accounts.
	mux.HandleFunc("POST /createUser", a.CreateUserHandler)
	mux.HandleFunc("POST /login", a.LoginUser)
	mux.HandleFunc("POST /logout", a.LogoutHandler)
	mux.Handle("GET /me", auth(a.MeHandler))

	// JSON API — events. Browsing is public; writing needs a login and,
	// for update/delete, ownership.
	mux.HandleFunc("GET /events", a.GetEvent)
	mux.HandleFunc("GET /events/{id}", a.GetEventByID)
	mux.Handle("POST /events", auth(a.CreateEvent))
	mux.Handle("PUT /events/{id}", auth(a.UpdateEvent))
	mux.Handle("DELETE /events/{id}", auth(a.DeleteEvent))
	mux.Handle("GET /me/events", auth(a.MyEventsHandler))

	// JSON API — tickets.
	mux.Handle("POST /events/{id}/tickets", auth(a.BookTicketsHandler))
	mux.Handle("GET /me/tickets", auth(a.MyTicketsHandler))
	mux.Handle("DELETE /tickets/{id}", auth(a.CancelTicketHandler))

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
	mux.Handle("GET /dashboard", auth(a.HomePageHandler))
	mux.HandleFunc("GET /events-partial", a.EventsPartialHandler)

	// Static assets: web/static/css/output.css (Tailwind build) and
	// web/static/js/*.js (vendored htmx + json-enc extension — no CDN
	// dependency, works offline).
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	return a.RecoverPanic(a.XssProtection(a.LogMessage(mux)))
}
