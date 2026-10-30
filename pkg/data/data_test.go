package data

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/OlayiwolaSherrifSalawu/theTicketing.git/pkg/model"
	"github.com/jackc/pgx/v5"
)

func TestSaveData(t *testing.T) {

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		return
	}
	err = conn.Ping(t.Context())
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		return
	}
	data := model.Data{
		Artists: []model.Artists{
			{
				ArtistId:     "8",
				Name:         "Dusin Oyekan",
				Image:        "https://spiritchant.com.ng/artist/dunsin-oyekan/",
				CreationDate: "2020-4-12",
				Members:      []string{"Theophilus Sunday", "Poju Oyemade"},
				ExternalId:   999,
				FirstAlbum:   "Code red",
			},
		},
	}
	fmt.Print(data)
}
