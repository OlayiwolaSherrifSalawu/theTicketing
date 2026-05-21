package api

import (
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

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
			next.ServeHTTP(w, r)
		})
}

func (a *Application) RecoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					w.Header().Set("Connection", "closed")
				}
			}()
			next.ServeHTTP(w, r)
		})
}

func (a *Application) RequireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			auth := a.getToken(r)
			if auth == "" {
				a.clientError(w, http.StatusUnauthorized)
				return
			}
			theAuth := strings.TrimPrefix(auth, "Bearer ")
			tokens, err := jwt.Parse(theAuth, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, UNEXPECTED_SIGNING_METHOD
				}
				return []byte(os.Getenv("JWT")), nil
			})
			if err != nil || !tokens.Valid {
				a.clientError(w, http.StatusUnauthorized)
				return
			}
			claims := tokens.Claims.(jwt.MapClaims)
			next.ServeHTTP(w, r)
		})
}
