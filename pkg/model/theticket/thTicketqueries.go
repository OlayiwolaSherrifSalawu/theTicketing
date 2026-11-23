package theticket

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/OlayiwolaSherrifSalawu/theTicketing.git/pkg/model"
)

// eventColumns is the single column list every event query selects, in
// the order scanEvent expects.
const eventColumns = "id, artist_id, location, start_time, available_tickets, tickets_types, event_name, total_capacity, organizer_id, description, category, price_cents, image_url"

const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// EventFilter narrows GetAllEvents. Empty strings mean "no filter".
// Limit <= 0 falls back to DefaultPageSize; it is capped at MaxPageSize.
type EventFilter struct {
	Name     string
	Location string
	Category string
	Limit    int
	Offset   int
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanEvent(row rowScanner) (*model.Event, error) {
	e := &model.Event{}
	var artistID, organizerID sql.NullString
	err := row.Scan(&e.Id, &artistID, &e.Location, &e.StartTime, &e.AvailableTickets, &e.TicketsTypes,
		&e.EventName, &e.TotalCapacity, &organizerID, &e.Description, &e.Category, &e.PriceCents, &e.ImageURL)
	if err != nil {
		return nil, err
	}
	if artistID.Valid {
		e.AritistId = &artistID.String
	}
	if organizerID.Valid {
		e.OrganizerId = &organizerID.String
	}
	return e, nil
}

func collectEvents(rows *sql.Rows) ([]*model.Event, error) {
	defer rows.Close()
	events := []*model.Event{}
	for rows.Next() {
		e, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

// InsertT creates an event and fills in its generated ID.
func (d *TheTicketModel) InsertT(ctx context.Context, e *model.Event) error {
	stmt := `INSERT INTO events(location, start_time, available_tickets, tickets_types, event_name, total_capacity,
		organizer_id, description, category, price_cents, image_url)
		VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id`
	err := d.DB.QueryRowContext(ctx, stmt, e.Location, e.StartTime, e.AvailableTickets, e.TicketsTypes, e.EventName,
		e.TotalCapacity, e.OrganizerId, e.Description, e.Category, e.PriceCents, e.ImageURL).Scan(&e.Id)
	return translate(err)
}

func (d *TheTicketModel) GetAllEvents(ctx context.Context, f EventFilter) ([]*model.Event, error) {
	stmt := "SELECT " + eventColumns + " FROM events WHERE 1=1"
	args := []any{}
	add := func(clause, val string) {
		args = append(args, val)
		stmt += " AND " + clause + " $" + strconv.Itoa(len(args))
	}
	if f.Location != "" {
		add("location ILIKE", "%"+f.Location+"%")
	}
	if f.Name != "" {
		add("event_name ILIKE", "%"+f.Name+"%")
	}
	if f.Category != "" {
		add("category ILIKE", f.Category)
	}

	limit := f.Limit
	if limit <= 0 {
		limit = DefaultPageSize
	}
	if limit > MaxPageSize {
		limit = MaxPageSize
	}
	offset := max(f.Offset, 0)
	args = append(args, limit, offset)
	stmt += " ORDER BY start_time, id LIMIT $" + strconv.Itoa(len(args)-1) + " OFFSET $" + strconv.Itoa(len(args))

	rows, err := d.DB.QueryContext(ctx, stmt, args...)
	if err != nil {
		return nil, err
	}
	return collectEvents(rows)
}

func (d *TheTicketModel) GetEventByID(ctx context.Context, id string) (*model.Event, error) {
	e, err := scanEvent(d.DB.QueryRowContext(ctx, "SELECT "+eventColumns+" FROM events WHERE id = $1", id))
	if err != nil {
		return nil, translate(err)
	}
	return e, nil
}

func (d *TheTicketModel) GetEventsByOrganizer(ctx context.Context, organizerID string) ([]*model.Event, error) {
	rows, err := d.DB.QueryContext(ctx,
		"SELECT "+eventColumns+" FROM events WHERE organizer_id = $1 ORDER BY start_time, id", organizerID)
	if err != nil {
		return nil, err
	}
	return collectEvents(rows)
}

// UpdateEvent overwrites the editable fields of an event the caller owns.
// Changing TotalCapacity shifts AvailableTickets by the same amount, so
// tickets already sold stay sold; a capacity below what's been sold fails
// with ErrCapacityTooLow. AvailableTickets on the input is ignored.
func (d *TheTicketModel) UpdateEvent(ctx context.Context, e *model.Event, organizerID string) (*model.Event, error) {
	// Right-hand sides in SET see the row's old values, so the
	// available_tickets expression uses the pre-update total_capacity.
	stmt := `UPDATE events SET
			event_name = $3, location = $4, start_time = $5, tickets_types = $6, description = $7,
			category = $8, price_cents = $9, image_url = $10,
			available_tickets = available_tickets + ($11 - total_capacity),
			total_capacity = $11
		WHERE id = $1 AND organizer_id = $2 AND available_tickets + ($11 - total_capacity) >= 0
		RETURNING ` + eventColumns
	updated, err := scanEvent(d.DB.QueryRowContext(ctx, stmt, e.Id, organizerID, e.EventName, e.Location, e.StartTime,
		e.TicketsTypes, e.Description, e.Category, e.PriceCents, e.ImageURL, e.TotalCapacity))
	if err == nil {
		return updated, nil
	}
	if translate(err) != ErrNotFound {
		return nil, err
	}
	if err := d.checkOwner(ctx, e.Id, organizerID); err != nil {
		return nil, err
	}
	return nil, ErrCapacityTooLow
}

// DeleteEvent removes an event the caller owns. Its tickets go with it
// (ON DELETE CASCADE).
func (d *TheTicketModel) DeleteEvent(ctx context.Context, id, organizerID string) error {
	res, err := d.DB.ExecContext(ctx, "DELETE FROM events WHERE id = $1 AND organizer_id = $2", id, organizerID)
	if err != nil {
		return err
	}
	if n, err := res.RowsAffected(); err != nil {
		return err
	} else if n == 0 {
		return d.checkOwner(ctx, id, organizerID)
	}
	return nil
}

// checkOwner explains why an owner-scoped write matched no rows:
// ErrNotFound if the event doesn't exist, ErrForbidden if someone else
// owns it, nil if the caller does own it.
func (d *TheTicketModel) checkOwner(ctx context.Context, id, organizerID string) error {
	var owner sql.NullString
	err := d.DB.QueryRowContext(ctx, "SELECT organizer_id FROM events WHERE id = $1", id).Scan(&owner)
	if err != nil {
		return translate(err)
	}
	if !owner.Valid || owner.String != organizerID {
		return ErrForbidden
	}
	return nil
}
