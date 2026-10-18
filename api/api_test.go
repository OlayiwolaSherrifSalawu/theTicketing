package api

import "testing"

func TestFetcher(t *testing.T) {
	data, err := Fetcher("https://groupietrackers.herokuapp.com/api")
	if err != nil {
		t.Fatalf("an error occured: %s", err)
	} else {
		t.Logf("%+v", data)
	}
}
