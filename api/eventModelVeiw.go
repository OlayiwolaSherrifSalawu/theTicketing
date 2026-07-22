package api

import (
	"fmt"
	"hash/fnv"

	"github.com/OlayiwolaSherrifSalawu/theTicketing.git/pkg/model"
)

// EventCard is what the home page template actually renders. It wraps a
// real model.Event and adds the display fields the current schema has no
// column for (Category, PriceLabel, ImageURL). Those three are explicitly
// PLACEHOLDERS — deterministically derived from the event's own ID so a
// given event always looks the same across renders, not randomly
// reshuffled on every page load. Delete this fabrication the moment
// `events` gains real category/price/image_url columns; StatusLabel and
// DateLabel are the two fields here backed by real data.
type EventCard struct {
	ID          string
	Name        string
	Location    string
	DateLabel   string
	StatusLabel string
	StatusClass string // tailwind classes for the status pill
	Category    string // PLACEHOLDER
	PriceLabel  string // PLACEHOLDER
	ImageURL    string // PLACEHOLDER
}

var placeholderCategories = []string{"Music", "Tech", "Arts", "Sports", "Food"}

func buildEventCard(e *model.Event) EventCard {
	seed := int(fnv32(e.Id))

	status, statusClass := "Available", "bg-emerald-50 text-emerald-700"
	switch {
	case e.AvailableTickets <= 0:
		status, statusClass = "Sold Out", "bg-ink-faint/20 text-ink-soft"
	case e.AvailableTickets <= 15:
		status, statusClass = fmt.Sprintf("%d left", e.AvailableTickets), "bg-amber-50 text-amber-700"
	}

	return EventCard{
		ID:          e.Id,
		Name:        e.EventName,
		Location:    e.Location,
		DateLabel:   e.StartTime.Format("Jan 2, 2006"),
		StatusLabel: status,
		StatusClass: statusClass,

		// --- everything below is fabricated, not real data ---
		Category:   placeholderCategories[seed%len(placeholderCategories)],
		PriceLabel: fmt.Sprintf("$%d.00", 29+(seed%471)),
		ImageURL:   fmt.Sprintf("https://picsum.photos/seed/%s/640/360", e.Id),
	}
}

func buildEventCards(events []*model.Event) []EventCard {
	cards := make([]EventCard, 0, len(events))
	for _, e := range events {
		cards = append(cards, buildEventCard(e))
	}
	return cards
}

func fnv32(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}
