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

// The Runner function updates DNS records for a given IP address using Cloudflare API.
func Runner(ctx context.Context) (string, error) {
	var wg sync.WaitGroup

	ip, err := GetIP()
	if err != nil {
		return "", err
	}

	cf.Init(viper.GetString("CF.APIKey"), viper.GetString("CF.APIEmail"), ctx)

	records := viper.GetStringMap("records")

	for recordName, item := range records {
		wg.Add(1)
		go runDNSUpdate(&wg, ip, recordName, item)
	}

	wg.Wait()

	return ip, nil
}

// This function updates a DNS record with a given IP address and record name using the Cloudflare API.
func runDNSUpdate(wg *sync.WaitGroup, ip, recordName string, item interface{}) {
	proxied := item.(map[string]interface{})["proxied"].(bool)

	record := cloudflare.DNSRecord{
		Name:      recordName,
		Type:      item.(map[string]interface{})["type"].(string),
		TTL:       item.(map[string]interface{})["ttl"].(int),
		Proxiable: true,
		Proxied:   &proxied,
		Content:   ip,
		ZoneName:  item.(map[string]interface{})["zonename"].(string),
	}

	cf.RunDNSUpdate(record)
	wg.Done()
}
