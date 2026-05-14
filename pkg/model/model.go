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
	Location         string    `json:"location"`
	StartTime        time.Time `json:"date"`
	AvailableTickets int       `json:"availableTickets"`
	TicketsTypes     string    `json:"ticketsTypes"`
	EventName        string    `json:"eventName"`
	TotalCapacity    int       `json:"totalCapacity"`
}
