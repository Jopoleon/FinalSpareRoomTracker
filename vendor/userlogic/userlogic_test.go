package userlogic

import (
	"bytes"
	"log"
	"net/http"
	"testing"

	"gopkg.in/mgo.v2/bson"
	"net/http/httptest"
	"net/url"
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

func TestSignUpSubmitHandler(t *testing.T) {

	ctlTest, err := NewController()
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
	handler := http.HandlerFunc(ctlTest.SignUpSubmitHandler)

	handler.ServeHTTP(rr, req)
	session, err := store.Get(req, "sessionRooms")
	if err != nil {
		log.Println(err)

	}
	session.Values["registered"] = "true"
	session.Values["loggedin"] = "false"
	session.Values["username"] = "testusername"
	session.Values["password"] = "testpass"
	session.Values["email"] = "egortictac@mail.ru"
	session.Values["isActivated"] = "false"
	session.Save(req, rr)

	log.Println(session.Values)

	dbsession := ctlTest.session.Clone()
	defer dbsession.Close()
	c := dbsession.DB(DBname).C("usersInfo")
	result := UserInfo{}
	err = c.Find(bson.M{"username": "testusername"}).One(&result)
	if err != nil {
		t.Fatal(err)
	}

	expectedusername := "testusername"
	expectedregistered := "true"
	if result.Username != expectedusername {
		t.Errorf("SingUpSubmit registred user: got %v want %v",
			result.Username, expectedusername)
	}
	if result.Registred != expectedregistered {
		t.Errorf("SingUpSubmit registration status: got %v want %v",
			result.Registred, expectedregistered)
	}
	ctlTest.session.Close()
}

func TestConfirmSignUpHandler(t *testing.T) {
	testctl, err := NewController()
	if err != nil {
		log.Fatal(err)
	}
	dbsession := testctl.session.Clone()
	defer dbsession.Close()
	c := dbsession.DB(DBname).C("usersInfo")
	result1 := UserInfo{}
	err = c.Find(bson.M{"username": "testusername"}).One(&result1)
	if err != nil {
		t.Fatal(err)
	}

	activKey := result1.ActivationKey
	reqUrl := "/confirm?" + activKey
	req, err := http.NewRequest("POST", reqUrl, nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(testctl.ConfirmSignUpHandler)

	handler.ServeHTTP(rr, req)
	result2 := UserInfo{}
	err = c.Find(bson.M{"username": "testusername"}).One(&result2)
	if err != nil {
		t.Fatal(err)
	}

	expectedactivation := "true"
	if result2.IsActivated != expectedactivation {
		t.Errorf("ConfirmSignUp activation status: got %v want %v",
			result2.IsActivated, expectedactivation)
	}

	testctl.session.Close()

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

	req, err := http.NewRequest("POST", "/loginsubmit", b)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	session, err := store.Get(req, "sessionRooms")
	if err != nil {
		log.Println(err)

	}
	session.Values["registered"] = "true"
	session.Values["loggedin"] = "false"
	session.Values["username"] = "testusername"
	session.Values["password"] = "testpass"
	session.Values["email"] = "egortictac@mail.ru"

	session.Save(req, rr)

	handler := http.HandlerFunc(testctl.LoginSubmitHandler)

	handler.ServeHTTP(rr, req)

	dbsession := testctl.session.Clone()
	defer dbsession.Close()
	c := dbsession.DB(DBname).C("usersInfo")
	result := UserInfo{}
	err = c.Find(bson.M{"username": "testusername"}).One(&result)

	expectedstatus := "true"
	log.Println(rr.Body.String())
	if result.Loggedin != expectedstatus {
		t.Errorf("LoginSubmitHandler registred user: got %v want %v",
			result.Loggedin, expectedstatus)
	}

	testctl.session.Close()

}

func TestLogoutSubmitHandler(t *testing.T) {
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

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(testctl.LogoutSubmitHandler)
	session, err := store.Get(req, "sessionRooms")
	if err != nil {
		log.Println(err)
	}
	session.Values["registered"] = "true"
	session.Values["loggedin"] = "false"
	session.Values["username"] = "testusername"
	session.Values["password"] = "testpass"
	session.Values["email"] = "egortictac@mail.ru"
	session.Values["isActivated"] = "false"
	session.Save(req, rr)

	handler.ServeHTTP(rr, req)
	dbsession := testctl.session.Clone()
	defer dbsession.Close()
	c := dbsession.DB(DBname).C("usersInfo")
	result := UserInfo{}
	err = c.Find(bson.M{"username": "testusername"}).One(&result)

	expectedstatus := "false"
	log.Println(rr.Body.String())
	if result.Loggedin != expectedstatus {
		t.Errorf("LogoutSubmitHandler user status: got %v want %v",
			result.Loggedin, expectedstatus)
	}
	testctl.session.Close()

}

func TestIsUserLogged(t *testing.T) {
	testctl, err := NewController()
	if err != nil {
		log.Fatal(err)
	}
	result := testctl.IsUserLogged("testusername")
	expected := false

	if result != expected {
		t.Errorf("IsUserLogged user status: got %v want %v",
			result, expected)
	}

	testctl.session.Close()
}

func TestIsUserActivated(t *testing.T) {
	testctl, err := NewController()
	if err != nil {
		log.Fatal(err)
	}
	result := testctl.IsUserActivated("testusername")
	expected := true

	if result != expected {
		t.Errorf("IsUserLogged user status: got %v want %v",
			result, expected)
	}

	testctl.session.Close()

}

func TestIndexHandler(t *testing.T) {

	testctl, err := NewController()
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest("POST", "/", nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(testctl.IndexHandler)
	session, err := store.Get(req, "sessionRooms")
	if err != nil {
		log.Println(err)
	}
	session.Values["registered"] = "true"
	session.Values["loggedin"] = "false"

	session.Save(req, rr)

	handler.ServeHTTP(rr, req)

	expectedredirect := "/login"
	result := rr.HeaderMap["Location"]

	if result[0] != expectedredirect {
		t.Errorf("IndexHandler redirect: got %v want %v",
			result[0], expectedredirect)
	}

}

func TestIsUserRegistered(t *testing.T) {
	testctl, err := NewController()
	if err != nil {
		log.Fatal(err)
	}
	result := testctl.IsUserRegistered("testusername")
	expected := true

	if result != expected {
		t.Errorf("IsUserLogged user status: got %v want %v",
			result, expected)
	}
	cleanUpTestUser(testctl)

	testctl.session.Close()

}

func cleanUpTestUser(ctl *Controller) {
	c := ctl.session.DB(DBname).C("usersInfo")
	err := c.Remove(bson.M{"username": "testusername"})
	if err != nil {
		log.Fatal(err)
	}
	ctl.session.Close()
}
