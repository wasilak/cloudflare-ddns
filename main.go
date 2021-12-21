package main

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sync"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/wasilak/cloudflare-ddns/library/cf"
)

func runDNSUpdate(wg *sync.WaitGroup, ip, recordName string, item interface{}) {
	proxied := item.(map[string]interface{})["proxied"].(bool)

	record := cf.GetDNSRecord(recordName)
	record.Type = item.(map[string]interface{})["type"].(string)
	record.Proxied = &proxied
	record.TTL = item.(map[string]interface{})["ttl"].(int)

	zoneName := item.(map[string]interface{})["zonename"].(string)
	cf.RunDNSUpdate(string(ip), zoneName, record)
	wg.Done()
}

func main() {

	// Most API calls require a Context
	ctx := context.Background()

	godotenv.Load()

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.cloudflare-ddns")
	viper.AddConfigPath("/etc/cloudflare-ddns/")
	viper.BindEnv("CF.APIKey", "CF_API_KEY")
	viper.BindEnv("CF.APIEmail", "CF_API_EMAIL")
	viperErr := viper.ReadInConfig()
	if viperErr != nil {
		log.Fatal(viperErr)
	}

	viper.SetDefault("LogFile", "/var/log/cloudflare-dns.log")

	file, err := os.OpenFile(viper.GetString("LogFile"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	mw := io.MultiWriter(os.Stdout, file)
	logrus.SetOutput(mw)
	log.SetFormatter(&log.JSONFormatter{})

	log.SetFormatter(&log.JSONFormatter{})

	res, err := http.Get("https://api.ipify.org")
	if err != nil {
		log.Fatal(err)
	}

	ip, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup

	cf.Init(viper.GetString("CF.APIKey"), viper.GetString("CF.APIEmail"), ctx)

	records := viper.GetStringMap("records")

	for recordName, item := range records {
		wg.Add(1)
		go runDNSUpdate(&wg, string(ip), recordName, item)
	}

	wg.Wait()
}
