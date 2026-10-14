package api

import (
	"io"
	"log"
	"net/http"
	"time"

	"github.com/OlayiwolaSherrifSalawu/theTicketing.git/pkg/model"
)

func Fetecher(api string) (*model.Data, error) {
	data := &model.Data{}
	// making five buffered channerl
	chanArtist := make(chan []model.Artists, 1)
	chanRelations := make(chan []model.Relation, 1)
	chanDate := make(chan []model.Date, 1)
	chanLocation := make(chan []model.Location, 1)
	chanError := make(chan []model.Data, 4)

	client := &http.Client{
		Timeout: time.Second * 3,
	}
	resp, err := client.Get(api)

	if err != nil {
		return &model.Data{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalln("Status code bad")
		return &model.Data{}, err
	}
	bytesBO, err := io.ReadAll(resp.Body)
	go func() {

	}()
}
