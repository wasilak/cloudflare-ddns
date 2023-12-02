package cf

import (
	"context"
	"os"

	"log/slog"

	cloudflare "github.com/cloudflare/cloudflare-go"
)

type CF struct {
	API *cloudflare.API
	CTX context.Context
}

// The function initializes a Cloudflare API client with the provided API key, email, and context.
func (cf *CF) Init(CFAPIKey, CFAPIEmail string, ctx context.Context) {
	cf.CTX = ctx
	var err error

	cf.API, err = cloudflare.New(CFAPIKey, CFAPIEmail)
	if err != nil {
		slog.ErrorContext(ctx, err.Error(), err)
		os.Exit(1)
	}
}

// The function takes a zone name as input and returns its corresponding zone ID using an API call.
func (cf *CF) GetZoneID(zoneName string) string {
	zoneID, err := cf.API.ZoneIDByName(zoneName)
	if err != nil {
		slog.ErrorContext(cf.CTX, err.Error(), err)
	}

	return zoneID
}

// This function retrieves a DNS record from Cloudflare using its ID.
func (cf *CF) GetDNSRecord(rc *cloudflare.ResourceContainer, record cloudflare.DNSRecord) (cloudflare.DNSRecord, error) {

	record, err := cf.API.GetDNSRecord(cf.CTX, rc, record.ID)
	if err != nil {
		return record, err
	}

	return record, nil
}

// The function creates a DNS record and logs its details.
func (cf *CF) createDNSRecord(rc *cloudflare.ResourceContainer, params cloudflare.CreateDNSRecordParams) {

	record, err := cf.API.CreateDNSRecord(cf.CTX, rc, params)

	if err != nil {
		slog.ErrorContext(cf.CTX, err.Error(), err)
	}

	slog.InfoContext(cf.CTX, "Record created",
		slog.String("Name", record.Name),
		slog.String("Content", record.Content),
		slog.Bool("Proxied", *record.Proxied),
		slog.Int("TTL", record.TTL),
		slog.String("CreatedOn", record.CreatedOn.String()),
		slog.String("ModifiedOn", record.ModifiedOn.String()),
		slog.Bool("Updated", false),
		slog.Bool("Created", true),
	)
}

// This function updates a DNS record and logs the changes.
func (cf *CF) updateDNSRecord(rc *cloudflare.ResourceContainer, params cloudflare.UpdateDNSRecordParams) {

	record, err := cf.API.UpdateDNSRecord(cf.CTX, rc, params)
	if err != nil {
		slog.ErrorContext(cf.CTX, "UpdateDNSRecord error", err)
	}

	slog.InfoContext(cf.CTX, "Record updated",
		slog.String("Name", record.Name),
		slog.String("Content", record.Content),
		slog.Bool("Proxied", *record.Proxied),
		slog.Int("TTL", record.TTL),
		// slog.String("CreatedOn", record.CreatedOn.String()),
		// slog.String("ModifiedOn", record.ModifiedOn.String()),
		slog.Bool("Updated", true),
		slog.Bool("Created", false),
	)
}

// The function updates a DNS record in Cloudflare by either creating a new record or updating an
// existing one.
func (cf *CF) RunDNSUpdate(record cloudflare.DNSRecord) {
	zoneID := cf.GetZoneID(record.ZoneName)

	rc := cloudflare.ZoneIdentifier(zoneID)

	// listing records, because we might not have their IDs
	recs, _, err := cf.API.ListDNSRecords(cf.CTX, rc, cloudflare.ListDNSRecordsParams{Name: record.Name})
	if err != nil {
		slog.ErrorContext(cf.CTX, err.Error(), err)
	}

	if len(recs) == 0 {
		createParams := cloudflare.CreateDNSRecordParams{
			Name:      record.Name,
			Type:      record.Type,
			Proxied:   record.Proxied,
			Proxiable: record.Proxiable,
			TTL:       record.TTL,
			Content:   record.Content,
			ZoneName:  record.ZoneName,
			ZoneID:    zoneID,
		}
		cf.createDNSRecord(rc, createParams)
	} else {
		for _, item := range recs {

			updateParams := cloudflare.UpdateDNSRecordParams{
				ID:      item.ID,
				Name:    item.Name,
				Type:    record.Type,
				Proxied: record.Proxied,
				TTL:     record.TTL,
				Content: record.Content,
			}

			cf.updateDNSRecord(rc, updateParams)
		}
	}
}
