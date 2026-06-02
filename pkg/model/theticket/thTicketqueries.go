package theticket

import (
	"context"

	"github.com/OlayiwolaSherrifSalawu/theTicketing.git/pkg/model"
)

func (d *TheTicketModel) InsertT(ctx *context.Context, Event *model.Event) error {
	stmt := "INSERT INTO events(id, location, start_time, available_tickets, tickets_types, events_name, total_capacity, created_at) values($1, $2, $3, $4, $5, $6, $7, $8)"
	tx, err := d.DB.BeginTx(*ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(*ctx, stmt, Event.Id, Event.Location, Event.StartTime, Event.AvailableTickets, Event.TicketsTypes, Event.EventName, Event.TotalCapacity)
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}
