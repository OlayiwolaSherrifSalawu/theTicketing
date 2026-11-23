package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/OlayiwolaSherrifSalawu/theTicketing.git/pkg/model/theticket"
)

type bookingInput struct {
	Quantity int `json:"quantity"`
}

// BookTicketsHandler books seats on the event in the path for the
// logged-in user. 201 with the ticket, 409 when sold out, 422 once the
// event has started, 404 for an unknown event.
func (a *Application) BookTicketsHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := extractUserFromContext(r.Context())
	eventID, ok := a.pathID(w, r)
	if !ok {
		return
	}
	var in bookingInput
	if err := decodeJSON(w, r, &in, authBodyLimit); err != nil {
		a.clientErrorJSON(w, http.StatusBadRequest, "", "Request body must be valid JSON.")
		return
	}
	if in.Quantity < theticket.MinTicketsPerBooking || in.Quantity > theticket.MaxTicketsPerBooking {
		a.clientErrorJSON(w, http.StatusUnprocessableEntity, "quantity",
			fmt.Sprintf("Quantity must be between %d and %d.", theticket.MinTicketsPerBooking, theticket.MaxTicketsPerBooking))
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), dbTimeout)
	defer cancel()
	ticket, err := a.TheTicket.BookTickets(ctx, eventID, userID, in.Quantity)
	if err != nil {
		a.storeError(w, err)
		return
	}
	if err := writeJSON(w, http.StatusCreated, ticket); err != nil {
		a.ErroMessage.Println(err)
	}
}

// MyTicketsHandler lists the logged-in user's bookings, cancelled ones
// included (check status).
func (a *Application) MyTicketsHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := extractUserFromContext(r.Context())
	ctx, cancel := context.WithTimeout(r.Context(), dbTimeout)
	defer cancel()
	tickets, err := a.TheTicket.GetTicketsByUser(ctx, userID)
	if err != nil {
		a.serverError(w, err)
		return
	}
	if err := writeJSON(w, http.StatusOK, tickets); err != nil {
		a.ErroMessage.Println(err)
	}
}

// CancelTicketHandler cancels one of the caller's bookings and releases
// its seats. Someone else's ticket is reported as 404, not 403, so ticket
// IDs can't be probed.
func (a *Application) CancelTicketHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := extractUserFromContext(r.Context())
	ticketID, ok := a.pathID(w, r)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), dbTimeout)
	defer cancel()
	if err := a.TheTicket.CancelTicket(ctx, ticketID, userID); err != nil {
		a.storeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
