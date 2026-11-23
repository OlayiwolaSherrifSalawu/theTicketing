package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/OlayiwolaSherrifSalawu/theTicketing.git/pkg/model"
)

// GroupieAPI is the public artist dataset used to seed the database.
const GroupieAPI = "https://groupietrackers.herokuapp.com/api"

type wrapperLocation struct {
	Index []model.Location `json:"index"`
}
type wrapperDates struct {
	Index []model.Date `json:"index"`
}
type wrapperRelation struct {
	Index []model.Relation `json:"index"`
}

// fetcherEndpoints are the keys the index document must list.
var fetcherEndpoints = []string{"artists", "locations", "dates", "relation"}

// Fetcher downloads the groupie dataset. api is the index URL, which
// returns {"artists": url, "locations": url, "dates": url, "relation": url};
// the four endpoints are fetched concurrently and every one must succeed.
func Fetcher(api string) (*model.Data, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	index := map[string]string{}
	if err := getJSON(client, api, &index); err != nil {
		return nil, fmt.Errorf("fetching index: %w", err)
	}
	for _, key := range fetcherEndpoints {
		if index[key] == "" {
			return nil, fmt.Errorf("index is missing the %q endpoint", key)
		}
	}

	var (
		artists   []model.Artists
		locations wrapperLocation
		dates     wrapperDates
		relations wrapperRelation
	)
	targets := map[string]any{
		"artists":   &artists,
		"locations": &locations,
		"dates":     &dates,
		"relation":  &relations,
	}

	// Each goroutine decodes into its own target and reports into its own
	// slot of errs, so nothing is shared and nothing can block.
	var wg sync.WaitGroup
	errs := make([]error, len(fetcherEndpoints))
	for i, key := range fetcherEndpoints {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := getJSON(client, index[key], targets[key]); err != nil {
				errs[i] = fmt.Errorf("fetching %s: %w", key, err)
			}
		}()
	}
	wg.Wait()
	if err := errors.Join(errs...); err != nil {
		return nil, err
	}

	return &model.Data{
		Artists:   artists,
		Locations: locations.Index,
		Dates:     dates.Index,
		Relations: relations.Index,
	}, nil
}

func getJSON(client *http.Client, url string, dst any) error {
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %s", resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(dst)
}
