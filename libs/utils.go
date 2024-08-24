package libs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"sync"

	"github.com/spf13/viper"
	"github.com/wasilak/cloudflare-ddns/libs/cf"
)

// The function retrieves the public IP address of the device it is running on.
func GetIP() (string, error) {
	res, err := http.Get("https://api.ipify.org")

	if err != nil {
		return "", err
	}

	ip, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	ipAddr := string(ip)

	if net.ParseIP(ipAddr) != nil {
		return ipAddr, nil
	}

	return "", fmt.Errorf("%s is not an IP address", ipAddr)
}

func PrepareRecords() []cf.ExtendedCloudflareDNSRecord {
	_, present := os.LookupEnv(viper.GetEnvPrefix() + "_RECORDS")

	if present {
		return prepareRecordsFromEnv()
	}

	return prepareRecordsFromConfig()
}

func prepareRecordsFromEnv() []cf.ExtendedCloudflareDNSRecord {
	var records []cf.ExtendedCloudflareDNSRecord

	byt := []byte(viper.GetString("records"))

	if err := json.Unmarshal(byt, &records); err != nil {
		panic(err)
	}

	return records
}

func prepareRecordsFromConfig() []cf.ExtendedCloudflareDNSRecord {
	var records []cf.ExtendedCloudflareDNSRecord
	viper.UnmarshalKey("records", &records)
	return records
}

// The Runner function updates DNS records for a given IP address using Cloudflare API.
func Runner(ctx context.Context, ip string, records []cf.ExtendedCloudflareDNSRecord, deleteRecords bool) error {
	var wg sync.WaitGroup

	cfAPI := cf.CF{}

	cfAPI.Init(viper.GetString("CF.APIKey"), viper.GetString("CF.APIEmail"), ctx)

	for _, record := range records {
		wg.Add(1)

		if record.Record.Type == "CNAME" {
			record.Record.Content = record.CNAME
		} else {
			record.Record.Content = ip
		}

		go runDNSUpdate(&wg, &cfAPI, record, deleteRecords)
	}

	wg.Wait()

	return nil
}

// This function updates a DNS record with a given IP address and record name using the Cloudflare API.
func runDNSUpdate(wg *sync.WaitGroup, cfAPI *cf.CF, record cf.ExtendedCloudflareDNSRecord, deleteRecords bool) {
	cfAPI.RunDNSUpdate(record, deleteRecords)
	wg.Done()
}
