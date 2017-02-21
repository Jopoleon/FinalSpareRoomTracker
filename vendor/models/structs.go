// Package models contains structs for using in database.
package models

type UserInfo struct {
	Username      string `json:username bson:username`
	Password      string `json:password bson:password`
	Email         string `json:email bson:email`
	Loggedin      string `json:loggedin bson:loggedin`
	Registred     string `json:registered bson:registered`
	IsActivated   string `json:isActivated bson:isActivated`
	ActivationKey string `json:ActivationKey bson:ActivationKey`
}

type RoomInfo struct {
	Username string `json:Username bson:Username`
	Location string `json:Location bson:Location`
	Title    string `json:Title bson:Title`
	Cost     string `json:Cost bson:Cost`
	ImageUrl string `json:ImageUrl bson:ImageUrl`
}

type TrackInfo struct {
	Username string `json:username bson:username`
	Location string `json:location bson:location`
}
