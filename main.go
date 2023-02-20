package main

import (
	"github.com/wasilak/cloudflare-ddns/cmd"
	// "github.com/wasilak/cloudflare-ddns/libs"
)

func main() {

	cmd.Execute()

	// if k.Bool("mail.enabled") {

	// 	currentTime := time.Now()
	// 	layout := "2006-01-02"

	// 	mail := notif.Mail{
	// 		From:    k.String("mail.from"),
	// 		To:      k.StringSlice("mail.to"),
	// 		Subject: fmt.Sprintf("%s %s", k.String("mail.subject"), currentTime.Format(layout)),
	// 		SMTP:    k.String("mail.smtp"),
	// 		Auth: notif.MailAuth{
	// 			Username: k.String("mail.auth.username"),
	// 			Password: k.String("mail.auth.password"),
	// 		},
	// 		Data: data[0],
	// 	}

	// 	_, err := mail.Send(false)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}

	// 	log.Printf("Email sent to: %v", k.StringSlice("mail.to"))
	// }

	// var wg sync.WaitGroup

	// cf.Init(viper.GetString("CF.APIKey"), viper.GetString("CF.APIEmail"), ctx)

	// records := viper.GetStringMap("records")

	// for recordName, item := range records {
	// 	wg.Add(1)
	// 	go runDNSUpdate(&wg, string(ip), recordName, item)
	// }

	// wg.Wait()
}
