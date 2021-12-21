package cf

import (
	"context"
	log "github.com/sirupsen/logrus"

	"github.com/cloudflare/cloudflare-go"
)

var api *cloudflare.API
var err error
var ctx context.Context

// Init func
func Init(CFAPIKey, CFAPIEmail string, CFctx context.Context) {
	ctx = CFctx
	api, err = cloudflare.New(CFAPIKey, CFAPIEmail)
	if err != nil {
		log.Fatal(err)
	}
}

// GetZoneID func
func GetZoneID(zoneName string) string {
	zoneID, err := api.ZoneIDByName(zoneName)
	if err != nil {
		log.Fatal(err)
	}

	return zoneID
}

// GetDNSRecord func
func GetDNSRecord(recordName string) cloudflare.DNSRecord {
	record := cloudflare.DNSRecord{
		Name: recordName,
	}

	return record
}

// GetDNSRecords func
func GetDNSRecords(zoneID string, record cloudflare.DNSRecord) []cloudflare.DNSRecord {
	records, err := api.DNSRecords(ctx, zoneID, record)
	if err != nil {
		log.Fatal(err)
	}

	return records
}

func createDNSRecord(zoneID string, record cloudflare.DNSRecord) {
	record.Proxiable = true
	newRecord, err := api.CreateDNSRecord(ctx, zoneID, record)

	logFields := log.Fields{
		"Name":       newRecord.Result.Name,
		"Content":    newRecord.Result.Content,
		"Proxied":    newRecord.Result.Proxied,
		"TTL":        newRecord.Result.TTL,
		"CreatedOn":  newRecord.Result.CreatedOn,
		"ModifiedOn": newRecord.Result.ModifiedOn,
		"Updated":    false,
		"Created":    true,
	}

	if err != nil {
		log.WithFields(logFields).Fatal(err)
	}

	log.WithFields(logFields).Info("Record created")
}

func updateDNSRecord(zoneID, recordID string, record cloudflare.DNSRecord, updated bool) {

	logFields := log.Fields{
		"Name":       record.Name,
		"Content":    record.Content,
		"Proxied":    record.Proxied,
		"TTL":        record.TTL,
		"CreatedOn":  record.CreatedOn,
		"ModifiedOn": record.ModifiedOn,
		"Updated":    false,
		"Created":    false,
	}

	err := api.UpdateDNSRecord(ctx, zoneID, record.ID, record)
	if err != nil {
		log.WithFields(logFields).Fatal(err)
	}

	logFields["Updated"] = updated
	log.WithFields(logFields).Info("Record updated")
}

// RunDNSUpdate func
func RunDNSUpdate(ip, zoneName string, record cloudflare.DNSRecord) {
	zoneID := GetZoneID(zoneName)

	records := GetDNSRecords(zoneID, record)

	record.Content = string(ip)

	if len(records) == 0 {
		createDNSRecord(zoneID, record)
	} else {
		for _, item := range records {

			// update has to be done using type from GetDNSRecords() response
			// not with custom built record Type for some reason,
			// hence overrides below...
			updated := false
			if item.Content != record.Content {
				updated = true
				item.Content = record.Content
			}
			if item.Proxied != record.Proxied {
				updated = true
				item.Proxied = record.Proxied
			}
			if item.Type != record.Type {
				updated = true
				item.Type = record.Type
			}
			if item.TTL != record.TTL {
				updated = true
				item.TTL = record.TTL
			}
			updateDNSRecord(zoneID, item.ID, item, updated)
		}
	}
}
