package data

import (
	"context"
	"database/sql"

	"github.com/OlayiwolaSherrifSalawu/theTicketing.git/pkg/model"
)

func SaveData(ctx context.Context, db *sql.DB, data *model.Data) {
	ds := map[int]map[string][]string{}
	var generateString string
	for _, val := range data.Relations {
		ds[val.Id] = val.JsonData
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return
	}
	defer tx.Rollback()

	for _, val := range data.Artists {
		SqlQuery := `INSERT INTO artist(name, formation_date, image, first_album, members, external_id) RETURNING artist_id `
		err := tx.QueryRowContext(ctx, SqlQuery, data.Artists, &val.Name, &val.CreationDate, &val.Image, &val.FirstAlbum, &val.Members).Scan(generateString)
		if err != nil {
			return
		}
	}
	tx.Commit()

}
