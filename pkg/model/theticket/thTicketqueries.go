package theticket

import (
	"context"
	"strconv"

	"github.com/OlayiwolaSherrifSalawu/theTicketing.git/pkg/model"
)

func (d *TheTicketModel) InsertT(ctx context.Context, Event *model.Event) error {
	stmt := "INSERT INTO events(location, start_time, available_tickets, tickets_types, event_name, total_capacity) values($1, $2, $3, $4, $5, $6) RETURNING id"
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
	func (d *TheTicketModel) GetAllEvents(ctx context.Context, url map[string]string) ([]*model.Event, error) {
		stmt := "select id, location, start_time, available_tickets, tickets_types, event_name, total_capacity from where 1=1"
		store := []any{}
		count := 1
		if val, ok := url["location"]; ok {
			val := "%" + val + "%"
			stmt = stmt + " AND location ILIKE $" + strconv.Itoa(count)
			store = append(store, val)
			count++
		}
		if val, ok := url["eventName"]; ok {
			val := "%" + val + "%"
			stmt = stmt + " AND event_name ILIKE $" + strconv.Itoa(count)
			store = append(store, val)
			count++
		}

		rows, err := d.DB.QueryContext(ctx, stmt, store...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		Events := []*model.Event{}
		for rows.Next() {
			event := model.Event{}
			err := rows.Scan(&event.Id, &event.Location, &event.StartTime, &event.AvailableTickets, &event.TicketsTypes,&event.EventName, &event.TotalCapacity)
			if err != nil {
				return nil, err
			}
			Events = append(Events, &event)
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return Events, nil
	}
