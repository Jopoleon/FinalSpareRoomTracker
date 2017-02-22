package scrape

import (
	"bytes"
	"encoding/json"
	"log"
	"testing"

	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
)

func TestScrapeRoomsWithLocation(t *testing.T) {
	goodtestlocation := "westminster"
	badtestlocation := "sadljh1287dask"
	result, err := ScrapeRoomsWithLocation(goodtestlocation)
	if err != nil {
		t.Fatalf("Expected no err, but got %s", err)
	}
	dat := make([]RoomInfo, 11)

	//marresult := json.Unmarshal(result, &dat)
	if err := json.Unmarshal(result, &dat); err != nil {
		t.Fatalf("Expected no err, but got %s", err)
	}
	if dat == nil {
		t.Fatalf("Expected not empty result from scrape, got %s", dat)
	}
	badresult, err := ScrapeRoomsWithLocation(badtestlocation)
	//badresult = nil
	log.Println(badresult)
	expectedErrorString := "Cant find such location, try another, or type it correct!"
	if err == nil {
		t.Fatalf("Expected err %s, but got %s", expectedErrorString, err)
	}

}

func TestScraperHandler(t *testing.T) {

	data := url.Values{}
	data.Set("value", "westminster")
	b := bytes.NewBufferString(data.Encode())

	req, err := http.NewRequest("POST", "/scrape", b)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(ScraperHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := `{"cost"}`

	if strings.Contains(rr.Body.String(), expected) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}
