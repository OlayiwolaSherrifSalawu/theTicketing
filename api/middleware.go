package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

const ct contexK = contexK("userID")

func (a *Application) LogMessage(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.InfoLogger.Printf("%s-%s, %s, %s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())
		next.ServeHTTP(w, r)
	})
}

func (a *Application) XssProtection(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("X-Frame-Options", "deny")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			next.ServeHTTP(w, r)
		})
}

// RecoverPanic turns a panic in any handler into a logged 500 instead of
// a dropped connection.
func (a *Application) RecoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					if err == http.ErrAbortHandler {
						panic(err)
					}
					w.Header().Set("Connection", "close")
					a.serverError(w, fmt.Errorf("panic: %v", err))
				}
			}()
			next.ServeHTTP(w, r)
		})
}

func (a *Application) RequireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			userID, ok := a.authenticate(r)
			if !ok {
				a.unauthorized(w, r)
				return
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ct, userID)))
		})
}

// authenticate validates the request's JWT and returns the user ID in it.
func (a *Application) authenticate(r *http.Request) (string, bool) {
	raw := a.getToken(r)
	if raw == "" {
		return "", false
	}
	token, err := jwt.Parse(raw, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, UNEXPECTED_SIGNING_METHOD
		}
		return a.JWTSecret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithExpirationRequired())
	if err != nil || !token.Valid {
		return "", false
	}
	userID, err := extractUserFromtoken(token)
	if err != nil {
		return "", false
	}
	return userID, true
}

// unauthorized sends browsers navigating to a page to the login screen,
// and gives everything else (API clients, htmx requests) a JSON 401.
func (a *Application) unauthorized(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet && r.Header.Get("HX-Request") == "" &&
		strings.Contains(r.Header.Get("Accept"), "text/html") {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	a.clientErrorJSON(w, http.StatusUnauthorized, "", "Authentication required.")
}
