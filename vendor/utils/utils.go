// Package utils implements some utility fucntions like sending email, generating ID etc.

package utils

import (
	"crypto/rand"
	"fmt"
	"net/smtp"
	"log"
	"models"
	"regexp"
	"strconv"
	"time"
)

var url_start = "http://www.spareroom.co.uk"
var searchEndpoint = "/flatshare/search.pl?"
var full_request_url = "http://www.spareroom.co.uk/flatshare/search.pl?flatshare_type={offered|wanted|buddyup}&location_type=area&search={searchterm}&miles_from_max={miles}&showme_rooms=Y&showme_1beds=Y&showme_buddyup_properties=Y&min_rent={mincost}&max_rent={maxcost}&per=pcm&no_of_rooms=&min_term=0&max_term=0&available_search=N&day_avail=&mon_avail=&year_avail=&min_age_req=&max_age_req=&min_beds=&max_beds=&keyword=&searchtype=advanced%20&editing=&mode=&nmsq_mode=&action=search&templateoveride=&show_results=&submit="
var url_end = "&action=search&templateoveride=&show_results=&submit="

var firtsTryUrlsReq = "http://www.spareroom.co.uk/flatshare/search.pl?flatshare_type=offered&location_type=area&search=westminster&miles_from_max=1&action=search&templateoveride=&show_results=&submit="
var loca = "belgravia"

var startUrl = "http://www.spareroom.co.uk/flatshare/search.pl?flatshare_type=offered&location_type=area&search="
var endUrl = "&miles_from_max=1&action=search&templateoveride=&show_results=&submit="

func GenerateKey32chars() string {
	buf := make([]byte, 16)
	_, err := rand.Read(buf)
	if err != nil {
		panic(err) 
	}
	return fmt.Sprintf("%x", buf)
	
}

func ValidateEmail(email string) bool {
	Re := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return Re.MatchString(email)
}

func SendEmailwithKey(key, addres, currentActivationUrl string) {
	fmt.Println("Start time: ", time.Now())
	
	type EmailUser struct {
		Username    string
		Password    string
		EmailServer string
		Port        int
	}

	emailUser := &EmailUser{"spareroommailserver", "Sendkey2017", "smtp.gmail.com", 587}

	auth := smtp.PlainAuth("",
		emailUser.Username,
		emailUser.Password,
		emailUser.EmailServer,
	)

	var err error

	link := currentActivationUrl + "?" + key
	
	msg := []byte("To: " + addres + "\r\n" +
		"Subject: Activation letter from SpareRoomScraper\r\n" +
		"\r\n" +
		"This is your activation link: \r\n" + link)
	err = smtp.SendMail(emailUser.EmailServer+":"+strconv.Itoa(emailUser.Port),
		auth,
		emailUser.Username,
		[]string{addres},
		msg)
	if err != nil {
		log.Print("ERROR: attempting to send a mail ", err)
	}
	fmt.Println("End time: ", time.Now())
}


func SendEmailTrackList(roominfoslice []models.RoomInfo, emailadr, user, loca string) {

	type EmailUser struct {
		Username    string
		Password    string
		EmailServer string
		Port        int
	}

	emailUser := &EmailUser{"spareroommailserver", "Sendkey2017", "smtp.gmail.com", 587}

	auth := smtp.PlainAuth("",
		emailUser.Username,
		emailUser.Password,
		emailUser.EmailServer,
	)

	var err error
	var msgbody string
	for _, info := range roominfoslice {
		msgbody += " " + info.Title + " " + info.Cost + " " + "https:" + info.ImageUrl + "\n\n"
	}

	msg := []byte("To: " + emailadr + "\r\n" +
		"Subject: New rooms in your tracking area: " + loca + "\r\n" +
		"\r\n" + " Hello " + user + "\n Here are some new rooms for you: \n \n" +
		msgbody)
	err = smtp.SendMail(emailUser.EmailServer+":"+strconv.Itoa(emailUser.Port),
		auth,
		emailUser.Username,
		[]string{emailadr},
		msg)
	if err != nil {
		log.Print("ERROR: attempting to send a mail ", err)
	}
	
}
