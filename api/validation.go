package api

import (
	"net/mail"
	"net/url"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/OlayiwolaSherrifSalawu/theTicketing.git/pkg/model"
)

const (
	minPasswordBytes = 8
	maxPasswordBytes = 72 // bcrypt ignores everything past 72 bytes
	maxDescription   = 5000
	maxCapacity      = 1_000_000
	maxPriceCents    = 10_000_000
)

// EventCategories are the categories an event may be filed under. They
// match the filter pills on the browse page.
var EventCategories = []string{"Music", "Tech", "Arts", "Sports", "Food"}

// validateSignup normalises d in place and reports the first problem.
func validateSignup(d *dto) *apiError {
	d.UserName = strings.TrimSpace(d.UserName)
	d.EmailAddress = strings.ToLower(strings.TrimSpace(d.EmailAddress))

	if n := utf8.RuneCountInString(d.UserName); n < 3 || n > 50 {
		return fieldError("userName", "Username must be 3–50 characters.")
	}
	if addr, err := mail.ParseAddress(d.EmailAddress); err != nil || addr.Address != d.EmailAddress {
		return fieldError("emailAddress", "Enter a valid email address.")
	}
	if len(d.Password) < minPasswordBytes {
		return fieldError("password", "Password must be at least 8 characters long.")
	}
	if len(d.Password) > maxPasswordBytes {
		return fieldError("password", "Password must be at most 72 bytes long.")
	}
	return nil
}

// eventInput is the JSON body for creating or updating an event. JSON
// names match model.Event so clients send the same shape they receive.
type eventInput struct {
	EventName        string    `json:"eventName"`
	Location         string    `json:"location"`
	StartTime        time.Time `json:"date"`
	TotalCapacity    int       `json:"totalCapacity"`
	AvailableTickets *int      `json:"availableTickets"`
	TicketsTypes     string    `json:"ticketsTypes"`
	Description      string    `json:"description"`
	Category         string    `json:"category"`
	PriceCents       int       `json:"priceCents"`
	ImageURL         string    `json:"imageUrl"`
}

// toEvent validates the input and builds the event it describes.
// availableTickets is optional and defaults to the full capacity; it is
// only honoured on create (updates derive it from capacity changes).
func (in eventInput) toEvent(now time.Time) (*model.Event, *apiError) {
	e := &model.Event{
		EventName:     strings.TrimSpace(in.EventName),
		Location:      strings.TrimSpace(in.Location),
		StartTime:     in.StartTime,
		TotalCapacity: in.TotalCapacity,
		TicketsTypes:  strings.TrimSpace(in.TicketsTypes),
		Description:   strings.TrimSpace(in.Description),
		Category:      strings.TrimSpace(in.Category),
		PriceCents:    in.PriceCents,
		ImageURL:      strings.TrimSpace(in.ImageURL),
	}

	switch {
	case e.EventName == "" || utf8.RuneCountInString(e.EventName) > 255:
		return nil, fieldError("eventName", "Event name is required (max 255 characters).")
	case e.Location == "" || utf8.RuneCountInString(e.Location) > 255:
		return nil, fieldError("location", "Location is required (max 255 characters).")
	case e.StartTime.IsZero():
		return nil, fieldError("date", "Start time is required.")
	case !e.StartTime.After(now):
		return nil, fieldError("date", "Start time must be in the future.")
	case e.TotalCapacity < 1 || e.TotalCapacity > maxCapacity:
		return nil, fieldError("totalCapacity", "Capacity must be between 1 and 1,000,000.")
	case e.PriceCents < 0 || e.PriceCents > maxPriceCents:
		return nil, fieldError("priceCents", "Price must be between 0 and 100,000.00.")
	case utf8.RuneCountInString(e.TicketsTypes) > 255:
		return nil, fieldError("ticketsTypes", "Ticket type must be at most 255 characters.")
	case utf8.RuneCountInString(e.Description) > maxDescription:
		return nil, fieldError("description", "Description must be at most 5,000 characters.")
	}

	e.AvailableTickets = e.TotalCapacity
	if in.AvailableTickets != nil {
		if *in.AvailableTickets < 0 || *in.AvailableTickets > e.TotalCapacity {
			return nil, fieldError("availableTickets", "Available tickets must be between 0 and the capacity.")
		}
		e.AvailableTickets = *in.AvailableTickets
	}

	if e.TicketsTypes == "" {
		if e.PriceCents == 0 {
			e.TicketsTypes = "FREE"
		} else {
			e.TicketsTypes = "General Admission"
		}
	}

	if e.Category == "" {
		e.Category = EventCategories[0]
	} else if i := slices.IndexFunc(EventCategories, func(c string) bool { return strings.EqualFold(c, e.Category) }); i >= 0 {
		e.Category = EventCategories[i]
	} else {
		return nil, fieldError("category", "Category must be one of: "+strings.Join(EventCategories, ", ")+".")
	}

	if e.ImageURL != "" {
		u, err := url.Parse(e.ImageURL)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			return nil, fieldError("imageUrl", "Image URL must be an http(s) link.")
		}
	}
	return e, nil
}
