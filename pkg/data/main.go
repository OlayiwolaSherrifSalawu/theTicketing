package data

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/OlayiwolaSherrifSalawu/theTicketing.git/pkg/model"
	"github.com/lib/pq"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Defaults for events created from the groupie dataset, which has no
// ticketing information of its own.
const (
	seedAvailable = 50
	seedCapacity  = 1000
	seedTicket    = "FREE"
	seedCategory  = "Music"
)

// SaveData stores artists and one event per (artist, location, date) in a
// single transaction. Artists whose external_id is already present are
// skipped, so running it twice inserts nothing new. It reports how many
// artists and events it inserted.
func SaveData(ctx context.Context, db *sql.DB, data *model.Data) (artists, events int, err error) {
	relations := map[int]map[string][]string{}
	for _, rel := range data.Relations {
		relations[rel.Id] = rel.JsonData
	}
	title := cases.Title(language.English)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, 0, err
	}
	defer tx.Rollback()

	for _, artist := range data.Artists {
		var exists bool
		err := tx.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM artist WHERE external_id = $1)", artist.ExternalId).Scan(&exists)
		if err != nil {
			return 0, 0, err
		}
		if exists {
			continue
		}

		var artistID string
		err = tx.QueryRowContext(ctx, `INSERT INTO artist(name, formation_date, image, first_album, members, external_id)
			VALUES($1, $2, $3, $4, $5, $6) RETURNING artist_id`,
			artist.Name, artist.CreationDate, artist.Image, artist.FirstAlbum, pq.Array(artist.Members), artist.ExternalId).Scan(&artistID)
		if err != nil {
			return 0, 0, fmt.Errorf("inserting artist %q: %w", artist.Name, err)
		}
		artists++

		description := describe(artist)
		for place, dates := range relations[artist.ExternalId] {
			location := formatLocation(title, place)
			for _, raw := range dates {
				date, err := time.Parse("02-01-2006", strings.TrimPrefix(raw, "*"))
				if err != nil {
					return 0, 0, fmt.Errorf("artist %q: bad date %q: %w", artist.Name, raw, err)
				}
				_, err = tx.ExecContext(ctx, `INSERT INTO events(artist_id, location, start_time, available_tickets,
						tickets_types, event_name, total_capacity, description, category, price_cents, image_url)
					VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, 0, $10)`,
					artistID, location, date, seedAvailable, seedTicket, artist.Name, seedCapacity,
					description, seedCategory, artist.Image)
				if err != nil {
					return 0, 0, fmt.Errorf("inserting event for %q: %w", artist.Name, err)
				}
				events++
			}
		}
	}

	return artists, events, tx.Commit()
}

func describe(a model.Artists) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s live", a.Name)
	if len(a.Members) > 0 {
		fmt.Fprintf(&b, " — featuring %s", strings.Join(a.Members, ", "))
	}
	b.WriteString(".")
	if a.FirstAlbum != "" {
		fmt.Fprintf(&b, " First album released %s.", a.FirstAlbum)
	}
	return b.String()
}

// formatLocation turns groupie's "north_carolina-usa" into
// "North Carolina, USA": words title-cased, short country codes upper-cased.
func formatLocation(title cases.Caser, raw string) string {
	parts := strings.Split(raw, "-")
	for i, p := range parts {
		p = strings.ReplaceAll(p, "_", " ")
		if i == len(parts)-1 && len(parts) > 1 && len(p) <= 3 {
			parts[i] = strings.ToUpper(p)
		} else {
			parts[i] = title.String(p)
		}
	}
	return strings.Join(parts, ", ")
}
