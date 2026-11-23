package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/OlayiwolaSherrifSalawu/theTicketing.git/pkg/model"
	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret-that-is-at-least-32-bytes-long"

// newTestApp builds an Application with no database. Only use it for
// paths that are rejected before any query runs.
func newTestApp(t *testing.T) *Application {
	t.Helper()
	app, err := NewApplication(nil, testSecret)
	if err != nil {
		t.Fatal(err)
	}
	return app
}

func signed(t *testing.T, method jwt.SigningMethod, key any, claims jwt.MapClaims) string {
	t.Helper()
	s, err := jwt.NewWithClaims(method, claims).SignedString(key)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestNewApplicationRejectsWeakSecret(t *testing.T) {
	for _, secret := range []string{"", "short"} {
		if _, err := NewApplication(nil, secret); err == nil {
			t.Errorf("NewApplication(%q) succeeded; want error", secret)
		}
	}
}

func TestFetcher(t *testing.T) {
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()
	mux.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{
			"artists":   srv.URL + "/artists",
			"locations": srv.URL + "/locations",
			"dates":     srv.URL + "/dates",
			"relation":  srv.URL + "/relation",
		})
	})
	mux.HandleFunc("/artists", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[{"id":1,"name":"Queen","members":["Freddie"]}]`))
	})
	mux.HandleFunc("/locations", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"index":[{"id":1,"locations":["london-uk"]}]}`))
	})
	mux.HandleFunc("/dates", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"index":[{"id":1,"dates":["*01-01-2030"]}]}`))
	})
	mux.HandleFunc("/relation", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"index":[{"id":1,"datesLocations":{"london-uk":["01-01-2030"]}}]}`))
	})
	mux.HandleFunc("/broken", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{
			"artists": srv.URL + "/artists", "locations": srv.URL + "/missing",
			"dates": srv.URL + "/dates", "relation": srv.URL + "/relation",
		})
	})
	mux.HandleFunc("/partial", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"artists": srv.URL + "/artists"})
	})

	data, err := Fetcher(srv.URL + "/api")
	if err != nil {
		t.Fatal(err)
	}
	if len(data.Artists) != 1 || data.Artists[0].Name != "Queen" || data.Artists[0].ExternalId != 1 {
		t.Errorf("artists = %+v", data.Artists)
	}
	if len(data.Locations) != 1 || len(data.Dates) != 1 || len(data.Relations) != 1 {
		t.Errorf("got %d locations, %d dates, %d relations; want 1 each", len(data.Locations), len(data.Dates), len(data.Relations))
	}
	if data.Relations[0].JsonData["london-uk"][0] != "01-01-2030" {
		t.Errorf("relations = %+v", data.Relations)
	}

	// A failing endpoint or a missing key must return an error, not hang.
	for _, path := range []string{"/broken", "/partial", "/nope"} {
		done := make(chan error, 1)
		go func() { _, err := Fetcher(srv.URL + path); done <- err }()
		select {
		case err := <-done:
			if err == nil {
				t.Errorf("Fetcher(%s) succeeded; want error", path)
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("Fetcher(%s) hung", path)
		}
	}
}

func TestValidateSignup(t *testing.T) {
	tests := []struct {
		in    dto
		field string
	}{
		{dto{"ola", "Ola@Example.com ", "password1"}, ""},
		{dto{"ol", "ola@example.com", "password1"}, "userName"},
		{dto{"ola", "not-an-email", "password1"}, "emailAddress"},
		{dto{"ola", "Ola <ola@example.com>", "password1"}, "emailAddress"},
		{dto{"ola", "ola@example.com", "short"}, "password"},
		{dto{"ola", "ola@example.com", strings.Repeat("x", 73)}, "password"},
	}
	for _, tt := range tests {
		in := tt.in
		err := validateSignup(&in)
		got := ""
		if err != nil {
			got = err.Field
		}
		if got != tt.field {
			t.Errorf("validateSignup(%+v) field = %q, want %q", tt.in, got, tt.field)
		}
	}
	d := dto{"ola", "  Ola@Example.com ", "password1"}
	validateSignup(&d)
	if d.EmailAddress != "ola@example.com" {
		t.Errorf("email not normalised: %q", d.EmailAddress)
	}
}

func TestEventInputToEvent(t *testing.T) {
	now := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)
	valid := func() eventInput {
		return eventInput{EventName: "Gig", Location: "Lagos", StartTime: now.Add(time.Hour), TotalCapacity: 100}
	}
	five, tooMany := 5, 101

	e, err := valid().toEvent(now)
	if err != nil {
		t.Fatal(err)
	}
	if e.AvailableTickets != 100 || e.Category != "Music" || e.TicketsTypes != "FREE" {
		t.Errorf("defaults not applied: %+v", e)
	}

	in := valid()
	in.AvailableTickets, in.Category, in.PriceCents = &five, "tech", 2500
	if e, err = in.toEvent(now); err != nil || e.AvailableTickets != 5 || e.Category != "Tech" || e.TicketsTypes != "General Admission" {
		t.Errorf("toEvent = %+v, %v", e, err)
	}

	bad := map[string]func(*eventInput){
		"eventName":        func(i *eventInput) { i.EventName = "  " },
		"location":         func(i *eventInput) { i.Location = "" },
		"date":             func(i *eventInput) { i.StartTime = now.Add(-time.Hour) },
		"totalCapacity":    func(i *eventInput) { i.TotalCapacity = 0 },
		"availableTickets": func(i *eventInput) { i.AvailableTickets = &tooMany },
		"priceCents":       func(i *eventInput) { i.PriceCents = -1 },
		"category":         func(i *eventInput) { i.Category = "Knitting" },
		"imageUrl":         func(i *eventInput) { i.ImageURL = "javascript:alert(1)" },
	}
	for field, mutate := range bad {
		in := valid()
		mutate(&in)
		if _, err := in.toEvent(now); err == nil || err.Field != field {
			t.Errorf("%s: got %v, want error on that field", field, err)
		}
	}
}

func TestRequireAuthentication(t *testing.T) {
	app := newTestApp(t)
	var gotUser string
	h := app.RequireAuthentication(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUser, _ = extractUserFromContext(r.Context())
	}))
	future := time.Now().Add(time.Hour).Unix()

	tests := []struct {
		name   string
		token  string
		status int
	}{
		{"missing", "", http.StatusUnauthorized},
		{"garbage", "not-a-jwt", http.StatusUnauthorized},
		{"expired", signed(t, jwt.SigningMethodHS256, []byte(testSecret), jwt.MapClaims{"userID": "u1", "exp": time.Now().Add(-time.Hour).Unix()}), http.StatusUnauthorized},
		{"no exp", signed(t, jwt.SigningMethodHS256, []byte(testSecret), jwt.MapClaims{"userID": "u1"}), http.StatusUnauthorized},
		{"wrong key", signed(t, jwt.SigningMethodHS256, []byte("another-secret-of-at-least-32-bytes!!"), jwt.MapClaims{"userID": "u1", "exp": future}), http.StatusUnauthorized},
		{"wrong alg", signed(t, jwt.SigningMethodHS512, []byte(testSecret), jwt.MapClaims{"userID": "u1", "exp": future}), http.StatusUnauthorized},
		{"alg none", signed(t, jwt.SigningMethodNone, jwt.UnsafeAllowNoneSignatureType, jwt.MapClaims{"userID": "u1", "exp": future}), http.StatusUnauthorized},
		{"non-string user", signed(t, jwt.SigningMethodHS256, []byte(testSecret), jwt.MapClaims{"userID": 42, "exp": future}), http.StatusUnauthorized},
		{"valid", signed(t, jwt.SigningMethodHS256, []byte(testSecret), jwt.MapClaims{"userID": "u1", "exp": future}), http.StatusOK},
	}
	for _, tt := range tests {
		gotUser = ""
		req := httptest.NewRequest(http.MethodPost, "/x", nil)
		if tt.token != "" {
			req.Header.Set("Authorization", "Bearer "+tt.token)
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != tt.status {
			t.Errorf("%s: status %d, want %d", tt.name, rec.Code, tt.status)
		}
		if tt.status == http.StatusOK && gotUser != "u1" {
			t.Errorf("%s: user in context = %q, want u1", tt.name, gotUser)
		}
	}

	// The cookie works too, and generateToken's output is accepted.
	token, err := app.generateToken("u2")
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.AddCookie(&http.Cookie{Name: "access", Value: token})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || gotUser != "u2" {
		t.Errorf("cookie auth: status %d user %q", rec.Code, gotUser)
	}

	// A browser navigating to a page is sent to the login screen.
	req = httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != "/login" {
		t.Errorf("browser redirect: status %d location %q", rec.Code, rec.Header().Get("Location"))
	}
}

func TestRecoverPanic(t *testing.T) {
	app := newTestApp(t)
	app.ErroMessage.SetOutput(new(strings.Builder))
	h := app.RecoverPanic(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("boom") }))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status %d, want 500", rec.Code)
	}
}

// TestRejectedBeforeDatabase drives the full router for requests that must
// fail before any query runs (the test app has no database).
func TestRejectedBeforeDatabase(t *testing.T) {
	app := newTestApp(t)
	app.InfoLogger.SetOutput(new(strings.Builder))
	router := app.routers()
	token, _ := app.generateToken("11111111-1111-1111-1111-111111111111")

	tests := []struct {
		method, path, body string
		authed             bool
		status             int
		field              string
	}{
		{"POST", "/events", `{}`, false, http.StatusUnauthorized, ""},
		{"GET", "/me", "", false, http.StatusUnauthorized, ""},
		{"GET", "/me/tickets", "", false, http.StatusUnauthorized, ""},
		{"POST", "/events/abc/tickets", `{"quantity":1}`, false, http.StatusUnauthorized, ""},
		{"POST", "/events", `{not json`, true, http.StatusBadRequest, ""},
		{"POST", "/events", `{"eventName":"x"}`, true, http.StatusUnprocessableEntity, "location"},
		{"PUT", "/events/not-a-uuid", `{}`, true, http.StatusNotFound, ""},
		{"GET", "/events/not-a-uuid", "", false, http.StatusNotFound, ""},
		{"POST", "/events/11111111-1111-1111-1111-111111111111/tickets", `{"quantity":0}`, true, http.StatusUnprocessableEntity, "quantity"},
		{"POST", "/events/11111111-1111-1111-1111-111111111111/tickets", `{"quantity":11}`, true, http.StatusUnprocessableEntity, "quantity"},
		{"DELETE", "/tickets/nope", "", true, http.StatusNotFound, ""},
		{"POST", "/createUser", `{"userName":"ola","emailAddress":"bad","password":"password1"}`, false, http.StatusUnprocessableEntity, "emailAddress"},
		{"GET", "/events?limit=0", "", false, http.StatusBadRequest, "limit"},
		{"GET", "/events?page=x", "", false, http.StatusBadRequest, "page"},
	}
	for _, tt := range tests {
		req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
		if tt.authed {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != tt.status {
			t.Errorf("%s %s: status %d, want %d (body %q)", tt.method, tt.path, rec.Code, tt.status, rec.Body.String())
			continue
		}
		if tt.field != "" {
			var e apiError
			if err := json.Unmarshal(rec.Body.Bytes(), &e); err != nil || e.Field != tt.field {
				t.Errorf("%s %s: error body %q, want field %q", tt.method, tt.path, rec.Body.String(), tt.field)
			}
		}
	}
}

func TestLogout(t *testing.T) {
	app := newTestApp(t)
	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	app.LogoutHandler(rec, req)
	if rec.Code != http.StatusNoContent || rec.Header().Get("HX-Redirect") != "/" {
		t.Errorf("status %d, HX-Redirect %q", rec.Code, rec.Header().Get("HX-Redirect"))
	}
	cookies := rec.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != "access" || cookies[0].MaxAge >= 0 {
		t.Errorf("logout cookie = %+v, want an expired access cookie", cookies)
	}
}

func TestBuildEventCard(t *testing.T) {
	card := buildEventCard(&model.Event{Id: "e1", EventName: "Gig", AvailableTickets: 0, Category: "Arts", PriceCents: 2550})
	if card.StatusLabel != "Sold Out" || card.Category != "Arts" || card.PriceLabel != "$25.50" {
		t.Errorf("card = %+v", card)
	}
	if !strings.Contains(card.ImageURL, "picsum.photos/seed/e1") {
		t.Errorf("fallback image = %q", card.ImageURL)
	}
	card = buildEventCard(&model.Event{Id: "e2", AvailableTickets: 3, ImageURL: "https://x/y.jpg"})
	if card.StatusLabel != "3 left" || card.PriceLabel != "Free" || card.ImageURL != "https://x/y.jpg" {
		t.Errorf("card = %+v", card)
	}
}
