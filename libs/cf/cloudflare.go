package cf

import (
	"context"
	"os"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"golang.org/x/exp/slog"
)

var (
	api    *cloudflare.API
	err    error
	ctx    context.Context
	logger *slog.Logger
)

// The function initializes a Cloudflare API client with the provided API key, email, and context.
func Init(CFAPIKey, CFAPIEmail string, CFctx context.Context) {
	ctx = CFctx

	logger = ctx.Value("logger").(*slog.Logger)

	api, err = cloudflare.New(CFAPIKey, CFAPIEmail)
	if err != nil {
		logger.Error(err.Error(), err)
		os.Exit(1)
	}
}

// The function takes a zone name as input and returns its corresponding zone ID using an API call.
func GetZoneID(zoneName string) string {
	zoneID, err := api.ZoneIDByName(zoneName)
	if err != nil {
		logger.Error(err.Error(), err)
	}

	return zoneID
}

// This function retrieves a DNS record from Cloudflare using its ID.
func GetDNSRecord(rc *cloudflare.ResourceContainer, record cloudflare.DNSRecord) (cloudflare.DNSRecord, error) {

	record, err := api.GetDNSRecord(ctx, rc, record.ID)
	if err != nil {
		return record, err
	}

	return record, nil
}

// The function creates a DNS record and logs its details.
func createDNSRecord(rc *cloudflare.ResourceContainer, params cloudflare.CreateDNSRecordParams) {

	record, err := api.CreateDNSRecord(ctx, rc, params)

	if err != nil {
		logger.Error(err.Error(), err)
	}

	logger.Info("Record created",
		slog.String("Name", record.Result.Name),
		slog.String("Content", record.Result.Content),
		slog.Bool("Proxied", *record.Result.Proxied),
		slog.Int("TTL", record.Result.TTL),
		slog.String("CreatedOn", record.Result.CreatedOn.String()),
		slog.String("ModifiedOn", record.Result.ModifiedOn.String()),
		slog.Bool("Updated", false),
		slog.Bool("Created", true),
	)
}

// This function updates a DNS record and logs the changes.
func updateDNSRecord(rc *cloudflare.ResourceContainer, params cloudflare.UpdateDNSRecordParams) {

	err := api.UpdateDNSRecord(ctx, rc, params)
	if err != nil {
		logger.Debug("UpdateDNSRecord error", err)
	}

	logger.Info("Record updated",
		slog.String("Name", params.Name),
		slog.String("Content", params.Content),
		slog.Bool("Proxied", *params.Proxied),
		slog.Int("TTL", params.TTL),
		// slog.String("CreatedOn", params.CreatedOn.String()),
		// slog.String("ModifiedOn", params.ModifiedOn.String()),
		slog.Bool("Updated", true),
		slog.Bool("Created", false),
	)
}

// The function updates a DNS record in Cloudflare by either creating a new record or updating an
// existing one.
func RunDNSUpdate(record cloudflare.DNSRecord) {
	zoneID := GetZoneID(record.ZoneName)

	rc := cloudflare.ZoneIdentifier(zoneID)

	// listing records, because we might not have their IDs
	recs, _, err := api.ListDNSRecords(ctx, rc, cloudflare.ListDNSRecordsParams{Name: record.Name})
	if err != nil {
		logger.Error(err.Error(), err)
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
		createDNSRecord(rc, createParams)
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

			updateDNSRecord(rc, updateParams)
		}
	}
}
