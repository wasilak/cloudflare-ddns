package libs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"sync"

	"github.com/spf13/viper"
	"github.com/wasilak/cloudflare-ddns/libs/api"
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
func Runner(ctx context.Context, records []cf.ExtendedCloudflareDNSRecord) error {
	var wg sync.WaitGroup

	for _, record := range records {
		wg.Add(1)

		if record.Record.Type == "CNAME" {
			if record.CNAME == "" {
				slog.With("record", record).ErrorContext(ctx, "This is CNAME record but CNAME value is empty")
				continue
			}
			record.Record.Content = record.CNAME
		} else {
			record.Record.Content = api.CurrentIp
		}

		go runDNSUpdate(&wg, ctx, record)
	}

	wg.Wait()

	return nil
}

// This function updates a DNS record with a given IP address and record name using the Cloudflare API.
func runDNSUpdate(wg *sync.WaitGroup, ctx context.Context, record cf.ExtendedCloudflareDNSRecord) {
	err := api.RunDNSUpdate(ctx, record)
	if err != nil {
		slog.With("record", record).ErrorContext(ctx, "RunDNSUpdate Error", "error", err)
	}
	wg.Done()
}

func GetAppName() string {
	appName := os.Getenv("OTEL_SERVICE_NAME")
	if appName == "" {
		appName = os.Getenv("APP_NAME")
		if appName == "" {
			appName = "cloudflare-ddns"
		}
	}
	return appName
}
