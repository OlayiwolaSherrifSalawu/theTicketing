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

type wrapperLocation struct {
	Index []model.Location `json:"index"`
}
type wrapperDates struct {
	Index []model.Date `json:"index"`
}
type wrapperRelation struct {
	Index []model.Relation `json:"index"`
}

func Fetcher(api string) (*model.Data, error) {

	responseMap := map[string]string{}
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

	// making five buffered ch only job annel
	chanArtist := make(chan []model.Artists, 1)
	chanRelations := make(chan []model.Relation, 1)
	chanDate := make(chan []model.Date, 1)
	chanLocation := make(chan []model.Location, 1)
	chanError := make(chan error, 4)
	artistSlice := []model.Artists{}
	locationStruct := &wrapperLocation{}
	dateStruct := &wrapperDates{}
	relationStruct := &wrapperRelation{}
	done := make(chan bool)
	for val := range responseMap {
		url := responseMap[val]
		wg.Add(1)
		go func(endPointKey, endPointUrl string) {
			defer wg.Done()

			clients := http.Client{
				Timeout: 3 * time.Second,
			}
			resps, err := clients.Get(endPointUrl)
			if err != nil {
				go errchan(err, chanError)
				return
			}
			defer resps.Body.Close()
			if resps.StatusCode != http.StatusOK {
				go errchan(fmt.Errorf("Bad status code"), chanError)
				return

			}
			respsBody, err := io.ReadAll(resps.Body)
			if err != nil {
				errchan(err, chanError)
				return
			}
			switch endPointKey {
			case "artists":
				errs := json.Unmarshal(respsBody, &artistSlice)
				if errs != nil {
					errchan(errs, chanError)
					return
				}
				chanArtist <- artistSlice

			case "locations":
				errs := json.Unmarshal(respsBody, &locationStruct)

				if errs != nil {
					errchan(errs, chanError)
					return
				}
				chanLocation <- locationStruct.Index

			case "dates":
				errs := json.Unmarshal(respsBody, &dateStruct)
				if errs != nil {
					errchan(errs, chanError)
					return
				}
				chanDate <- dateStruct.Index

			case "relation":
				errs := json.Unmarshal(respsBody, &relationStruct)
				if errs != nil {
					errchan(errs, chanError)
					return
				}
				chanRelations <- relationStruct.Index
			}

		}(val, url)
	}
	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case err := <-chanError:
		return &model.Data{}, err
	case <-done:

	}
	return &model.Data{
		Artists:   <-chanArtist,
		Locations: <-chanLocation,
		Dates:     <-chanDate,
		Relations: <-chanRelations,
	}, nil
}

func errchan(errs error, errchan chan error) {
	errchan <- errs
}
