package cf

import (
	"context"
	"fmt"
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

// Init func
func Init(CFAPIKey, CFAPIEmail string, CFctx context.Context) {
	ctx = CFctx

	logger = ctx.Value("logger").(*slog.Logger)

	api, err = cloudflare.New(CFAPIKey, CFAPIEmail)
	if err != nil {
		logger.Error(err.Error(), err)
		os.Exit(1)
	}
}

func GetZoneID(zoneName string) string {
	zoneID, err := api.ZoneIDByName(zoneName)
	if err != nil {
		logger.Error(err.Error(), err)
	}

	return zoneID
}

func GetDNSRecord(rc *cloudflare.ResourceContainer, record cloudflare.DNSRecord) (cloudflare.DNSRecord, error) {

	record, err := api.GetDNSRecord(ctx, rc, record.ID)
	if err != nil {
		return record, err
	}

	return record, nil
}

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

func updateDNSRecord(rc *cloudflare.ResourceContainer, params cloudflare.UpdateDNSRecordParams) {

	fmt.Printf("%+v\n", params)
	err := api.UpdateDNSRecord(ctx, rc, params)
	if err != nil {
		fmt.Printf("UpdateDNSRecord error: %+v\n", err)
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
