package trackinglogic

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"models"
	"utils"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/sessions"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var store = sessions.NewCookieStore([]byte("nRrHLlHcHH0u7fUz25Hje9m7uJ5SnJzP"))

var mongoUrl = "localhost"

//"mongodb://egor2:qwer1234@ds153729.mlab.com:53729/spareroom"

var startUrl = "http://www.spareroom.co.uk/flatshare/search.pl?flatshare_type=offered&location_type=area&search="
var endUrl = "&miles_from_max=1&action=search&templateoveride=&show_results=&submit="

var DBname = "spareroom"

//sendMailLimitThreshold is how many new rooms tracker saves before sending email with list of this rooms
var sendMailLimitThreshold = 20

//trackerCycleTimeStep is how many SECONDS tracker sleeps befor make another cycle of cheking new rooms
var trackerCycleTimeStep = (10 * time.Second)

type Controller struct {
	
	session *mgo.Session 
}

func NewController() (*Controller, error) {
	
	uri := mongoUrl
	if uri == "" {
		return nil, fmt.Errorf("no DB connection string provided")
	}
	session, err := mgo.Dial(uri)
	if err != nil {
		return nil, err
	}
	return &Controller{
		session: session,
	}, nil
}

func (ctl *Controller) StartTrackingAllUsers() {
	
	for {
		log.Println("inside startTrackingAllUsers")
		pairs := ctl.MakeSliceOfTrackPairs()
		
		for _, pair := range pairs {
			go ctl.TrackingScraper(pair.Username, pair.Location)
		}
		time.Sleep(trackerCycleTimeStep)
	}
}

func (ctl *Controller) MakeSliceOfTrackPairs() []models.TrackInfo {
	dbsession := ctl.session.Clone()
	defer dbsession.Close()

	TrackInfoColletion := dbsession.DB(DBname).C("usersTrackInfo")
	
	num, err := TrackInfoColletion.Count()
	if err != nil {
		log.Println(err)
	}
	

	TrackInfoPairs := make([]models.TrackInfo, num)

	err = TrackInfoColletion.Find(bson.M{}).All(&TrackInfoPairs)
	if err != nil {
		log.Println(err)
	}
	
	return TrackInfoPairs
}

func (ctl *Controller) InitCollectionPullScrape(username, location string) error {
	log.Println(username, " Location for scrape: ", location)
	url := startUrl + location + endUrl
	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Fatal(err)
		return err
	}
	var ErrorString = "Cant find such location, try another, or type it correct!"
	if doc.Find("#maincontent ul.listing-results article.panel-listing-result").Text() == "" {
		
		return errors.New(ErrorString)
	}

	dbsession := ctl.session.Clone()
	defer dbsession.Close()

	pullcollectionname := username + location + "pull"
	var pullRoomInfo models.RoomInfo
	doc.Find("#maincontent ul.listing-results article.panel-listing-result").Each(func(i int, s *goquery.Selection) {
		pullRoomInfo = models.RoomInfo{
			Username: username,
			Location: location,
			Title:    s.Find("header.desktop a h1").Text(),
			Cost:     s.Find("strong.listingPrice").First().Text(),
			ImageUrl: s.Find("figure img").AttrOr("src", "No photo"),
		}
		RoomInfoColletion := dbsession.DB(DBname).C(pullcollectionname)
		err = RoomInfoColletion.Insert(pullRoomInfo)
		if err != nil {
			log.Println(err)
			return
		}
	})

	return nil
}

func (ctl *Controller) TrackingScraper(username, location string) error {

	log.Println("TrackingScraper INSIDE")
	url := startUrl + location + endUrl
	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Fatal(err)
		return err
	}
	var ErrorString = "Cant find such location, try another, or type it correct!"
	if doc.Find("#maincontent ul.listing-results article.panel-listing-result").Text() == "" {
		return errors.New(ErrorString)
	}

	dbsession := ctl.session.Clone()
	defer dbsession.Close()
	trackcollectionname := username + location + "track"
	pullcollectionname := username + location + "pull"
	var trackingRoomInfo models.RoomInfo
	doc.Find("#maincontent ul.listing-results article.panel-listing-result").Each(func(i int, s *goquery.Selection) {
		trackingRoomInfo = models.RoomInfo{
			Username: username,
			Location: location,
			Title:    s.Find("header.desktop a h1").Text(),
			Cost:     s.Find("strong.listingPrice").First().Text(),
			ImageUrl: s.Find("figure img").AttrOr("src", "No photo"),
		}
		cpull := dbsession.DB(DBname).C(pullcollectionname)
		result := models.RoomInfo{}
		err = cpull.Find(bson.M{"title": trackingRoomInfo.Title}).One(&result)
		if result.Title == "" || result.Title == " " {
			ctrack := dbsession.DB(DBname).C(trackcollectionname)
			limit, err23 := ctrack.Count()
			log.Println("Amount of rooms in tracker ", trackcollectionname, " is :", limit)
			if err23 != nil {
				log.Printf("Count problem in trackingScraper %v\n", err)

			}
			log.Println("Limit in ", trackcollectionname, " :", limit)
			if (limit + 5) < sendMailLimitThreshold {

				err = ctrack.Insert(trackingRoomInfo)
				if err != nil {
					log.Printf("Insert problem to ctrack in trackingScraper %v\n", err)

				}
				err = cpull.Insert(trackingRoomInfo)
				if err != nil {
					log.Printf("Insert problem to cpull in trackingScraper %v\n", err)
					//return err
				}
				cpulldoclimit, err := cpull.Count()
				if err != nil {
					log.Printf("Count problem of ", pullcollectionname, " in trackingScraper: %v\n", err)
					//return err
				}

				if cpulldoclimit > 100 {
					_, err := cpull.RemoveAll(bson.M{})
					if err != nil {
						log.Printf("RemoveAll of ", pullcollectionname, " in trackingScraper: %v\n", err)
						//return err
					}
				}

			} else {
				log.Println("It's time to send Email with notification to ", username)

				ctrack := dbsession.DB(DBname).C(trackcollectionname)
				crackCount, _ := ctrack.Count()
				sendmailresult := make([]models.RoomInfo, crackCount)
				ctrack.Find(bson.M{}).All(&sendmailresult)
				usersInfocol := dbsession.DB(DBname).C("usersInfo")
				userresults := models.UserInfo{}
				usersInfocol.Find((bson.M{"username": username})).One(&userresults)
				//sending email with list of 11 new rooms
				go utils.SendEmailTrackList(sendmailresult, userresults.Email, userresults.Username, location)

				kill, err := ctrack.RemoveAll(bson.M{})
				if err != nil {
					log.Printf("RemoveAll of ", pullcollectionname, " in trackingScraper: %v\n", err)
					//return err
				}
				log.Println(kill)
			}
		}

	})
	return nil
}

func (ctl *Controller) DeleteTrackerPair(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "sessionRooms")
	if err != nil {
		log.Println(err)

	}
	log.Println(session.Values)

	username := session.Values["username"].(string)
	location := strings.ToLower(r.FormValue("location"))
	session.Save(r, w)
	log.Println("Recievd data from DELETE request: ", username, " ", location)

	dbsession := ctl.session.Clone()
	defer dbsession.Close()
	TrackInfoColletion := dbsession.DB(DBname).C("usersTrackInfo")
	result := models.TrackInfo{}
	err = TrackInfoColletion.Find(bson.M{"username": username}).One(&result)
	log.Printf("Data base trackinfo: %+v \n", result)
	err = TrackInfoColletion.Remove(bson.M{"username": username})

	if err != nil {
		log.Printf("Track remove fail %v\n", err)
		w.Write([]byte("Username: " + username + " isn't tracking or not found" +
			location + " .Check spelling of your username and name of location"))

	} else {
		log.Println("Delete from db completed")
		w.Write([]byte("Location: " + location + " deleted from tracking"))

		trackcollForDrop := username + location + "track"
		pullcollForDrop := username + location + "pull"
		log.Println("Track col: ", trackcollForDrop, " Pull col: ", pullcollForDrop)
		trackcol := dbsession.DB(DBname).C(trackcollForDrop)
		err = trackcol.DropCollection()
		if err != nil {
			log.Printf("Track collection drop problem: %v\n", err)
			w.Write([]byte("Track collection drop problem"))
			return
		}
		pullcol := dbsession.DB(DBname).C(pullcollForDrop)
		err = pullcol.DropCollection()

		if err != nil {
			log.Printf("Pull colletcion drop problem: %v\n", err)
			w.Write([]byte("Pull colletcion drop problem"))
			return
		}
	}

}

func (ctl *Controller) AddTrackingPair(w http.ResponseWriter, r *http.Request) {

	session, err := store.Get(r, "sessionRooms")
	if err != nil {
		log.Println(err)

	}
	log.Println(session.Values)

	session.Save(r, w)
	r.ParseForm()
	username := session.Values["username"].(string)
	location := strings.ToLower(r.FormValue("location"))

	log.Println("AddTrackingPair Recievd data from request: ", username, " ", location)

	dbsession := ctl.session.Clone()
	defer dbsession.Close()

	TrackInfoColletion := dbsession.DB(DBname).C("usersTrackInfo")
	result := models.TrackInfo{}
	if location == "" {
		w.Write([]byte("You cannot track with empty location name"))
		return
	}

	err = TrackInfoColletion.Find(bson.M{"username": username, "location": location}).One(&result)
	if err != nil {
		log.Println(err)
	}
	// checking is user with such name already tracking location
	if result.Username == "" {
		log.Printf("Data base trackinfo: %+v \n", result)
		err = TrackInfoColletion.Insert(&models.TrackInfo{
			Username: username,
			Location: location,
		})
		if err != nil {
			log.Println(err)
		}
		log.Println("Insert in DB completed")

		initerr := ctl.InitCollectionPullScrape(username, location)
		if initerr != nil {
			log.Println(err)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, "Cant find such location, try another, or type it correct!")
			return
		}

		w.Write([]byte("Now you are tracking " + location))
	} else {
		log.Println("User " + username + " already tracking " + location)
		w.Write([]byte("User " + username + " already tracking " + location))
		return
	}

}
