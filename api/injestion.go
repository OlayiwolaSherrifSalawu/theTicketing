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
	for val := range responseMap {
		url := responseMap[val]
		wg.Add(1)
		go func(endPointKey, endPointUrl string) error {
			defer wg.Done()
			endPointKey = val
			endPointUrl = url
			clients := http.Client{
				Timeout: 3 * time.Second,
			}
			resps, err := clients.Get(endPointUrl)
			if err != nil {
				go errchan(err, chanError)
			}
			defer resps.Body.Close()
			if resps.StatusCode != http.StatusOK {
				go errchan(fmt.Errorf("Bad status code"), chanError)
				return fmt.Errorf("Bad status code")

			}
			respsBody, err := io.ReadAll(resps.Body)
			if err != nil {
				return err
			}
			switch endPointKey {
			case "artist":
				errs := json.Unmarshal(respsBody, &artistSlice)
				if errs != nil {
					go errchan(errs, chanError)
					return errs
				}
				chanArtist <- artistSlice

			case "location":
				errs := json.Unmarshal(respsBody, &locationStruct)
				if errs != nil {
					return errs
				}
				chanLocation <- locationStruct.Index

			case "dates":
				errs := json.Unmarshal(respsBody, &dateStruct)
				if errs != nil {
					go errchan(errs, chanError)
					return errs
				}
				chanDate <- dateStruct.Index

			case "relations":
				errs := json.Unmarshal(respsBody, &relationStruct)
				if errs != nil {
					go errchan(errs, chanError)
					return errs
				}
				chanRelations <- relationStruct.Index
			}
			return nil
		}(val, url)

	}
	var errSlice error
	errSlice = <-chanError
	if errSlice != nil {
		return &model.Data{}, errSlice
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
