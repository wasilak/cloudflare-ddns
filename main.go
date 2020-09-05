package main

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/cloudflare/cloudflare-go"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {

	godotenv.Load()

	viper.SetConfigName("config")                 // name of config file (without extension)
	viper.SetConfigType("yaml")                   // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath("/etc/cloudflare-ddns/")  // path to look for the config file in
	viper.AddConfigPath("$HOME/.cloudflare-ddns") // call multiple times to add many search paths
	viper.AddConfigPath(".")                      // optionally look for config in the working directory
	viper.BindEnv("CF.APIKey", "CF_API_KEY")
	viper.BindEnv("CF.APIEmail", "CF_API_EMAIL")
	viperErr := viper.ReadInConfig() // Find and read the config file
	if viperErr != nil {             // Handle errors reading the config file
		log.Fatal(viperErr)
	}

	viper.SetDefault("Record.TTL", 120)
	viper.SetDefault("Record.Type", "A")
	viper.SetDefault("Record.Proxied", false)
	viper.SetDefault("LogFile", "/var/log/cloudflare-dns.log")

	file, err := os.OpenFile(viper.GetString("LogFile"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	mw := io.MultiWriter(os.Stdout, file)
	logrus.SetOutput(mw)
	log.SetFormatter(&log.JSONFormatter{})

	log.SetFormatter(&log.JSONFormatter{})
	if len(viper.GetString("Record.Name")) == 0 {
		log.Fatal("Record name not provided")
	}

	res, err := http.Get("https://api.ipify.org")
	if err != nil {
		log.Fatal(err)
	}

	ip, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	CFAPIKey := viper.GetString("CF.APIKey")
	CFAPIEmail := viper.GetString("CF.APIEmail")

	api, err := cloudflare.New(CFAPIKey, CFAPIEmail)
	if err != nil {
		log.Fatal(err)
	}

	zoneID, err := api.ZoneIDByName(viper.GetString("ZoneName"))
	if err != nil {
		log.Fatal(err)
	}

	vpnRecord := cloudflare.DNSRecord{
		Name: viper.GetString("Record.Name"),
	}

	records, err := api.DNSRecords(zoneID, vpnRecord)
	if err != nil {
		log.Fatal(err)
	}

	if len(records) == 0 {
		vpnRecord.Type = viper.GetString("Record.Type")
		vpnRecord.Proxied = viper.GetBool("Record.Proxied")
		vpnRecord.TTL = viper.GetInt("Record.TTL")
		vpnRecord.Content = string(ip)
		record, err := api.CreateDNSRecord(zoneID, vpnRecord)
		if err != nil {
			log.WithFields(
				log.Fields{
					"Name":       record.Result.Name,
					"Content":    record.Result.Content,
					"Proxiable":  record.Result.Proxiable,
					"TTL":        record.Result.TTL,
					"CreatedOn":  record.Result.CreatedOn,
					"ModifiedOn": record.Result.ModifiedOn,
					"Created":    false,
					"Updated":    false,
				},
			).Fatal(err)
		}
		log.WithFields(
			log.Fields{
				"Name":       record.Result.Name,
				"Content":    record.Result.Content,
				"Proxiable":  record.Result.Proxiable,
				"TTL":        record.Result.TTL,
				"CreatedOn":  record.Result.CreatedOn,
				"ModifiedOn": record.Result.ModifiedOn,
				"Created":    true,
				"Updated":    false,
			},
		).Info("Record created")
	} else {
		for _, record := range records {
			if string(ip) != record.Content {
				record.Content = string(ip)
				err := api.UpdateDNSRecord(zoneID, record.ID, record)
				if err != nil {
					log.WithFields(
						log.Fields{
							"Name":       record.Name,
							"Content":    record.Content,
							"Proxiable":  record.Proxiable,
							"TTL":        record.TTL,
							"CreatedOn":  record.CreatedOn,
							"ModifiedOn": record.ModifiedOn,
							"Created":    false,
							"Updated":    false,
						},
					).Fatal(err)
				}
				log.WithFields(
					log.Fields{
						"Name":       record.Name,
						"Content":    record.Content,
						"Proxiable":  record.Proxiable,
						"TTL":        record.TTL,
						"CreatedOn":  record.CreatedOn,
						"ModifiedOn": record.ModifiedOn,
						"Created":    false,
						"Updated":    true,
					},
				).Info("Record updated")
			} else {
				log.WithFields(
					log.Fields{
						"Name":       record.Name,
						"Content":    record.Content,
						"Proxiable":  record.Proxiable,
						"TTL":        record.TTL,
						"CreatedOn":  record.CreatedOn,
						"ModifiedOn": record.ModifiedOn,
						"Created":    false,
						"Updated":    false,
					},
				).Info("Record not updated")
			}
		}
	}
}
