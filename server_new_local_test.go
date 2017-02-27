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
	//"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net/http/httptest"
	"net/url"
	//"strings"
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

func TestSignUpSubmitHandler(t *testing.T) {

	ctl, err := NewController()
	if err != nil {
		log.Fatal(err)
	}

	data := url.Values{}
	data.Set("username", "testusername")
	data.Add("password", "testpass")
	data.Add("email", "egortictac@mail.ru")
	data.Add("urlforactivation", "testurl")

	b := bytes.NewBufferString(data.Encode())

	req, err := http.NewRequest("POST", "/signupsubmit", b)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(ctl.SignUpSubmitHandler)

	handler.ServeHTTP(rr, req)
	session, err := store.Get(req, "sessionRooms")
	if err != nil {
		log.Println(err)

	}

	log.Println(session.Values)

	dbsession := ctl.session.Clone()
	defer dbsession.Close()
	c := dbsession.DB(DBname).C("usersInfo")
	result := UserInfo{}
	err = c.Find(bson.M{"username": "testusername"}).One(&result)
	if err != nil {
		t.Fatal(err)
	}

	expectedusername := "testusername"

	if result.Username != expectedusername {
		t.Errorf("SingUpSubmit registred user: got %v want %v",
			result.Username, expectedusername)
	}
}
func TestLoginSubmitHandler(t *testing.T) {
	testctl, err := NewController()
	if err != nil {
		log.Fatal(err)
	}

	data := url.Values{}
	data.Set("username", "testusername")
	data.Add("password", "testpass")
	b := bytes.NewBufferString(data.Encode())

	req, err := http.NewRequest("POST", "/signupsubmit", b)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		t.Fatal(err)
	}
	//log.Println("inside test SignUpSubmitHandler")
	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(testctl.LoginSubmitHandler)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method

	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	session, err := store.Get(req, "sessionRooms")
	if err != nil {
		log.Println(err)
	}
	log.Println(session.Values)
	session.Values["email"] = "egortictac@mail.ru"
	session.Save(req, rr)

	dbsession := ctl.session.Clone()
	defer dbsession.Close()
	c := dbsession.DB(DBname).C("usersInfo")
	result := UserInfo{}
	err = c.Find(bson.M{"username": "testusername"}).One(&result)

	expectedstatus := "false"
	log.Println(rr.Body.String())
	if result.Loggedin != expectedstatus {
		t.Errorf("LoginSubmitHandler registred user: got %v want %v",
			result.Loggedin, expectedstatus)
	}

}

func TestConfirmSignUpHandler(t *testing.T) {

}
