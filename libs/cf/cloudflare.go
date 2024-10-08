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

type ExtendedCloudflareDNSRecord struct {
	Record          cloudflare.DNSRecord `mapstructure:"record"`
	KeepAfterDelete bool                 `mapstructure:"keep_after_delete,omitempty"`
	CNAME           string               `mapstructure:"CNAME,omitempty"`
	ZoneName        string               `mapstructure:"zone_name"`
}

// The function initializes a Cloudflare API client with the provided API key, email, and context.
func (cf *CF) Init(CFAPIKey, CFAPIEmail string, ctx context.Context) {
	cf.CTX = ctx
	var err error

	cf.API, err = cloudflare.New(CFAPIKey, CFAPIEmail)
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		os.Exit(1)
	}
}

// The function takes a zone name as input and returns its corresponding zone ID using an API call.
func (cf *CF) GetZoneID(zoneName string) string {
	zoneID, err := cf.API.ZoneIDByName(zoneName)
	if err != nil {
		slog.With("zoneName", zoneName).ErrorContext(cf.CTX, "Error GetZoneID", "error", err)
		os.Exit(1)
	}

	return zoneID
}

// This function retrieves a DNS record from Cloudflare using its ID.
func (cf *CF) GetDNSRecord(rc *cloudflare.ResourceContainer, record ExtendedCloudflareDNSRecord) (ExtendedCloudflareDNSRecord, error) {

	recordGet, err := cf.API.GetDNSRecord(cf.CTX, rc, record.Record.ID)
	if err != nil {
		return ExtendedCloudflareDNSRecord{}, err
	}

	convertedRecord := ExtendedCloudflareDNSRecord{
		Record: recordGet,
	}

	return convertedRecord, nil
}

// The function creates a DNS record and logs its details.
func (cf *CF) createDNSRecord(rc *cloudflare.ResourceContainer, params cloudflare.CreateDNSRecordParams) error {

	record, err := cf.API.CreateDNSRecord(cf.CTX, rc, params)
	if err != nil {
		return err
	}

	slog.With("params", params).InfoContext(cf.CTX, "Record created",
		slog.String("Name", record.Name),
		slog.String("Content", record.Content),
		slog.Bool("Proxied", *record.Proxied),
		slog.Int("TTL", record.TTL),
		slog.String("CreatedOn", record.CreatedOn.String()),
		slog.String("ModifiedOn", record.ModifiedOn.String()),
		slog.Bool("Updated", false),
		slog.Bool("Created", true),
	)

	return nil
}

// This function updates a DNS record and logs the changes.
func (cf *CF) updateDNSRecord(rc *cloudflare.ResourceContainer, params cloudflare.UpdateDNSRecordParams) {

	record, err := cf.API.UpdateDNSRecord(cf.CTX, rc, params)
	if err != nil {
		slog.With("params", params).ErrorContext(cf.CTX, "UpdateDNSRecord error")
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

// This function deletes a DNS record and logs the changes.
func (cf *CF) deleteDNSRecord(rc *cloudflare.ResourceContainer, record ExtendedCloudflareDNSRecord) {

	err := cf.API.DeleteDNSRecord(cf.CTX, rc, record.Record.ID)
	if err != nil {
		slog.With("record", record).ErrorContext(cf.CTX, "DeleteDNSRecord error", "msg", err)
	}

	slog.InfoContext(cf.CTX, "Record deleted",
		slog.String("Name", record.Record.Name),
		slog.String("Content", record.Record.Content),
		slog.Bool("Proxied", *record.Record.Proxied),
		slog.Int("TTL", record.Record.TTL),
	)
}

// The function updates a DNS record in Cloudflare by either creating a new record or updating an
// existing one.
func (cf *CF) RunDNSUpdate(record ExtendedCloudflareDNSRecord, deleteRecords bool) {

	zoneID := cf.GetZoneID(record.ZoneName)
	rc := cloudflare.ZoneIdentifier(zoneID)

	// listing records, because we might not have their IDs
	recs, _, err := cf.API.ListDNSRecords(cf.CTX, rc, cloudflare.ListDNSRecordsParams{Name: record.Record.Name})
	if err != nil {
		slog.With("record", record).ErrorContext(cf.CTX, "Error", "error", err)
	}

	if len(recs) == 0 {
		createParams := cloudflare.CreateDNSRecordParams{
			Name:      record.Record.Name,
			Type:      record.Record.Type,
			Proxied:   record.Record.Proxied,
			Proxiable: record.Record.Proxiable,
			TTL:       record.Record.TTL,
			Content:   record.Record.Content,
		}
		cf.createDNSRecord(rc, createParams)
		if err != nil {
			slog.With("params", createParams).ErrorContext(cf.CTX, err.Error())
		}
	} else {
		for _, item := range recs {

			if deleteRecords && !record.KeepAfterDelete {
				record.Record.ID = item.ID
				cf.deleteDNSRecord(rc, record)
			} else {
				updateParams := cloudflare.UpdateDNSRecordParams{
					ID:      item.ID,
					Name:    item.Name,
					Type:    record.Record.Type,
					Proxied: record.Record.Proxied,
					TTL:     record.Record.TTL,
					Content: record.Record.Content,
				}

				cf.updateDNSRecord(rc, updateParams)
			}
		}
	}
}
