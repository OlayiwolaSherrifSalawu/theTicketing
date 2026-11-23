// Command seed loads the groupie artist dataset into the database as
// artists and free events. It is safe to run repeatedly.
package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/OlayiwolaSherrifSalawu/theTicketing.git/api"
	"github.com/OlayiwolaSherrifSalawu/theTicketing.git/pkg/data"
)

func main() {
	source := flag.String("source", api.GroupieAPI, "index URL of the artist dataset")
	envFile := flag.String("env", "config.env", "optional env file with database settings")
	flag.Parse()

	conn, err := api.CreateConnection(*envFile)
	if err != nil {
		log.Fatal(err)
	}
	db, err := api.OpenDb(conn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err := api.RunDBMigration(db); err != nil {
		log.Fatal(err)
	}

	dataset, err := api.Fetcher(*source)
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	artists, events, err := data.SaveData(ctx, db, dataset)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("seeded %d new artists and %d events (%d artists in source)", artists, events, len(dataset.Artists))
}
