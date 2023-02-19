package main

import (
	"github.com/wasilak/cloudflare-ddns/cmd"
	// "github.com/wasilak/cloudflare-ddns/library/cf"
)

// func runDNSUpdate(wg *sync.WaitGroup, ip, recordName string, item interface{}) {
// 	proxied := item.(map[string]interface{})["proxied"].(bool)

// 	record := cf.GetDNSRecord(recordName)
// 	record.Type = item.(map[string]interface{})["type"].(string)
// 	record.Proxied = &proxied
// 	record.TTL = item.(map[string]interface{})["ttl"].(int)

// 	zoneName := item.(map[string]interface{})["zonename"].(string)
// 	cf.RunDNSUpdate(string(ip), zoneName, record)
// 	wg.Done()
// }

func main() {

	cmd.Execute()

	// Most API calls require a Context
	// ctx := context.Background()

	// viper.SetDefault("LogFile", "/var/log/cloudflare-dns.log")

	// file, err := os.OpenFile(k.String("LogFile"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	// mw := io.MultiWriter(os.Stdout, file)
	// logrus.SetOutput(mw)
	// log.SetFormatter(&log.JSONFormatter{})

	// log.SetFormatter(&log.JSONFormatter{})

	// res, err := http.Get("https://api.ipify.org")
	// if err != nil {
	// 	log.Fatal(err)
	// }

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

	// ip, err := ioutil.ReadAll(res.Body)
	// if err != nil {
	// 	log.Fatal(err)
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
