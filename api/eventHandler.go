package api

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/OlayiwolaSherrifSalawu/theTicketing.git/pkg/model/theticket"
	"github.com/google/uuid"
)

const eventBodyLimit = 64 << 10

// CreateEvent publishes an event owned by the logged-in user.
func (a *Application) CreateEvent(w http.ResponseWriter, r *http.Request) {
	userID, _ := extractUserFromContext(r.Context())
	var in eventInput
	if err := decodeJSON(w, r, &in, eventBodyLimit); err != nil {
		a.clientErrorJSON(w, http.StatusBadRequest, "", "Request body must be valid JSON.")
		return
	}
	event, verr := in.toEvent(time.Now())
	if verr != nil {
		a.clientErrorJSON(w, http.StatusUnprocessableEntity, verr.Field, verr.Message)
		return
	}
	event.OrganizerId = &userID

	ctx, cancel := context.WithTimeout(r.Context(), dbTimeout)
	defer cancel()
	if err := a.TheTicket.InsertT(ctx, event); err != nil {
		a.storeError(w, err)
		return
	}
	if err := writeJSON(w, http.StatusCreated, event); err != nil {
		a.ErroMessage.Println(err)
	}
}

// GetEvent lists events. Query params: eventName, location, category
// (filters), page (1-based) and limit (default 20, max 100).
func (a *Application) GetEvent(w http.ResponseWriter, r *http.Request) {
	filter, verr := eventFilters(r)
	if verr != nil {
		a.clientErrorJSON(w, http.StatusBadRequest, verr.Field, verr.Message)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), dbTimeout)
	defer cancel()
	events, err := a.TheTicket.GetAllEvents(ctx, filter)
	if err != nil {
		a.serverError(w, err)
		return
	}
	if err := writeJSON(w, http.StatusOK, events); err != nil {
		a.ErroMessage.Println(err)
	}
}

func (a *Application) GetEventByID(w http.ResponseWriter, r *http.Request) {
	id, ok := a.pathID(w, r)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), dbTimeout)
	defer cancel()
	event, err := a.TheTicket.GetEventByID(ctx, id)
	if err != nil {
		a.storeError(w, err)
		return
	}
	if err := writeJSON(w, http.StatusOK, event); err != nil {
		a.ErroMessage.Println(err)
	}
}

// UpdateEvent replaces the editable fields of an event the caller owns.
func (a *Application) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	userID, _ := extractUserFromContext(r.Context())
	id, ok := a.pathID(w, r)
	if !ok {
		return
	}
	var in eventInput
	if err := decodeJSON(w, r, &in, eventBodyLimit); err != nil {
		a.clientErrorJSON(w, http.StatusBadRequest, "", "Request body must be valid JSON.")
		return
	}
	event, verr := in.toEvent(time.Now())
	if verr != nil {
		a.clientErrorJSON(w, http.StatusUnprocessableEntity, verr.Field, verr.Message)
		return
	}
	event.Id = id

	ctx, cancel := context.WithTimeout(r.Context(), dbTimeout)
	defer cancel()
	updated, err := a.TheTicket.UpdateEvent(ctx, event, userID)
	if err != nil {
		a.storeError(w, err)
		return
	}
	if err := writeJSON(w, http.StatusOK, updated); err != nil {
		a.ErroMessage.Println(err)
	}
}

func (a *Application) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	userID, _ := extractUserFromContext(r.Context())
	id, ok := a.pathID(w, r)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), dbTimeout)
	defer cancel()
	if err := a.TheTicket.DeleteEvent(ctx, id, userID); err != nil {
		a.storeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// MyEventsHandler lists the events the logged-in user organises.
func (a *Application) MyEventsHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := extractUserFromContext(r.Context())
	ctx, cancel := context.WithTimeout(r.Context(), dbTimeout)
	defer cancel()
	events, err := a.TheTicket.GetEventsByOrganizer(ctx, userID)
	if err != nil {
		a.serverError(w, err)
		return
	}
	if err := writeJSON(w, http.StatusOK, events); err != nil {
		a.ErroMessage.Println(err)
	}
}

// pathID reads the {id} path segment. Anything that isn't a UUID can't
// match a row, so it's a 404 rather than a database error.
func (a *Application) pathID(w http.ResponseWriter, r *http.Request) (string, bool) {
	id := r.PathValue("id")
	if _, err := uuid.Parse(id); err != nil {
		a.clientErrorJSON(w, http.StatusNotFound, "", "Not found.")
		return "", false
	}
	return id, true
}

// eventFilters reads the list filters and pagination from the query string.
func eventFilters(r *http.Request) (theticket.EventFilter, *apiError) {
	q := r.URL.Query()
	f := theticket.EventFilter{
		Name:     q.Get("eventName"),
		Location: q.Get("location"),
		Category: q.Get("category"),
		Limit:    theticket.DefaultPageSize,
	}
	if v := q.Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > theticket.MaxPageSize {
			return f, fieldError("limit", "limit must be between 1 and 100.")
		}
		f.Limit = n
	}
	if v := q.Get("page"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			return f, fieldError("page", "page must be a positive integer.")
		}
		f.Offset = (n - 1) * f.Limit
	}
	return f, nil
}
