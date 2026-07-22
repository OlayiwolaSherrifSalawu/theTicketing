package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)


type responseRecorder struct {
	status int
	body   bytes.Buffer
	header http.Header
}

func newResponseRecorder() *responseRecorder {
	return &responseRecorder{status: http.StatusOK, header: make(http.Header)}
}

func (r *responseRecorder) Header() http.Header         { return r.header }
func (r *responseRecorder) Write(b []byte) (int, error) { return r.body.Write(b) }
func (r *responseRecorder) WriteHeader(status int)      { r.status = status }

// LoginPageHandler renders the login page on a normal GET.
func (a *Application) LoginPageHandler(w http.ResponseWriter, r *http.Request) {
	if err := a.Templates.Render(w, http.StatusOK, "login.tmpl", map[string]any{}); err != nil {
		a.serverError(w, err)
	}
}

func (a *Application) LoginFormHandler(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		a.serverError(w, err)
		return
	}

	var submitted dto
	_ = json.Unmarshal(bodyBytes, &submitted)

	r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	rec := newResponseRecorder()
	a.LoginUser(rec, r)

	if rec.status >= http.StatusBadRequest {
		data := map[string]any{
			"Email": submitted.EmailAddress,
			"Error": "Incorrect email or password.",
		}
		if err := a.Templates.RenderPartial(w, rec.status, "login.tmpl", "login_form", data); err != nil {
			a.serverError(w, err)
		}
		return
	}

	
	for _, cookie := range rec.header.Values("Set-Cookie") {
		w.Header().Add("Set-Cookie", cookie)
	}

	w.Header().Set("HX-Redirect", "/dashboard")
	w.WriteHeader(http.StatusOK)
}
func (a *Application) SignupPageHandler(w http.ResponseWriter, r *http.Request) {
	if err := a.Templates.Render(w, http.StatusOK, "signup.tmpl", map[string]any{}); err != nil {
		a.serverError(w, err)
	}
}

func (a *Application) SignupFormHandler(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		a.serverError(w, err)
		return
	}

	
	var submitted dto
	_ = json.Unmarshal(bodyBytes, &submitted)

	r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	rec := newResponseRecorder()
	a.CreateUserHandler(rec, r)

	if rec.status >= http.StatusBadRequest {
		data := map[string]any{
			"Username": submitted.UserName,
			"Email":    submitted.EmailAddress,
		}

		var apiErr apiError
		if err := json.Unmarshal(rec.body.Bytes(), &apiErr); err == nil && apiErr.Message != "" {
			if apiErr.Field == "password" {
				data["PasswordError"] = apiErr.Message
			} else {
				data["Error"] = apiErr.Message
			}
		} else {
			data["Error"] = "Something went wrong creating your account. Please try again."
		}

		if err := a.Templates.RenderPartial(w, rec.status, "signup.tmpl", "signup_form", data); err != nil {
			a.serverError(w, err)
		}
		return
	}
	w.Header().Set("HX-Redirect", "/login")
	w.WriteHeader(http.StatusOK)
}
