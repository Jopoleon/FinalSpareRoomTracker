package userlogic

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	//"os"
	"reflect"
	//"strconv"
	//"errors"
	//"time"
	//"crypto/sha1"
	//"io/ioutil"
	//"encoding/json"

	"models"
	//"scrape"
	"utils"

	//"github.com/PuerkitoBio/goquery"
	//"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var store = sessions.NewCookieStore([]byte("nRrHLlHcHH0u7fUz25Hje9m7uJ5SnJzP"))

//"mongodb://egor2:qwer1234@ds153729.mlab.com:53729/spareroom"
var mongoUrl = "localhost"

//"mongodb://egor2:qwer1234@ds153729.mlab.com:53729/spareroom"

var startUrl = "http://www.spareroom.co.uk/flatshare/search.pl?flatshare_type=offered&location_type=area&search="
var endUrl = "&miles_from_max=1&action=search&templateoveride=&show_results=&submit="

var DBname = "spareroom"

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

func (ctl *Controller) SignUpSubmitHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "sessionRooms")
	if err != nil {
		log.Println(err)
		//http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	err = r.ParseForm()
	if err != nil {
		log.Println("r.ParseForm() ", err)
	}
	username := r.FormValue("username")
	password := r.FormValue("password")
	email := r.FormValue("email")
	currentUrl := r.FormValue("urlforactivation")
	log.Println(currentUrl)
	if !utils.ValidateEmail(email) {
		w.Write([]byte("Email incorrect "))
		return
	}
	if username != "" && password != "" && email != "" {

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
				&models.UserInfo{
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
			result := models.UserInfo{}
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
			//http.Redirect(w, r, "/", 302)
		} else {
			w.Write([]byte("You already have logged out"))
			//http.Redirect(w, r, "/", 302)
			return
		}

	} else {
		w.Write([]byte("You dont have cookie session, please login first"))
		//http.Redirect(w, r, "/", 302)
	}

}

func (ctl *Controller) ConfirmSignUpHandler(w http.ResponseWriter, r *http.Request) {
	keyInUrl := r.URL.RawQuery
	dbsession := ctl.session.Clone()
	log.Println("Key from email link: ", keyInUrl)
	defer dbsession.Close()
	c := dbsession.DB(DBname).C("usersInfo")
	result := models.UserInfo{}

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

	var username string
	//log.Println("session.Values['username']")
	if session.Values["username"] != nil && session.Values["username"] != "" {
		username = session.Values["username"].(string)
	} else {
		username = ""
	}
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

func (ctl *Controller) IsUserRegistered(username string) bool {
	//log.Print("IsRegistred username type: ", reflect.TypeOf(username))
	dbsession := ctl.session.Clone()
	defer dbsession.Close()

	c := dbsession.DB(DBname).C("usersInfo")
	result := models.UserInfo{}
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
	result := models.UserInfo{}
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
	result := models.UserInfo{}
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
