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

// The MailData type contains a single field for storing an IP address as a string.
// @property {string} IP - The "IP" property is a string that represents an IP address. It is likely
// used to store the IP address of an email sender or recipient in an email application or system.
type MailData struct {
	IP string
}

// The "Mail" type represents an email message with sender, recipient(s), subject, body,
// authentication, SMTP server, and IP data.
// @property {string} From - The email address of the sender.
// @property {[]string} To - To is a slice of strings that represents the email addresses of the
// recipients of the email.
// @property {string} Subject - The subject of the email that will be sent.
// @property {string} Body - The "Body" property of the "Mail" struct represents the main content of
// the email message. It typically contains the message text, but can also include HTML markup, images,
// and attachments.
// @property {MailAuth} Auth - Auth is a struct that contains the authentication information for the
// email. It may include fields such as username and password for the SMTP server.
// @property {string} SMTP - SMTP stands for Simple Mail Transfer Protocol. It is a protocol used for
// sending and receiving email messages over the internet. In the context of the Mail struct, the SMTP
// property specifies the address of the SMTP server that will be used to send the email.
// @property {MailData} IP - The "IP" property in the "Mail" struct is of type "MailData". It is likely
// used to store information related to the IP address of the mail server being used to send the email.
// This could include details such as the server's hostname, port number, and any other relevant
// connection
type Mail struct {
	From    string
	To      []string
	Subject string
	Body    string
	Auth    MailAuth
	SMTP    string
	IP      MailData
}

// The above code defines a struct type called MailAuth with two string fields, Username and Password.
// @property {string} Username - The `Username` property is a string that represents the username or
// email address used for authentication when sending or receiving emails.
// @property {string} Password - The `Password` property is a string type field that represents the
// password for a user's email account. It is typically used in conjunction with the `Username`
// property to authenticate and authorize access to the email account.
type MailAuth struct {
	Username string
	Password string
}

// The `Send` function is a method of the `Mail` struct that sends an email message using the specified
// SMTP server and authentication credentials. It takes a boolean parameter `dryRun` which is not used
// in the function and returns a boolean value and an error. The function first calls the
// `parseTemplate` method to parse an HTML template file and populate the `Body` field of the `Mail`
// struct with the parsed content. It then creates a new `gomail.Message` object and sets the email
// headers and body using the values from the `Mail` struct. Finally, it creates a new `gomail.Dialer`
// object with the specified SMTP server and authentication credentials, and uses it to send the email
// message. If there is an error during the sending process, the function panics. The function returns
// a boolean value `true` and a `nil` error if the email is sent successfully.
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

// The `parseTemplate` function is a method of the `Mail` struct that takes a template file name and a
// data interface as input parameters. It uses the `template.Must` function to parse the HTML template
// file from the embedded file system and populate the `Body` field of the `Mail` struct with the
// parsed content. It then returns an error if there is any issue during the parsing process.
func (m *Mail) parseTemplate(templateFileName string, data interface{}) error {
	t := template.Must(template.ParseFS(templateFiles, "templates/*"))

	buf := new(bytes.Buffer)
	if err := t.Execute(buf, data); err != nil {
		return err
	}
	m.Body = buf.String()

	return nil
}

// The Notify function sends an email notification with the given IP address if email notifications are
// enabled in the configuration.
func Notify(ctx context.Context, ip string) error {
	if viper.GetBool("mail.enabled") {

		logger := ctx.Value("logger").(*slog.Logger)
		logger.Debug(fmt.Sprintf("%+v", viper.AllSettings()))

		mailData := MailData{
			IP: ip,
		}

		// The code is creating a new `Mail` struct and initializing its fields with values obtained from the
		// configuration file using the `viper` package. The `From` field is set to the value of the
		// `mail.from` configuration key, the `To` field is set to the value of the `mail.to` configuration
		// key as a slice of strings, the `Subject` field is set to the value of the `mail.subject`
		// configuration key, the `SMTP` field is set to the value of the `mail.smtp` configuration key, the
		// `Auth` field is set to a new `MailAuth` struct with its `Username` and `Password` fields set to
		// the values of the `mail.auth.username` and `mail.auth.password` configuration keys respectively,
		// and the `IP` field is set to a `MailData` struct with its `IP` field set to the `ip` parameter
		// passed to the `Notify` function.
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

		logger.Debug("Email sent to", viper.GetStringSlice("mail.to"))
	}

	return nil
}
