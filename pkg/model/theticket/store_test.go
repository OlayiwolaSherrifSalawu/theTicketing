package theticket_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/OlayiwolaSherrifSalawu/theTicketing.git/api"
	"github.com/OlayiwolaSherrifSalawu/theTicketing.git/pkg/model"
	"github.com/OlayiwolaSherrifSalawu/theTicketing.git/pkg/model/theticket"
	"github.com/google/uuid"
)

// openStore connects to TEST_DATABASE_URL and migrates it, or skips.
func openStore(t *testing.T) *theticket.TheTicketModel {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	db, err := sql.Open("postgres", url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	api.MigrationsSource = "file://../../../database"
	if err := api.RunDBMigration(db); err != nil {
		t.Fatal(err)
	}
	return &theticket.TheTicketModel{DB: db}
}

func newUser(t *testing.T, s *theticket.TheTicketModel) *model.User {
	t.Helper()
	id := uuid.NewString()
	u := &model.User{ID: id, UserName: "u-" + id[:8], EmailAddress: id + "@example.com", TimeStamp: time.Now(), HashPassword: "x"}
	if err := s.Insert(context.Background(), u); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.DB.Exec("DELETE FROM users WHERE id = $1", id) })
	return u
}

func newEvent(t *testing.T, s *theticket.TheTicketModel, owner *model.User, capacity, available int, start time.Time) *model.Event {
	t.Helper()
	e := &model.Event{
		EventName: "Test " + uuid.NewString()[:8], Location: "Lagos", StartTime: start,
		TotalCapacity: capacity, AvailableTickets: available, TicketsTypes: "GA",
		Category: "Music", PriceCents: 1500, OrganizerId: &owner.ID,
	}
	if err := s.InsertT(context.Background(), e); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.DB.Exec("DELETE FROM events WHERE id = $1", e.Id) })
	return e
}

func TestUserDuplicates(t *testing.T) {
	s := openStore(t)
	ctx := context.Background()
	u := newUser(t, s)

	dupEmail := *u
	dupEmail.ID, dupEmail.UserName = uuid.NewString(), "other-"+u.ID[:8]
	if err := s.Insert(ctx, &dupEmail); !errors.Is(err, theticket.ErrDuplicateEmail) {
		t.Errorf("duplicate email: %v", err)
	}
	dupName := *u
	dupName.ID, dupName.EmailAddress = uuid.NewString(), "other-"+u.EmailAddress
	if err := s.Insert(ctx, &dupName); !errors.Is(err, theticket.ErrDuplicateUsername) {
		t.Errorf("duplicate username: %v", err)
	}
	if _, err := s.GetByID(ctx, uuid.NewString()); !errors.Is(err, theticket.ErrNotFound) {
		t.Errorf("GetByID missing: %v", err)
	}
}

func TestEventOwnership(t *testing.T) {
	s := openStore(t)
	ctx := context.Background()
	owner, other := newUser(t, s), newUser(t, s)
	e := newEvent(t, s, owner, 100, 100, time.Now().Add(24*time.Hour))

	got, err := s.GetEventByID(ctx, e.Id)
	if err != nil || got.EventName != e.EventName || got.OrganizerId == nil || *got.OrganizerId != owner.ID {
		t.Fatalf("GetEventByID = %+v, %v", got, err)
	}
	mine, err := s.GetEventsByOrganizer(ctx, owner.ID)
	if err != nil || len(mine) != 1 || mine[0].Id != e.Id {
		t.Errorf("GetEventsByOrganizer = %v, %v", mine, err)
	}

	edit := *e
	edit.EventName = "Renamed"
	if _, err := s.UpdateEvent(ctx, &edit, other.ID); !errors.Is(err, theticket.ErrForbidden) {
		t.Errorf("update by non-owner: %v", err)
	}
	missing := edit
	missing.Id = uuid.NewString()
	if _, err := s.UpdateEvent(ctx, &missing, owner.ID); !errors.Is(err, theticket.ErrNotFound) {
		t.Errorf("update missing: %v", err)
	}

	// Sell 30, then capacity changes must keep those 30 sold.
	for range 3 {
		if _, err := s.BookTickets(ctx, e.Id, other.ID, 10); err != nil {
			t.Fatal(err)
		}
	}
	edit.TotalCapacity = 50
	updated, err := s.UpdateEvent(ctx, &edit, owner.ID)
	if err != nil || updated.EventName != "Renamed" || updated.AvailableTickets != 20 {
		t.Errorf("update = %+v, %v; want renamed with 20 available", updated, err)
	}
	edit.TotalCapacity = 29
	if _, err := s.UpdateEvent(ctx, &edit, owner.ID); !errors.Is(err, theticket.ErrCapacityTooLow) {
		t.Errorf("capacity below sold: %v", err)
	}

	if err := s.DeleteEvent(ctx, e.Id, other.ID); !errors.Is(err, theticket.ErrForbidden) {
		t.Errorf("delete by non-owner: %v", err)
	}
	if err := s.DeleteEvent(ctx, e.Id, owner.ID); err != nil {
		t.Errorf("delete: %v", err)
	}
	if _, err := s.GetEventByID(ctx, e.Id); !errors.Is(err, theticket.ErrNotFound) {
		t.Errorf("after delete: %v", err)
	}
}

func TestGetAllEventsFilters(t *testing.T) {
	s := openStore(t)
	ctx := context.Background()
	owner := newUser(t, s)
	e := newEvent(t, s, owner, 10, 10, time.Now().Add(time.Hour))

	got, err := s.GetAllEvents(ctx, theticket.EventFilter{Name: e.EventName[5:], Category: "music"})
	if err != nil || len(got) != 1 || got[0].Id != e.Id {
		t.Errorf("filtered = %v, %v", got, err)
	}
	got, err = s.GetAllEvents(ctx, theticket.EventFilter{Name: e.EventName, Category: "Tech"})
	if err != nil || len(got) != 0 {
		t.Errorf("wrong category matched: %v, %v", got, err)
	}
	got, err = s.GetAllEvents(ctx, theticket.EventFilter{Name: e.EventName, Offset: 1})
	if err != nil || len(got) != 0 {
		t.Errorf("offset past only result: %v, %v", got, err)
	}
}

func TestConcurrentBookingNeverOversells(t *testing.T) {
	s := openStore(t)
	ctx := context.Background()
	owner := newUser(t, s)
	e := newEvent(t, s, owner, 10, 10, time.Now().Add(24*time.Hour))
	buyers := make([]*model.User, 20)
	for i := range buyers {
		buyers[i] = newUser(t, s)
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	booked, soldOut := 0, 0
	for _, b := range buyers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := s.BookTickets(ctx, e.Id, b.ID, 1)
			mu.Lock()
			defer mu.Unlock()
			switch {
			case err == nil:
				booked++
			case errors.Is(err, theticket.ErrSoldOut):
				soldOut++
			default:
				t.Errorf("unexpected error: %v", err)
			}
		}()
	}
	wg.Wait()

	if booked != 10 || soldOut != 10 {
		t.Errorf("booked %d, sold out %d; want 10 and 10", booked, soldOut)
	}
	got, err := s.GetEventByID(ctx, e.Id)
	if err != nil || got.AvailableTickets != 0 {
		t.Errorf("available = %d, %v; want 0", got.AvailableTickets, err)
	}
}

func TestBookAndCancel(t *testing.T) {
	s := openStore(t)
	ctx := context.Background()
	owner, buyer, stranger := newUser(t, s), newUser(t, s), newUser(t, s)
	e := newEvent(t, s, owner, 5, 5, time.Now().Add(24*time.Hour))

	ticket, err := s.BookTickets(ctx, e.Id, buyer.ID, 3)
	if err != nil {
		t.Fatal(err)
	}
	if ticket.TotalCents != 4500 || ticket.Status != model.TicketConfirmed {
		t.Errorf("ticket = %+v", ticket)
	}
	if _, err := s.BookTickets(ctx, e.Id, buyer.ID, 3); !errors.Is(err, theticket.ErrSoldOut) {
		t.Errorf("over-book: %v", err)
	}
	if _, err := s.BookTickets(ctx, uuid.NewString(), buyer.ID, 1); !errors.Is(err, theticket.ErrNotFound) {
		t.Errorf("unknown event: %v", err)
	}

	mine, err := s.GetTicketsByUser(ctx, buyer.ID)
	if err != nil || len(mine) != 1 || mine[0].EventName != e.EventName {
		t.Errorf("GetTicketsByUser = %+v, %v", mine, err)
	}

	if err := s.CancelTicket(ctx, ticket.Id, stranger.ID); !errors.Is(err, theticket.ErrNotFound) {
		t.Errorf("cancel by stranger: %v", err)
	}
	if err := s.CancelTicket(ctx, ticket.Id, buyer.ID); err != nil {
		t.Fatal(err)
	}
	if err := s.CancelTicket(ctx, ticket.Id, buyer.ID); !errors.Is(err, theticket.ErrAlreadyCancelled) {
		t.Errorf("double cancel: %v", err)
	}
	got, _ := s.GetEventByID(ctx, e.Id)
	if got.AvailableTickets != 5 {
		t.Errorf("available after cancel = %d, want 5", got.AvailableTickets)
	}
}

func TestStartedEventsAreLocked(t *testing.T) {
	s := openStore(t)
	ctx := context.Background()
	owner, buyer := newUser(t, s), newUser(t, s)
	e := newEvent(t, s, owner, 5, 5, time.Now().Add(24*time.Hour))
	ticket, err := s.BookTickets(ctx, e.Id, buyer.ID, 1)
	if err != nil {
		t.Fatal(err)
	}
	s.DB.Exec("UPDATE events SET start_time = now() - interval '1 hour' WHERE id = $1", e.Id)

	if _, err := s.BookTickets(ctx, e.Id, buyer.ID, 1); !errors.Is(err, theticket.ErrEventStarted) {
		t.Errorf("book started event: %v", err)
	}
	if err := s.CancelTicket(ctx, ticket.Id, buyer.ID); !errors.Is(err, theticket.ErrEventStarted) {
		t.Errorf("cancel started event: %v", err)
	}
	mine, _ := s.GetTicketsByUser(ctx, buyer.ID)
	if len(mine) != 1 || mine[0].Status != model.TicketConfirmed {
		t.Errorf("failed cancel must roll back: %+v", mine)
	}
}
