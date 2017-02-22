package main

import (
	//"fmt"
	//"github.com/PuerkitoBio/goquery"
	//"html/template"
	//"log"
	"net/http"
	//"os"
	//"reflect"
	"testing"
	//"strconv"
	//"errors"
	//"time"
	"bytes"
	"log"
	//"net/http"
	//"github.com/gorilla/sessions"
	"net/http/httptest"
	"net/url"
	"strings"
)

// //Homepage
// 	http.HandleFunc("/", ctl.IndexHandler)
// 	//Scrape for rooms
// 	http.HandleFunc("/scrapelocation", scrape.ScraperHandler)
// 	//trial scrape for unregistered users
// 	http.HandleFunc("/trialscrapelocation", scrape.TrialScraperHandler)
// 	//sign up in system
// 	http.HandleFunc("/signup", SignUpHandler)
// 	//submit signup information
// 	http.HandleFunc("/signupsubmit", ctl.SignUpSubmitHandler)

// 	http.HandleFunc("/confirm", ctl.ConfirmSignUpHandler)

// 	http.HandleFunc("/login", LoginHandler)
// 	http.HandleFunc("/loginsubmit", ctl.LoginSubmitHandler)

// 	http.HandleFunc("/logout", ctl.LogoutSubmitHandler)

// 	http.HandleFunc("/watchlocation", ctl.AddTrackingPair)
// 	http.HandleFunc("/deletewatchlocation", ctl.DeleteTrackerPair)
var ctl *Controller

func TestSignUpSubmitHandler(t *testing.T) {

	data := url.Values{}
	data.Set("username", "testusername")
	data.Add("password", "testpass")
	data.Add("email", "egortictac2@mail.ru")
	data.Add("urlforactivation", "testurl")

	b := bytes.NewBufferString(data.Encode())

	req, err := http.NewRequest("POST", "/signupsubmit", b)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		t.Fatal(err)
	}
	//log.Println("inside test SignUpSubmitHandler")
	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(ctl.SignUpSubmitHandler)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method

	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)
	session, err := store.Get(req, "sessionRooms")
	if err != nil {
		log.Println(err)
		//http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	log.Println(session.Values)

	// Check the status code is what we expect.
	// if status := rr.Code; status != http.StatusOK {
	// 	t.Errorf("handler returned wrong status code: got %v want %v",
	// 		status, http.StatusOK)
	// }

	// Check the response body is what we expect.
	expected := `{"shipHight":  150}`
	log.Println(rr.Body.String())

	if strings.Contains(rr.Body.String(), expected) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}
