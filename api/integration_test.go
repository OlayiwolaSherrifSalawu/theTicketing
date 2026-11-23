package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestEndToEnd drives the real router against TEST_DATABASE_URL: signup,
// login cookie, event CRUD with ownership, booking, cancelling, logout.
func TestEndToEnd(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	db, err := sql.Open("postgres", url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() }) // registered first, so it runs after the row cleanups
	MigrationsSource = "file://../database"
	if err := RunDBMigration(db); err != nil {
		t.Fatal(err)
	}
	app, err := NewApplication(db, testSecret)
	if err != nil {
		t.Fatal(err)
	}
	app.SecureCookies = false // httptest serves plain http
	app.InfoLogger.SetOutput(io.Discard)
	srv := httptest.NewServer(app.routers())
	defer srv.Close()

	type client struct {
		*http.Client
		email string
	}
	newClient := func(t *testing.T) client {
		jar, _ := cookiejar.New(nil)
		id := uuid.NewString()
		c := client{&http.Client{Jar: jar}, id + "@example.com"}
		t.Cleanup(func() { db.Exec("DELETE FROM users WHERE email_address = $1", c.email) })
		return c
	}
	do := func(t *testing.T, c client, method, path string, body any, wantStatus int, out any) {
		t.Helper()
		var r io.Reader
		if body != nil {
			b, _ := json.Marshal(body)
			r = bytes.NewReader(b)
		}
		req, _ := http.NewRequest(method, srv.URL+path, r)
		resp, err := c.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		raw, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != wantStatus {
			t.Fatalf("%s %s: status %d, want %d; body %s", method, path, resp.StatusCode, wantStatus, raw)
		}
		if out != nil {
			if err := json.Unmarshal(raw, out); err != nil {
				t.Fatalf("%s %s: decoding %s: %v", method, path, raw, err)
			}
		}
	}
	signupAndLogin := func(t *testing.T, c client) {
		t.Helper()
		name := "u" + strings.ReplaceAll(c.email[:12], "-", "")
		user := map[string]string{"userName": name, "emailAddress": c.email, "password": "password123"}
		do(t, c, "POST", "/createUser", user, http.StatusCreated, nil)
		do(t, c, "POST", "/createUser", user, http.StatusConflict, nil)
		do(t, c, "POST", "/login", map[string]string{"emailAddress": c.email, "password": "wrong-password"}, http.StatusUnauthorized, nil)
		var tok map[string]string
		do(t, c, "POST", "/login", map[string]string{"emailAddress": strings.ToUpper(c.email), "password": "password123"}, http.StatusOK, &tok)
		if tok["token"] == "" {
			t.Fatal("login returned no token")
		}
	}

	organizer, buyer := newClient(t), newClient(t)
	signupAndLogin(t, organizer)
	signupAndLogin(t, buyer)

	// The login cookie alone must authenticate.
	var me map[string]any
	do(t, organizer, "GET", "/me", nil, http.StatusOK, &me)
	if me["emailAddress"] != organizer.email || me["password"] != nil {
		t.Errorf("/me = %v", me)
	}

	newEvent := map[string]any{
		"eventName": "E2E " + uuid.NewString()[:8], "location": "Lagos",
		"date": time.Now().Add(48 * time.Hour).Format(time.RFC3339), "totalCapacity": 5,
		"priceCents": 1000, "category": "tech",
	}
	var event map[string]any
	do(t, organizer, "POST", "/events", newEvent, http.StatusCreated, &event)
	id := event["id"].(string)
	t.Cleanup(func() { db.Exec("DELETE FROM events WHERE id = $1", id) })
	if event["category"] != "Tech" || event["availableTickets"].(float64) != 5 {
		t.Errorf("created event = %v", event)
	}

	do(t, buyer, "GET", "/events/"+id, nil, http.StatusOK, nil)
	var mine []map[string]any
	do(t, organizer, "GET", "/me/events", nil, http.StatusOK, &mine)
	if len(mine) != 1 || mine[0]["id"] != id {
		t.Errorf("/me/events = %v", mine)
	}
	do(t, buyer, "PUT", "/events/"+id, newEvent, http.StatusForbidden, nil)
	do(t, buyer, "DELETE", "/events/"+id, nil, http.StatusForbidden, nil)

	var ticket map[string]any
	do(t, buyer, "POST", "/events/"+id+"/tickets", map[string]int{"quantity": 2}, http.StatusCreated, &ticket)
	if ticket["totalCents"].(float64) != 2000 {
		t.Errorf("ticket = %v", ticket)
	}
	do(t, buyer, "POST", "/events/"+id+"/tickets", map[string]int{"quantity": 4}, http.StatusConflict, nil)
	do(t, buyer, "GET", "/events/"+id, nil, http.StatusOK, &event)
	if event["availableTickets"].(float64) != 3 {
		t.Errorf("available after booking = %v, want 3", event["availableTickets"])
	}

	var tickets []map[string]any
	do(t, buyer, "GET", "/me/tickets", nil, http.StatusOK, &tickets)
	if len(tickets) != 1 || tickets[0]["eventName"] != newEvent["eventName"] {
		t.Errorf("/me/tickets = %v", tickets)
	}
	ticketID := ticket["id"].(string)
	do(t, organizer, "DELETE", "/tickets/"+ticketID, nil, http.StatusNotFound, nil)
	do(t, buyer, "DELETE", "/tickets/"+ticketID, nil, http.StatusNoContent, nil)
	do(t, buyer, "DELETE", "/tickets/"+ticketID, nil, http.StatusConflict, nil)
	do(t, buyer, "GET", "/events/"+id, nil, http.StatusOK, &event)
	if event["availableTickets"].(float64) != 5 {
		t.Errorf("available after cancel = %v, want 5", event["availableTickets"])
	}

	// The browse page renders the real price and category.
	resp, err := buyer.Get(srv.URL + "/events-partial?eventName=" + newEvent["eventName"].(string)[4:])
	if err != nil {
		t.Fatal(err)
	}
	page, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if !strings.Contains(string(page), "$10.00") || !strings.Contains(string(page), "Tech") {
		t.Errorf("events partial missing real price/category:\n%s", page)
	}
	do(t, buyer, "GET", "/dashboard", nil, http.StatusOK, nil)

	newEvent["totalCapacity"] = 8
	do(t, organizer, "PUT", "/events/"+id, newEvent, http.StatusOK, &event)
	if event["availableTickets"].(float64) != 8 {
		t.Errorf("available after capacity change = %v, want 8", event["availableTickets"])
	}
	do(t, organizer, "DELETE", "/events/"+id, nil, http.StatusNoContent, nil)
	do(t, organizer, "GET", "/events/"+id, nil, http.StatusNotFound, nil)

	do(t, buyer, "POST", "/logout", nil, http.StatusNoContent, nil)
	do(t, buyer, "GET", "/me", nil, http.StatusUnauthorized, nil)
}
