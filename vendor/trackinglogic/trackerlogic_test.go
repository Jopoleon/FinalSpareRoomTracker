package trackinglogic

import (
	//"errors"
	//"fmt"
	//"html/template"
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	//"os"
	//"strings"
	//"time"
	"testing"

	"models"
	//"scrape"
	//"userlogic"
	//"utils"

	//"github.com/PuerkitoBio/goquery"
	//"github.com/gorilla/sessions"
	//"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var testusername = "testusername"
var testlocation = "westminster"

var trackcollectionname = testusername + testlocation + "track"
var pullcollectionname = testusername + testlocation + "pull"

func TestInitCollectionPullScrape(t *testing.T) {
	testctl, err := NewController()
	if err != nil {
		t.Fatal(err)
	}
	testctl.InitCollectionPullScrape(testusername, testlocation)

	dbsession := testctl.session.Clone()
	defer dbsession.Close()
	result := models.RoomInfo{}
	c := dbsession.DB(DBname).C(pullcollectionname)
	err = c.Find(bson.M{"username": testusername}).One(&result)
	if err != nil {
		t.Fatal(err)
	}
	expectedusername := "testusername"
	expectedlocation := "westminster"
	if result.Username != expectedusername {
		t.Errorf("InitCollectionPullScrape user: got %v want %v",
			result.Username, expectedusername)
	}
	if result.Location != expectedlocation {
		t.Errorf("InitCollectionPullScrape location: got %v want %v",
			result.Location, expectedlocation)
	}

	testctl.session.Close()
}

func TestAddTrackingPair(t *testing.T) {
	testctl, err := NewController()
	if err != nil {
		t.Fatal(err)
	}
	data := url.Values{}
	data.Set("location", testlocation)
	b := bytes.NewBufferString(data.Encode())
	req, err := http.NewRequest("POST", "/watchlocation", b)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	session, err := store.Get(req, "sessionRooms")
	if err != nil {
		log.Println(err)

	}
	session.Values["username"] = testusername
	session.Save(req, rr)
	handler := http.HandlerFunc(testctl.AddTrackingPair)
	handler.ServeHTTP(rr, req)

	dbsession := testctl.session.Clone()
	defer dbsession.Close()
	result := models.TrackInfo{}
	c := dbsession.DB(DBname).C("usersTrackInfo")
	err = c.Find(bson.M{"username": testusername}).One(&result)
	if err != nil {
		t.Fatal(err)
	}
	expectedusername := "testusername"
	expectedlocation := "westminster"
	if result.Username != expectedusername {
		t.Errorf("AddTrackingPair user: got %v want %v",
			result.Username, expectedusername)
	}
	if result.Location != expectedlocation {
		t.Errorf("AddTrackingPair location: got %v want %v",
			result.Location, expectedlocation)
	}
	testctl.session.Close()
}

func contains(s []models.TrackInfo, e models.TrackInfo) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func TestMakeSliceOfTrackPairs(t *testing.T) {
	testctl, err := NewController()
	if err != nil {
		t.Fatal(err)
	}
	result := testctl.MakeSliceOfTrackPairs()

	expected := models.TrackInfo{
		Username: testusername,
		Location: testlocation,
	}

	if !contains(result, expected) {
		t.Errorf("MakeSliceOfTrackPairs doesnt add testuser: got %v \n want contain  %v",
			result, expected)
	}

	testctl.session.Close()
}

func TestTrackingScraper(t *testing.T) {
	testctl, err := NewController()
	if err != nil {
		t.Fatal(err)
	}

	testctl.TrackingScraper(testusername, testlocation)

	// dbsession := testctl.session.Clone()
	// defer dbsession.Close()
	// result := models.TrackInfo{}
	// c := dbsession.DB(DBname).C(trackcollectionname)
	// err = c.Find(bson.M{"username": testusername}).One(&result)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// log.Println("asdASDASDAS@(&^#HBS", result)

	testctl.session.Close()
}

func TestDeleteTrackerPair(t *testing.T) {
	testctl, err := NewController()
	if err != nil {
		t.Fatal(err)
	}
	data := url.Values{}
	data.Set("location", testlocation)
	b := bytes.NewBufferString(data.Encode())
	req, err := http.NewRequest("POST", "/watchlocation", b)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	session, err := store.Get(req, "sessionRooms")
	if err != nil {
		log.Println(err)

	}
	session.Values["username"] = testusername
	session.Save(req, rr)
	handler := http.HandlerFunc(testctl.DeleteTrackerPair)
	handler.ServeHTTP(rr, req)

	dbsession := testctl.session.Clone()
	defer dbsession.Close()
	result := models.TrackInfo{}
	c := dbsession.DB(DBname).C("usersTrackInfo")
	err = c.Find(bson.M{"username": testusername}).One(&result)

	expectederr := "not found"
	if err.Error() != expectederr {
		t.Errorf("DeleteTrackerPair error content: got %v want %v",
			err.Error(), expectederr)
	}

	expected := ""
	if result.Username != expected {
		t.Errorf("DeleteTrackerPair username error: got %v want %v",
			result.Username, expected)
	}

	testctl.session.Close()

}
