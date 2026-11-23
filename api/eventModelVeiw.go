package api

import (
	"fmt"

	"github.com/OlayiwolaSherrifSalawu/theTicketing.git/pkg/model"
)

// EventCard is what the home page template renders for one event.
type EventCard struct {
	ID          string
	Name        string
	Location    string
	DateLabel   string
	StatusLabel string
	StatusClass string // tailwind classes for the status pill
	Category    string
	PriceLabel  string
	ImageURL    string
}

func buildEventCard(e *model.Event) EventCard {
	status, statusClass := "Available", "bg-emerald-50 text-emerald-700"
	switch {
	case e.AvailableTickets <= 0:
		status, statusClass = "Sold Out", "bg-ink-faint/20 text-ink-soft"
	case e.AvailableTickets <= 15:
		status, statusClass = fmt.Sprintf("%d left", e.AvailableTickets), "bg-amber-50 text-amber-700"
	}

	image := e.ImageURL
	if image == "" {
		// Stable stock photo per event until the organiser uploads one.
		image = fmt.Sprintf("https://picsum.photos/seed/%s/640/360", e.Id)
	}

	return EventCard{
		ID:          e.Id,
		Name:        e.EventName,
		Location:    e.Location,
		DateLabel:   e.StartTime.Format("Jan 2, 2006"),
		StatusLabel: status,
		StatusClass: statusClass,
		Category:    e.Category,
		PriceLabel:  formatPrice(e.PriceCents),
		ImageURL:    image,
	}
}

func formatPrice(cents int) string {
	if cents <= 0 {
		return "Free"
	}
	return fmt.Sprintf("$%d.%02d", cents/100, cents%100)
}

func buildEventCards(events []*model.Event) []EventCard {
	cards := make([]EventCard, 0, len(events))
	for _, e := range events {
		cards = append(cards, buildEventCard(e))
	}
	return cards
}
