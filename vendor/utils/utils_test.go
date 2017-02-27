package utils

import (
	"net/smtp"
	"testing"
)

func TestGenerateKey32chars(t *testing.T) {
	expectedlen := 32
	result := len(GenerateKey32chars())
	if result != expectedlen {
		t.Fatalf("Expected length of key %s, got %s", expectedlen, result)
	}
}

func TestValidateEmail(t *testing.T) {
	expcetedtrue := true
	expectedfalse := false
	goodemailadr := "test@test.com"
	bademailadr := "dassda.com@asd"
	resultgood := ValidateEmail(goodemailadr)
	resultbad := ValidateEmail(bademailadr)
	if resultbad != expectedfalse {
		t.Fatalf("Expected validation %s, got %s", expectedfalse, resultbad)
	}
	if resultgood != expcetedtrue {
		t.Fatalf("Expected validation %s, got %s", expcetedtrue, resultgood)
	}
}

type EmailConfig struct {
	Username   string
	Password   string
	ServerHost string
	ServerPort string
	SenderAddr string
}

type EmailSender interface {
	Send(to []string, body []byte) error
}

func NewEmailSender(conf EmailConfig) EmailSender {
	return &emailSender{conf, smtp.SendMail}
}

type emailSender struct {
	conf EmailConfig
	send func(string, smtp.Auth, string, []string, []byte) error
}

func (e *emailSender) Send(to []string, body []byte) error {
	addr := e.conf.ServerHost + ":" + e.conf.ServerPort
	auth := smtp.PlainAuth("", e.conf.Username, e.conf.Password, e.conf.ServerHost)
	return e.send(addr, auth, e.conf.SenderAddr, to, body)
}
func TestSendEmailwithKey(t *testing.T) {
	f, r := mockSend(nil)
	sender := &emailSender{send: f}
	body := "Hello World"
	err := sender.Send([]string{"me@example.com"}, []byte(body))

	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if string(r.msg) != body {
		t.Errorf("wrong message body.\n\nexpected: %\n got: %s", body, r.msg)
	}
}

func mockSend(errToReturn error) (func(string, smtp.Auth, string, []string, []byte) error, *emailRecorder) {
	r := new(emailRecorder)
	return func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		*r = emailRecorder{addr, a, from, to, msg}
		return errToReturn
	}, r
}

type emailRecorder struct {
	addr string
	auth smtp.Auth
	from string
	to   []string
	msg  []byte
}
