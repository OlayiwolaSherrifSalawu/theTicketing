package data

import (
	"context"
	"database/sql"
	"strconv"
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
		formationDate := strconv.Itoa(val.CreationDate)
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
				dates = time.Time.Local(date).Format(dates)
				queryStr := ` INSERT INTO events(artistId, location, date, availableTickets, tickeTypes, eventName, totalCapacity)VALUES($1, $2, $3, $4, $5, $6, $7)`
				tx.QueryRowContext(ctx, queryStr, generateString, key, date, 50, "FREE", val.Name, 1000)
			}
		}
	}
	tx.Commit()

	return nil
}
