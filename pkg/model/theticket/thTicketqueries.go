package theticket

import (
	"context"

	"github.com/OlayiwolaSherrifSalawu/theTicketing.git/pkg/model"
)

func (d *TheTicketModel) InsertT(ctx context.Context, Event *model.Event) error {
	stmt := "INSERT INTO events(location, start_time, available_tickets, tickets_types, events_name, total_capacity) values($1, $2, $3, $4, $5, $6) RETURNING id"
	tx, err := d.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	err = tx.QueryRowContext(ctx, stmt, Event.Location, Event.StartTime, Event.AvailableTickets, Event.TicketsTypes, Event.EventName, Event.TotalCapacity).Scan(&Event.Id)
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}
