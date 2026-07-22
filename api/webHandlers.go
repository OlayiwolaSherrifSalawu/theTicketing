package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

// responseRecorder is a minimal http.ResponseWriter that buffers what's
// written to it instead of sending it over the wire. It's deliberately not
// httptest.ResponseRecorder: that type lives in a *_test-oriented package
// and pulling it into a production request path is a smell worth avoiding —
// this ~10-line version does exactly what we need with no test-only baggage.
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

// SignupPageHandler renders the signup page on a normal GET.
func (a *Application) SignupPageHandler(w http.ResponseWriter, r *http.Request) {
	if err := a.Templates.Render(w, http.StatusOK, "signup.tmpl", map[string]any{}); err != nil {
		a.serverError(w, err)
	}
}

// SignupFormHandler is the htmx target for the signup form. It doesn't
// duplicate CreateUserHandler's decode/validate/insert logic — it captures
// whatever CreateUserHandler would have written to a real client, and
// decodes that back into a struct to decide what HTML to render. This
// keeps CreateUserHandler as the single source of truth for "what counts
// as a valid signup," for both the JSON API and this HTML form.
func (a *Application) SignupFormHandler(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		a.serverError(w, err)
		return
	}

	// Best-effort decode, only used to repopulate the form on error —
	// CreateUserHandler does its own (authoritative) decode below.
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
			// CreateUserHandler fell back to a's plain-text clientError/
			// serverError helpers (not our JSON one) — e.g. a malformed
			// body or a DB failure. Don't leak that raw text into the UI;
			// show a safe generic message instead.
			data["Error"] = "Something went wrong creating your account. Please try again."
		}

		if err := a.Templates.RenderPartial(w, rec.status, "signup.tmpl", "signup_form", data); err != nil {
			a.serverError(w, err)
		}
		return
	}

	// Success: send the browser to the login page. HX-Redirect tells htmx
	// to do a full navigation rather than swap this response's body in.
	w.Header().Set("HX-Redirect", "/login")
	w.WriteHeader(http.StatusOK)
}
