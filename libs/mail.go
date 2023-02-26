package libs

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"
	"log"

	"github.com/spf13/viper"
	"golang.org/x/exp/slog"
	gomail "gopkg.in/mail.v2"
)

//go:embed templates
var templateFiles embed.FS

type MailData struct {
	IP string
}

type Mail struct {
	From    string
	To      []string
	Subject string
	Body    string
	Auth    MailAuth
	SMTP    string
	IP      MailData
}

type MailAuth struct {
	Username string
	Password string
}

func (m *Mail) Send(dryRun bool) (bool, error) {

	err := m.parseTemplate("template.html", m.IP)
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
	t := template.Must(template.ParseFS(templateFiles, "templates/*"))

	buf := new(bytes.Buffer)
	if err := t.Execute(buf, data); err != nil {
		return err
	}
	m.Body = buf.String()

	return nil
}

func Notify(ctx context.Context, ip string) error {
	if viper.GetBool("mail.enabled") {

		logger := ctx.Value("logger").(*slog.Logger)
		logger.Debug(fmt.Sprintf("%+v", viper.AllSettings()))

		mailData := MailData{
			IP: ip,
		}

		mail := Mail{
			From:    viper.GetString("mail.from"),
			To:      viper.GetStringSlice("mail.to"),
			Subject: viper.GetString("mail.subject"),
			SMTP:    viper.GetString("mail.smtp"),
			Auth: MailAuth{
				Username: viper.GetString("mail.auth.username"),
				Password: viper.GetString("mail.auth.password"),
			},
			IP: mailData,
		}

		_, err := mail.Send(false)
		if err != nil {
			return err
		}

		logger.Debug("Email sent to:", viper.GetStringSlice("mail.to"))
	}

	return nil
}
