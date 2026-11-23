package model

import (
	"time"
)

// data model Define Data model

type User struct {
	ID           string    `json:"id"`
	UserName     string    `json:"userName"`
	EmailAddress string    `json:"emailAddress"`
	TimeStamp    time.Time `json:"timeStamp"`
	HashPassword string    `json:"-"`
}

type Event struct {
	Id               string    `json:"id"`
	AritistId        *string   `json:"artistid,omitempty"`
	Location         string    `json:"location"`
	StartTime        time.Time `json:"date"`
	AvailableTickets int       `json:"availableTickets"`
	TicketsTypes     string    `json:"ticketsTypes"`
	EventName        string    `json:"eventName"`
	TotalCapacity    int       `json:"totalCapacity"`
	OrganizerId      *string   `json:"organizerId,omitempty"`
	Description      string    `json:"description"`
	Category         string    `json:"category"`
	PriceCents       int       `json:"priceCents"`
	ImageURL         string    `json:"imageUrl"`
}

// Ticket is one booking: a user holding Quantity seats for an event.
// EventName, StartTime and Location are denormalised from events when
// listing a user's tickets, so the client needs no second request.
type Ticket struct {
	Id         string    `json:"id"`
	EventId    string    `json:"eventId"`
	UserId     string    `json:"userId"`
	Quantity   int       `json:"quantity"`
	TotalCents int       `json:"totalCents"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"createdAt"`
	EventName  string    `json:"eventName,omitempty"`
	StartTime  time.Time `json:"eventDate,omitempty"`
	Location   string    `json:"location,omitempty"`
}

const (
	TicketConfirmed = "confirmed"
	TicketCancelled = "cancelled"
)

type Artists struct {
	ArtistId     string   `json:"-"`
	Name         string   `json:"name"`
	CreationDate string   `json:"creationDate"`
	Image        string   `json:"image"`
	Members      []string `json:"members"`
	ExternalId   int      `json:"id"`
	FirstAlbum   string   `json:"firstAlbum"`
}

type Location struct {
	Id        int      `json:"id"`
	Locations []string `json:"locations"`
}

type Date struct {
	Id    int      `json:"id"`
	Dates []string `json:"dates"`
}

type Relation struct {
	Id       int
	JsonData map[string][]string `json:"datesLocations"`
}

// writing the go routines  for my reading
type Data struct {
	Artists   []Artists
	Locations []Location
	Dates     []Date
	Relations []Relation
}
