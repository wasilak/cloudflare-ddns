package notif

import (
	"bytes"
	"html/template"
	"log"
	"net/smtp"

	diff "github.com/r3labs/diff/v2"
	gomail "gopkg.in/mail.v2"
)

var auth smtp.Auth

type MPOResponse struct {
	ID        int            `json:"-" diff:"-"`
	Changelog diff.Changelog `json:"changelog" diff:"-"`
	Idzad     int            `json:"idzad" diff:"-"`
	Typn      string         `json:"typn" diff:"-"`
	Uidmgo    string         `json:"uidmgo" diff:"-"`
	X         int            `json:"x" diff:"-"`
	Y         int            `json:"y" diff:"-"`
	Address   string         `json:"adres"`
	District  string         `json:"dzielnica"`
	Note      string         `json:"notatka"`
	Date      string         `json:"date" diff:"-"`
}

type Mail struct {
	From    string
	To      []string
	Subject string
	Body    string
	Auth    MailAuth
	SMTP    string
	Data    MPOResponse
}

type MailAuth struct {
	Username string
	Password string
}

func (m *Mail) Send(dryRun bool) (bool, error) {

	err := m.parseTemplate("template.html", m.Data)
	if err != nil {
		log.Fatal(err)
	}

	gm := gomail.NewMessage()
	gm.SetHeader("From", m.From)
	gm.SetHeader("To", m.To...)
	gm.SetHeader("Subject", m.Subject)
	gm.SetBody("text/html", m.Body)

	d := gomail.NewDialer("smtp.gmail.com", 587, m.Auth.Username, m.Auth.Password)

	if err := d.DialAndSend(gm); err != nil {
		panic(err)
	}

	return true, nil
}

func (m *Mail) parseTemplate(templateFileName string, data interface{}) error {
	t, err := template.ParseFiles(templateFileName)
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	if err = t.Execute(buf, data); err != nil {
		return err
	}
	m.Body = buf.String()

	return nil
}
