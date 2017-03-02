package main

import (
	
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"scrape"
	"trackinglogic"
	"userlogic"
	

	
	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"gopkg.in/mgo.v2"
	
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
var mongoUrl = "localhost"

var startUrl = "http://www.spareroom.co.uk/flatshare/search.pl?flatshare_type=offered&location_type=area&search="
var endUrl = "&miles_from_max=1&action=search&templateoveride=&show_results=&submit="

var DBname = "spareroom"

//sendMailLimitThreshold is how many new rooms tracker saves before sending email with list of this rooms
var sendMailLimitThreshold = 20

//trackerCycleTimeStep is how many SECONDS tracker sleeps befor make another cycle of cheking new rooms
var trackerCycleTimeStep = (10 * time.Second)

func main() {
	

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	userCtl, err := userlogic.NewController()
	if err != nil {
		log.Fatal(err)
	}

	trackerCtl, err := trackinglogic.NewController()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Server started on port: ", port)

	//Homepage
	http.HandleFunc("/", userCtl.IndexHandler)
	//Scrape for rooms
	http.HandleFunc("/scrapelocation", scrape.ScraperHandler)
	//trial scrape for unregistered users
	http.HandleFunc("/trialscrapelocation", scrape.TrialScraperHandler)
	//sign up in system
	http.HandleFunc("/signup", SignUpHandler)
	//submit signup information
	http.HandleFunc("/signupsubmit", userCtl.SignUpSubmitHandler)

	http.HandleFunc("/confirm", userCtl.ConfirmSignUpHandler)

	http.HandleFunc("/login", LoginHandler)
	http.HandleFunc("/loginsubmit", userCtl.LoginSubmitHandler)

	http.HandleFunc("/logout", userCtl.LogoutSubmitHandler)

	http.HandleFunc("/watchlocation", trackerCtl.AddTrackingPair)
	http.HandleFunc("/deletewatchlocation", trackerCtl.DeleteTrackerPair)

	// startTrackingAllUsers invikes goroutine wich activates all users requests for tracking
	go trackerCtl.StartTrackingAllUsers()

	//invoke static files(javascript, css, etc.)
	http.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, r.URL.Path[1:])
	})

	http.ListenAndServe(":"+port, context.ClearHandler(http.DefaultServeMux))

}

//https://www.andjosh.com/2015/01/31/middleware-in-go/
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
