package main

import (
	//"crypto/rand"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"reflect"
	//"strconv"
	"errors"
	"time"
	//"crypto/sha1"
	//"io/ioutil"
	//"encoding/json"

	"models"
	"scrape"
	"utils"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type UserInfo struct {
	Username      string `json:username bson:username`
	Password      string `json:password bson:password`
	Email         string `json:email bson:email`
	Loggedin      string `json:loggedin bson:loggedin`
	Registred     string `json:registered bson:registered`
	IsActivated   string `json:isActivated bson:isActivated`
	ActivationKey string `json:ActivationKey bson:ActivationKey`
}

var store = sessions.NewCookieStore([]byte("nRrHLlHcHH0u7fUz25Hje9m7uJ5SnJzP"))

//"mongodb://egor2:qwer1234@ds153729.mlab.com:53729/spareroom"
var mongoUrl = "mongodb://egor2:qwer1234@ds153729.mlab.com:53729/spareroom"

var startUrl = "http://www.spareroom.co.uk/flatshare/search.pl?flatshare_type=offered&location_type=area&search="
var endUrl = "&miles_from_max=1&action=search&templateoveride=&show_results=&submit="

var DBname = "spareroom"

func main() {
	ctl, err := NewController()
	if err != nil {
		log.Fatal(err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	fmt.Println("Server started on port: ", port)

	//Homepage
	http.HandleFunc("/", ctl.IndexHandler)
	//Scrape for rooms
	http.HandleFunc("/scrapelocation", scrape.ScraperHandler)
	//trial scrape for unregistered users
	http.HandleFunc("/trialscrapelocation", scrape.TrialScraperHandler)
	//sign up in system
	http.HandleFunc("/signup", SignUpHandler)
	//submit signup information
	http.HandleFunc("/signupsubmit", ctl.SignUpSubmitHandler)

	http.HandleFunc("/confirm", ctl.ConfirmSignUpHandler)

	http.HandleFunc("/login", LoginHandler)
	http.HandleFunc("/loginsubmit", ctl.LoginSubmitHandler)

	http.HandleFunc("/logout", ctl.LogoutSubmitHandler)

	http.HandleFunc("/watchlocation", ctl.AddTrackingPair)
	http.HandleFunc("/deletewatchlocation", ctl.DeleteTrackerPair)

	// startTrackingAllUsers invikes goroutine wich activates all users requests for tracking
	go ctl.startTrackingAllUsers()

	//invoke static files(javascript, css, etc.)
	http.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, r.URL.Path[1:])
	})

	http.ListenAndServe(":"+port, context.ClearHandler(http.DefaultServeMux))

}

//https://www.andjosh.com/2015/01/31/middleware-in-go/
type Controller struct {
	// This will be our extensible type that will
	// be used as a common context type for our routes
	session *mgo.Session // our cloneable session
}

func (ctl *Controller) initCollectionPullScrape(username, location string) error {
	log.Println(username, " Location for scrape: ", location)
	url := startUrl + location + endUrl
	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Fatal(err)
		return err
	}
	var ErrorString = "Cant find such location, try another, or type it correct!"
	if doc.Find("#maincontent ul.listing-results article.panel-listing-result").Text() == "" {
		//return []byte(ErrorString), errors.New(ErrorString)
		return errors.New(ErrorString)
	}

	dbsession := ctl.session.Clone()
	defer dbsession.Close()
	//mapRoomInfo := make(map[string]RoomInfo)
	//mapRoomInfo := make([]RoomInfo, 11)
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
	//RoomInfoColletion := dbsession.DB(DBname).C(usercollection)

	return nil
}

func (ctl *Controller) trackingScraper(username, location string) error {
	//log.Println(username, " Location for scrape: ", location)
	url := startUrl + location + endUrl
	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Fatal(err)
		return err
	}
	var ErrorString = "Cant find such location, try another, or type it correct!"
	if doc.Find("#maincontent ul.listing-results article.panel-listing-result").Text() == "" {
		//return []byte(ErrorString), errors.New(ErrorString)
		return errors.New(ErrorString)
	}

	dbsession := ctl.session.Clone()
	defer dbsession.Close()
	//mapRoomInfo := make(map[string]RoomInfo)
	//mapRoomInfo := make([]RoomInfo, 11)
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

		//log.Println("Type of title trackingScraper:: ", reflect.TypeOf(trackingRoomInfo.Title))

		//log.Println("Titile: ", trackingRoomInfo.Title)
		cpull := dbsession.DB(DBname).C(pullcollectionname)
		result := models.RoomInfo{}
		err = cpull.Find(bson.M{"title": trackingRoomInfo.Title}).One(&result)
		//log.Printf("Result of search in pull: %+v \n", result)
		//checking is in pull collection such Titile
		//log.Printf("Result title in pull: %+v \n", result.Title)
		if result.Title == "" || result.Title == " " {
			ctrack := dbsession.DB(DBname).C(trackcollectionname)
			limit, err23 := ctrack.Count()
			if err23 != nil {
				log.Printf("Count problem in trackingScraper %v\n", err)
				//return err
			}
			log.Println("Limit in ", trackcollectionname, " :", limit)
			if (limit + 5) < 16 {
				//log.Println("Inserting ", trackingRoomInfo, " in ", trackcollectionname)
				err = ctrack.Insert(trackingRoomInfo)
				if err != nil {
					log.Printf("Insert problem to ctrack in trackingScraper %v\n", err)
					//return err
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

				ctrack := dbsession.DB(DBname).C("egorwestminstertrack")
				crackCount, _ := ctrack.Count()
				sendmailresult := make([]models.RoomInfo, crackCount)
				ctrack.Find(bson.M{}).All(&sendmailresult)
				//log.Println(sendmailresult)
				usersInfocol := dbsession.DB(DBname).C("usersInfo")
				userresults := models.UserInfo{}
				usersInfocol.Find((bson.M{"username": username})).One(&userresults)
				//sending email with list of 11 new rooms
				go utils.SendEmailTrackList(sendmailresult, userresults.Email)

				_, err := ctrack.RemoveAll(bson.M{})
				if err != nil {
					log.Printf("RemoveAll of ", pullcollectionname, " in trackingScraper: %v\n", err)
					//return err
				}
			}
		}

	})
	return nil
}
func (ctl *Controller) DeleteTrackerPair(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "sessionRooms")
	if err != nil {
		log.Println(err)
		//http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	log.Println(session.Values)
	//var username string
	username := session.Values["username"].(string)
	location := r.FormValue("location")
	//r.ParseForm()
	// if session.Values["username"] == nil {
	// 	username = r.FormValue("username")
	// } else {
	// 	username = session.Values["username"].(string)
	// }
	session.Save(r, w)
	log.Println("Recievd data from DELETE request: ", username, " ", location)

	dbsession := ctl.session.Clone()
	defer dbsession.Close()
	TrackInfoColletion := dbsession.DB(DBname).C("usersTrackInfo")
	result := models.TrackInfo{}
	err = TrackInfoColletion.Find(bson.M{"username": username}).One(&result)
	log.Printf("Data base trackinfo: %+v \n", result)
	err = TrackInfoColletion.Remove(bson.M{"username": username})

	//, "location": location
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
	log.Println("AddTrackingPair handler used")
	session, err := store.Get(r, "sessionRooms")
	if err != nil {
		log.Println(err)

	}
	log.Println(session.Values)

	// if session.Values["username"] == nil {
	// 	username = r.FormValue("username")
	// } else {
	// 	username = session.Values["username"].(string)
	// }
	session.Save(r, w)
	r.ParseForm()
	username := session.Values["username"].(string)
	location := r.FormValue("location")
	//username := r.FormValue("username")

	log.Println("Recievd data from request: ", username, " ", location)

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

		initerr := ctl.initCollectionPullScrape(username, location)
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

func (ctl *Controller) startTrackingAllUsers() {
	//log.Println("inside startTrackingAllUsers")
	for {
		log.Println("inside startTrackingAllUsers")
		pairs := ctl.makeSliceOfTrackPairs()
		//log.Println("Pairs: ", pairs)
		for _, pair := range pairs {
			ctl.trackingScraper(pair.Username, pair.Location)
		}
		time.Sleep(10 * time.Second)
	}
}

func (ctl *Controller) makeSliceOfTrackPairs() []models.TrackInfo {
	dbsession := ctl.session.Clone()
	defer dbsession.Close()

	TrackInfoColletion := dbsession.DB(DBname).C("usersTrackInfo")
	//result := TrackInfo{}
	num, err := TrackInfoColletion.Count()
	if err != nil {
		log.Println(err)
	}
	//log.Println("Amount of documents: ", num)

	TrackInfoPairs := make([]models.TrackInfo, num)

	err = TrackInfoColletion.Find(bson.M{}).All(&TrackInfoPairs)
	if err != nil {
		log.Println(err)
	}
	//log.Println("Inside makeSliceOfTrackPairs: ", TrackInfoPairs)
	return TrackInfoPairs
}

func NewController() (*Controller, error) {
	// This function will return to us a
	// Controller that has our common DB context.
	// We can then use it for multiple routes
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

func (ctl *Controller) SignUpSubmitHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "sessionRooms")
	if err != nil {
		log.Println(err)
		//http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	r.ParseForm()
	username := r.FormValue("username")
	password := r.FormValue("password")
	email := r.FormValue("email")
	currentUrl := r.FormValue("urlforactivation")
	log.Println(currentUrl)
	if !utils.ValidateEmail(email) {
		w.Write([]byte("Email incorrect "))
		return
	}
	if (username != "") && (password != "") && (email != "") {

		if !ctl.IsUserRegistered(username) {

			session.Values["registered"] = "true"
			session.Values["loggedin"] = "false"
			session.Values["username"] = username
			session.Values["password"] = password
			session.Values["email"] = email
			session.Values["isActivated"] = "false"
			newActivationKey := utils.GenerateKey32chars()
			session.Values["ActivationKey"] = newActivationKey
			session.Save(r, w)

			go utils.SendEmailwithKey(newActivationKey, email, currentUrl)

			dbsession := ctl.session.Clone()
			defer dbsession.Close()

			RoomInfoColletion := dbsession.DB(DBname).C("usersInfo")
			err = RoomInfoColletion.Insert(
				&UserInfo{
					Registred:     session.Values["registered"].(string),
					Loggedin:      session.Values["loggedin"].(string),
					Username:      username, //from request
					Password:      password, //from request
					Email:         session.Values["email"].(string),
					IsActivated:   session.Values["isActivated"].(string),
					ActivationKey: newActivationKey,
				})
			if err != nil {
				log.Println(err)
			}
			w.Write([]byte("Registration successful! Check your email, and activate account"))
			//return
		} else {
			w.Write([]byte("User with such name or email is already exists "))
		}
	} else {
		w.Write([]byte("Some of registration fields are empty!"))
	}
	// log.Println(session.Values["password"], session.Values["username"])
}

func (ctl *Controller) LoginSubmitHandler(w http.ResponseWriter, r *http.Request) {

	session, err := store.Get(r, "sessionRooms")
	if err != nil {
		log.Println(err)
	}

	r.ParseForm()
	username := r.FormValue("username")
	password := r.FormValue("password")
	log.Print("Username login request: ", username, password)

	if ctl.IsUserRegistered(username) {
		log.Print("username type: ", reflect.TypeOf(username))
		if ctl.IsUserActivated(username) {
			log.Println("loginsubmit isactive true")
			dbsession := ctl.session.Clone()
			defer dbsession.Close()
			c := dbsession.DB(DBname).C("usersInfo")
			result := UserInfo{}
			err := c.Find(bson.M{"username": username}).One(&result)
			if err != nil {
				log.Println(err, "some shit")
				w.Write([]byte("Username not found "))
				return
			}
			log.Printf("Data base userinfo: %+v \n", result)
			log.Print("pass from DB: ", result.Password, " pass from cookie: ")
			if result.Password == password {
				log.Println("inside cheking password")
				session.Values["loggedin"] = "true"
				session.Values["username"] = username
				session.Save(r, w)

				colQuerier := bson.M{"username": username}
				change := bson.M{"$set": bson.M{"loggedin": "true"}}
				err = c.Update(colQuerier, change)
				if err != nil {
					panic(err)
				}

				w.Write([]byte("You are logged!"))
				return
			} else {
				w.Write([]byte("Wrong password!"))
				return
			}
		} else {
			//w.Write([]byte("Account not activated"))
			w.Write([]byte("Your account with username: " + username + " is not activated. Check your email: " + session.Values["email"].(string)))
			return

		}
	}

}

func (ctl *Controller) LogoutSubmitHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "sessionRooms")
	if err != nil {
		log.Println(err)

	}
	log.Println("Inside Logout Handler, username for logout: ", session.Values["username"].(string))
	if session.Values["username"].(string) != "" {
		username := session.Values["username"].(string)
		if ctl.IsUserLogged(username) {

			dbsession := ctl.session.Clone()
			defer dbsession.Close()

			c := dbsession.DB(DBname).C("usersInfo")
			session.Values["loggedin"] = "false"
			session.Values["username"] = ""
			session.Save(r, w)
			colQuerier := bson.M{"username": username}
			change := bson.M{"$set": bson.M{"loggedin": "false"}}
			err = c.Update(colQuerier, change)
			if err != nil {
				panic(err)
			}
			w.Write([]byte("Succsesfuly logedout from " + username))
		} else {
			w.Write([]byte("You already have logged out"))
			http.Redirect(w, r, "/login", 302)
			return
		}

	} else {
		w.Write([]byte("You dont have cookie session, please login first"))
	}

}

func (ctl *Controller) ConfirmSignUpHandler(w http.ResponseWriter, r *http.Request) {
	keyInUrl := r.URL.RawQuery
	dbsession := ctl.session.Clone()
	log.Println("Key from email link: ", keyInUrl)
	defer dbsession.Close()
	c := dbsession.DB(DBname).C("usersInfo")
	result := UserInfo{}

	err := c.Find(bson.M{"activationkey": keyInUrl}).One(&result)
	log.Println("Key from database: ", result.ActivationKey)
	if err != nil {
		log.Println(err)
		w.Write([]byte("Wrong activation key "))
		return
	}
	colQuerier := bson.M{"activationkey": keyInUrl}
	change := bson.M{"$set": bson.M{"isactivated": "true"}}
	err = c.Update(colQuerier, change)
	if err != nil {
		panic(err)
	}
	w.Write([]byte("Your account is active now"))
	http.Redirect(w, r, "/login", 302)
	return
}

func (ctl *Controller) IndexHandler(w http.ResponseWriter, r *http.Request) {

	session, err := store.Get(r, "sessionRooms")
	if err != nil {
		log.Println(err)

	}
	log.Println("IndexHandler used")
	//log.Println("Session Values map: ", session.Values)

	var username string
	log.Println("session.Values['username']")
	if session.Values["username"] != nil && session.Values["username"] != "" {
		username = session.Values["username"].(string)
	} else {
		username = ""
	}
	// log.Println("index IsLogged: ", ctl.IsUserLogged(username))
	// log.Println("index IsActive: ", ctl.IsUserActivated(username))
	// log.Println("index IsReged: ", ctl.IsUserRegistered(username))

	//log.Println("indexhandler current Username: ", username)
	//log.Println("Login status from DB: ", ctl.IsUserLogged(username))
	if session.Values["loggedin"] == nil || session.Values["loggedin"] == "false" || !ctl.IsUserLogged(username) {
		session.Save(r, w)
		http.Redirect(w, r, "/login", 302)
		return
	} else {
		t, err := template.ParseFiles("static/index.html")
		if err != nil {
			fmt.Fprintf(w, err.Error())
		}
		t.ExecuteTemplate(w, "index.html", nil)
	}

}

func SignUpHandler(w http.ResponseWriter, r *http.Request) {

	t, err := template.ParseFiles("static/signup.html")
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}
	t.ExecuteTemplate(w, "signup.html", nil)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {

	t, err := template.ParseFiles("static/login.html")
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}
	t.ExecuteTemplate(w, "login.html", nil)

}

//check status ONLY in database!!!
func (ctl *Controller) IsUserRegistered(username string) bool {
	//log.Print("IsRegistred username type: ", reflect.TypeOf(username))
	dbsession := ctl.session.Clone()
	defer dbsession.Close()

	c := dbsession.DB(DBname).C("usersInfo")
	result := UserInfo{}
	err := c.Find(bson.M{"username": username}).One(&result)
	if err != nil {
		log.Println(err, "IsUserRegistered")

		return false
	}
	if result.Registred == "true" {
		return true
	} else {

		return false
	}
}
func (ctl *Controller) IsUserLogged(username string) bool {
	//log.Print("IsLogged username type: ", reflect.TypeOf(username))
	dbsession := ctl.session.Clone()
	defer dbsession.Close()

	c := dbsession.DB(DBname).C("usersInfo")
	result := UserInfo{}
	err := c.Find(bson.M{"username": username}).One(&result)
	if err != nil {
		log.Println(err, "IsUserLogged")
		return false
	}
	if result.Loggedin == "true" {
		return true
	} else {
		return false
	}
}
func (ctl *Controller) IsUserActivated(username string) bool {
	//log.Print("IsActive username type: ", reflect.TypeOf(username))
	dbsession := ctl.session.Clone()
	defer dbsession.Close()

	c := dbsession.DB(DBname).C("usersInfo")
	result := UserInfo{}
	err := c.Find(bson.M{"username": username}).One(&result)
	if err != nil {
		log.Println(err, "IsUserActivated")
		return false
	}
	if result.IsActivated == "true" {
		return true
	} else {
		return false
	}
}
