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
	chanArtist := make(chan model.Artists)
	chanRelations := make(chan model.Relation)
	chanDate := make(chan model.Date)
	chanLocation := make(chan model.Location)
	chanData := make(chan model.Data)

	client := &http.Client{
		Timeout: time.Second * 3,
	}
	resp, err := client.Get(api)

	if err != nil {
		log.Println("error getting url")
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
