package theticket

import (
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

// Sentinel errors the HTTP layer maps onto status codes. Queries return
// these (wrapped or bare) instead of raw driver errors whenever the cause
// is the caller's fault rather than the database's.
var (
	ErrNotFound          = errors.New("record not found")
	ErrForbidden         = errors.New("not the owner of this record")
	ErrDuplicateEmail    = errors.New("email address already registered")
	ErrDuplicateUsername = errors.New("username already taken")
	ErrSoldOut           = errors.New("not enough tickets available")
	ErrEventStarted      = errors.New("event has already started")
	ErrCapacityTooLow    = errors.New("capacity is lower than tickets already sold")
	ErrAlreadyCancelled  = errors.New("ticket is already cancelled")
)

const uniqueViolation = "23505"

// translate maps driver-level errors onto the sentinels above. Anything it
// doesn't recognise is returned unchanged.
func translate(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == uniqueViolation {
		switch pqErr.Constraint {
		case "users_email_address_key":
			return ErrDuplicateEmail
		case "users_username_key":
			return ErrDuplicateUsername
		}
	}
	return err
}
