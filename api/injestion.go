package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/OlayiwolaSherrifSalawu/theTicketing.git/pkg/model"
)

func Fetcher(api string) (*model.Data, error) {

	responseMap := map[string]string{
		"artist":   "",
		"location": "",
		"dates":    "",
		"relation": "",
	}
	client := &http.Client{
		Timeout: time.Second * 3,
	}
	resp, err := client.Get(api)

	if err != nil {
		return &model.Data{}, err
	}
	defer resp.Body.Close()

	var wg sync.WaitGroup

	if resp.StatusCode != http.StatusOK {
		return &model.Data{}, fmt.Errorf("Bad status code not ok.")
	}
	decoded := json.NewDecoder(resp.Body)
	if err := decoded.Decode(&responseMap); err != nil {
		return &model.Data{}, err
	}
	decoded.Decode(&responseMap)
	// making five buffered channerl
	chanArtist := make(chan []model.Artists, 1)
	chanRelations := make(chan []model.Relation, 1)
	chanDate := make(chan []model.Date, 1)
	chanLocation := make(chan []model.Location, 1)
	chanError := make(chan []error, 4)

	for val := range responseMap {
		url := responseMap[val]
		go func() {
			clients := http.Client{
				Timeout: 3 * time.Second,
			}
			resps, err := clients.Get(url)
			if err != nil {
				go errchan(err, chanError)
			}
			defer resps.Body.Close()
			if resps.StatusCode != http.StatusOK {
				go errchan(fmt.Errorf("Bad status code"), chanError)
				return

			}
			respsBody, err := io.ReadAll(resp.Body)
			if err != nil {

			}
			switch val {
			case "artist":
				errs := json.Unmarshal(respsBody, &chanArtist)
				if errs != nil {
					go errchan(errs, chanError)
				}
				wg.Add(1)
			case "location":
				errs := json.Unmarshal(respsBody, &chanLocation)
				if errs != nil {
					go errchan(errs, chanError)
				}
				wg.Add(1)
			case "dates":
				errs := json.Unmarshal(respsBody, &chanDate)
				if errs != nil {
					go errchan(errs, chanError)
				}
				wg.Add(1)
			case "relations":
				errs := json.Unmarshal(respsBody, &chanRelations)
				if errs != nil {
					go errchan(errs, chanError)
				}
				wg.Add(1)

			}
			wg.Done()
		}()

	}
	errSlice := []error{}
	errSlice = <-chanError
	if len(errSlice) > 0 {
		return &model.Data{}, errSlice[0]
	}
	return &model.Data{
		Artists:   <-chanArtist,
		Locations: <-chanLocation,
		Dates:     <-chanDate,
		Relations: <-chanRelations,
	}, nil
}
func errchan(errs error, errchan chan []error) {
	errchan <- append(<-errchan, errs)
}
