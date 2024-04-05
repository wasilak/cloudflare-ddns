package libs

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"

	"github.com/cloudflare/cloudflare-go"
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

func PrepareRecordsFromConfig() []cloudflare.DNSRecord {
	var records []cloudflare.DNSRecord
	viper.UnmarshalKey("records", &records)
	return records
}

// The Runner function updates DNS records for a given IP address using Cloudflare API.
func Runner(ctx context.Context, records []cloudflare.DNSRecord) (string, error) {
	var wg sync.WaitGroup

	ip, err := GetIP()
	if err != nil {
		return "", err
	}

	cfAPI := cf.CF{}

	cfAPI.Init(viper.GetString("CF.APIKey"), viper.GetString("CF.APIEmail"), ctx)

	for _, record := range records {
		wg.Add(1)
		record.Content = ip
		go runDNSUpdate(&wg, &cfAPI, record)
	}

	wg.Wait()

	return ip, nil
}

// This function updates a DNS record with a given IP address and record name using the Cloudflare API.
func runDNSUpdate(wg *sync.WaitGroup, cfAPI *cf.CF, record cloudflare.DNSRecord) {
	cfAPI.RunDNSUpdate(record)
	wg.Done()
}
