package api

import (
	"net/http"
	"strings"
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
			auth := r.Header.Get("Authorization")
			if auth == "" {
				a.clientError(w, http.StatusUnauthorized)
				return
			}
			theAuth := strings.TrimPrefix(auth, "Bearer ")
		})
}
