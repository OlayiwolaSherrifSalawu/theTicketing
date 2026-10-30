package data

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/OlayiwolaSherrifSalawu/theTicketing.git/pkg/model"
)

func SaveData(ctx context.Context, db *sql.DB, data *model.Data) error {
	ds := map[int]map[string][]string{}
	for _, val := range data.Relations {
		ds[val.Id] = val.JsonData
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, val := range data.Artists {
		var generateString string
		formationDate := (val.CreationDate)
		SqlQuery := `INSERT INTO artist(name, formation_date, image, first_album, members, external_id)VALUES($1,$2,$3,$4,$5,$6) RETURNING artist_id `
		err := tx.QueryRowContext(ctx, SqlQuery, val.Name, formationDate, val.Image, val.FirstAlbum, val.Members, val.ExternalId).Scan(&generateString)
		if err != nil {
			return err
		}
		for key, vals := range ds[val.ExternalId] {
			key = strings.ReplaceAll(strings.ReplaceAll(key, "-", " "), "_", " ")
			key = strings.ToTitle(key)
			for _, dates := range vals {
				var date time.Time
				dates = strings.TrimPrefix(dates, "*")
				date, err = time.Parse("02-01-2006", dates)
				if err != nil {
					return err
				}
				queryStr := `INSERT INTO events(artist_id, location, start_time, available_tickets, tickets_types, event_name, total_capacity)VALUES($1, $2, $3, $4, $5, $6, $7)`
				_, err = tx.ExecContext(ctx, queryStr, generateString, key, date, 50, "FREE", val.Name, 1000)
				if err != nil {
					return err
				}
			}
		}
	}

	return tx.Commit()
}
