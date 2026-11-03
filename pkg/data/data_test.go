package data

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/OlayiwolaSherrifSalawu/theTicketing.git/api"
	"github.com/OlayiwolaSherrifSalawu/theTicketing.git/pkg/model"
)

func TestSaveData(t *testing.T) {

	conn, err := api.CreateConnection("config.env")
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
	db, err := api.OpenDb(conn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		return
	}
	err = SaveData(context.Background(), db, &data)
	if err!=nil{
		t.Fatalf("this test failed because %s", err)
	}
}
