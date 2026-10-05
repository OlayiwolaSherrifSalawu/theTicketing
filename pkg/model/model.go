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
	ArtistId      string   `json:"artist_id"`
	Name          string   `json:"name"`
	FormationDate int      `json:"formationDate"`
	ImageUrl      string   `json:"imageUrl"`
	Members       []string `json:"members"`
	ExternalId    int      `json:"externalId"`
}

type Location struct {
	Id         string   `json:"id"`
	EventsData []string `json:"eventsData"`
}

type Date struct {
	Id           string `json:"id"`
	ConcertsDate []int  `json:"concertsDates"`
}

type Relation struct {
	Id       int
	JsonData map[string]string `json:"jsonData"`
}
