package data

import (
	"context"
	"database/sql"
	"math/rand/v2"
	"os"
	"testing"

	"github.com/OlayiwolaSherrifSalawu/theTicketing.git/api"
	"github.com/OlayiwolaSherrifSalawu/theTicketing.git/pkg/model"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func TestFormatLocation(t *testing.T) {
	title := cases.Title(language.English)
	for raw, want := range map[string]string{
		"north_carolina-usa":      "North Carolina, USA",
		"london-uk":               "London, UK",
		"playa_del_carmen-mexico": "Playa Del Carmen, Mexico",
		"saitama":                 "Saitama",
	} {
		if got := formatLocation(title, raw); got != want {
			t.Errorf("formatLocation(%q) = %q, want %q", raw, got, want)
		}
	}
}

func TestSaveDataIsIdempotent(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	db, err := sql.Open("postgres", url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() }) // registered first, so it runs after the row cleanups
	api.MigrationsSource = "file://../../database"
	if err := api.RunDBMigration(db); err != nil {
		t.Fatal(err)
	}

	externalID := 1_000_000 + rand.IntN(1_000_000)
	data := &model.Data{
		Artists: []model.Artists{{
			Name:         "Test Artist",
			Image:        "https://example.com/a.jpg",
			CreationDate: "2020",
			Members:      []string{"One", "Two"},
			ExternalId:   externalID,
			FirstAlbum:   "01-01-2021",
		}},
		Relations: []model.Relation{{
			Id:       externalID,
			JsonData: map[string][]string{"seattle-usa": {"*01-02-2030", "02-02-2030"}},
		}},
	}
	t.Cleanup(func() { db.Exec("DELETE FROM artist WHERE external_id = $1", externalID) })

	artists, events, err := SaveData(context.Background(), db, data)
	if err != nil {
		t.Fatal(err)
	}
	if artists != 1 || events != 2 {
		t.Fatalf("first run inserted %d artists, %d events; want 1, 2", artists, events)
	}

	var location string
	if err := db.QueryRow(`SELECT e.location FROM events e JOIN artist a ON a.artist_id = e.artist_id
		WHERE a.external_id = $1 LIMIT 1`, externalID).Scan(&location); err != nil {
		t.Fatal(err)
	}
	if location != "Seattle, USA" {
		t.Errorf("location = %q, want %q", location, "Seattle, USA")
	}

	artists, events, err = SaveData(context.Background(), db, data)
	if err != nil {
		t.Fatal(err)
	}
	if artists != 0 || events != 0 {
		t.Fatalf("second run inserted %d artists, %d events; want 0, 0", artists, events)
	}
}
