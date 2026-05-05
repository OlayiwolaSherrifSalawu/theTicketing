package model

import "time"

// data model Define Data model

type User struct {
	ID           string    `json:"id"`
	UserName     string    `json:"userName"`
	EmailAddress string    `json:"emailAddress"`
	Password     string    `json:"passWord"`
	TimeStamp    time.Time `json:"timeStamp"`
	HashPassword string    `json:"hashPassword"`
}

type Event struct {
	Location         string    `json:"location"`
	Date             time.Time `json:"date"`
	Eventtime        time.Time `json:"eventTime"`
	AvailableTickets string    `json:"availableTickets"`
	TicketsTypes     string    `json:"ticketsTypes"`
	EventName        string    `json:"eventName"`
	TotalCapacity    int       `json:"totalCapacity"`
}
for i,j:=0,0; i<len(s);  j++{
	
}