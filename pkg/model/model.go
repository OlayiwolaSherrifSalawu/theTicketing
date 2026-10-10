package model

import "time"

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
	Location         string    `json:"location"`
	StartTime        time.Time `json:"date"`
	AvailableTickets int       `json:"availableTickets"`
	TicketsTypes     string    `json:"ticketsTypes"`
	EventName        string    `json:"eventName"`
	TotalCapacity    int       `json:"totalCapacity"`
}

type Artists struct {
	ArtistId     string   `json:"-"`
	Name         string   `json:"name"`
	CreationDate int      `json:"creationDate"`
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

// getting this stuff started
