package data

import (
	"context"
	"database/sql"
	"strconv"

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
		SqlQuery := `INSERT INTO artist(name, formation_date, image, first_album, members, external_id) RETURNING artist_id VALUES($1,$2,$3,$4,$5,$6)`
		err := tx.QueryRowContext(ctx, SqlQuery, data.Artists, val.Name, formationDate, val.Image, val.FirstAlbum, val.Members).Scan(&generateString)
		if err != nil {
			return err
		}
	}
	tx.Commit()

	return nil
}
