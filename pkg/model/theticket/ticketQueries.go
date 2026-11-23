package theticket

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/OlayiwolaSherrifSalawu/theTicketing.git/pkg/model"
)

const (
	MinTicketsPerBooking = 1
	MaxTicketsPerBooking = 10
)

// BookTickets reserves qty seats on an event for a user.
//
// The seat count is decremented by a single conditional UPDATE
// ("... WHERE available_tickets >= qty"). Postgres row-locks the event
// for the duration of the transaction, and a concurrent booker blocks on
// that lock and then re-checks the WHERE clause against the committed
// value, so two buyers can never both take the last seat.
func (d *TheTicketModel) BookTickets(ctx context.Context, eventID, userID string, qty int) (*model.Ticket, error) {
	tx, err := d.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var priceCents int
	err = tx.QueryRowContext(ctx, `UPDATE events SET available_tickets = available_tickets - $2
		WHERE id = $1 AND available_tickets >= $2 AND start_time > now()
		RETURNING price_cents`, eventID, qty).Scan(&priceCents)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, d.whyNotBookable(ctx, tx, eventID)
	}
	if err != nil {
		return nil, err
	}

	t := &model.Ticket{
		EventId:    eventID,
		UserId:     userID,
		Quantity:   qty,
		TotalCents: priceCents * qty,
		Status:     model.TicketConfirmed,
	}
	err = tx.QueryRowContext(ctx, `INSERT INTO tickets(event_id, user_id, quantity, total_cents, status)
		VALUES($1, $2, $3, $4, $5) RETURNING id, created_at`,
		t.EventId, t.UserId, t.Quantity, t.TotalCents, t.Status).Scan(&t.Id, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return t, nil
}

// whyNotBookable is called after the booking UPDATE matched no row, to
// tell the caller which of its conditions failed.
func (d *TheTicketModel) whyNotBookable(ctx context.Context, tx *sql.Tx, eventID string) error {
	var start time.Time
	err := tx.QueryRowContext(ctx, "SELECT start_time FROM events WHERE id = $1", eventID).Scan(&start)
	if err != nil {
		return translate(err)
	}
	if !start.After(time.Now()) {
		return ErrEventStarted
	}
	return ErrSoldOut
}

// GetTicketsByUser lists a user's bookings, newest first, with the event's
// name, date and location attached.
func (d *TheTicketModel) GetTicketsByUser(ctx context.Context, userID string) ([]*model.Ticket, error) {
	rows, err := d.DB.QueryContext(ctx, `SELECT t.id, t.event_id, t.user_id, t.quantity, t.total_cents, t.status,
			t.created_at, e.event_name, e.start_time, e.location
		FROM tickets t JOIN events e ON e.id = t.event_id
		WHERE t.user_id = $1
		ORDER BY t.created_at DESC, t.id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tickets := []*model.Ticket{}
	for rows.Next() {
		t := &model.Ticket{}
		if err := rows.Scan(&t.Id, &t.EventId, &t.UserId, &t.Quantity, &t.TotalCents, &t.Status,
			&t.CreatedAt, &t.EventName, &t.StartTime, &t.Location); err != nil {
			return nil, err
		}
		tickets = append(tickets, t)
	}
	return tickets, rows.Err()
}

// CancelTicket cancels one of the user's confirmed bookings and returns its
// seats to the event. Bookings for events that have already started can't
// be cancelled.
func (d *TheTicketModel) CancelTicket(ctx context.Context, ticketID, userID string) error {
	tx, err := d.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var eventID string
	var qty int
	err = tx.QueryRowContext(ctx, `UPDATE tickets SET status = $3
		WHERE id = $1 AND user_id = $2 AND status = $4
		RETURNING event_id, quantity`,
		ticketID, userID, model.TicketCancelled, model.TicketConfirmed).Scan(&eventID, &qty)
	if errors.Is(err, sql.ErrNoRows) {
		return whyNotCancellable(ctx, tx, ticketID, userID)
	}
	if err != nil {
		return err
	}

	res, err := tx.ExecContext(ctx, `UPDATE events SET available_tickets = available_tickets + $2
		WHERE id = $1 AND start_time > now()`, eventID, qty)
	if err != nil {
		return err
	}
	if n, err := res.RowsAffected(); err != nil {
		return err
	} else if n == 0 {
		return ErrEventStarted
	}
	return tx.Commit()
}

func whyNotCancellable(ctx context.Context, tx *sql.Tx, ticketID, userID string) error {
	var owner, status string
	err := tx.QueryRowContext(ctx, "SELECT user_id, status FROM tickets WHERE id = $1", ticketID).Scan(&owner, &status)
	if err != nil {
		return translate(err)
	}
	if owner != userID {
		// Don't reveal that someone else's ticket exists.
		return ErrNotFound
	}
	return ErrAlreadyCancelled
}
