package cf

import (
	"context"
	"log/slog"

	cloudflare "github.com/cloudflare/cloudflare-go/v4"
	"github.com/cloudflare/cloudflare-go/v4/dns"
	"github.com/cloudflare/cloudflare-go/v4/option"
	"github.com/cloudflare/cloudflare-go/v4/zones"
)

type CF struct {
	Client *cloudflare.Client
}

type ExtendedCloudflareDNSRecord struct {
	Record   *dns.RecordResponse `mapstructure:"record" json:"record"`
	CNAME    string              `mapstructure:"CNAME,omitempty" json:"CNAME,omitempty"`
	ZoneName string              `mapstructure:"zone_name" json:"zone_name"`
}

// The function initializes a Cloudflare API client with the provided API key, email, and context.
func (cf *CF) Init(CFAPIKey, CFAPIEmail string) {
	cf.Client = cloudflare.NewClient(
		option.WithAPIKey(CFAPIKey),     // defaults to os.LookupEnv("CLOUDFLARE_API_KEY")
		option.WithAPIEmail(CFAPIEmail), // defaults to os.LookupEnv("CLOUDFLARE_EMAIL")
	)
}

func (cf *CF) GetZonesList(ctx context.Context, zoneName string) (string, error) {
	zones, err := cf.Client.Zones.List(ctx, zones.ZoneListParams{
		Name: cloudflare.F(zoneName),
	})
	if err != nil {
		slog.With("zoneName", zoneName).ErrorContext(ctx, "Error GetZoneID", "error", err)
		return "", err
	}

	if len(zones.Result) == 0 {
		slog.With("zoneName", zoneName).ErrorContext(ctx, "Zone not found")
		return "", nil
	}

	return zones.Result[0].ID, nil
}

// This function retrieves a DNS record from Cloudflare using its name.
func (cf *CF) ListDNSRecords(ctx context.Context, zoneID string) ([]ExtendedCloudflareDNSRecord, error) {
	result, err := cf.Client.DNS.Records.List(ctx, dns.RecordListParams{
		ZoneID: cloudflare.F(zoneID),
	})
	if err != nil {
		return nil, err
	}

	records := make([]ExtendedCloudflareDNSRecord, 0)

	for _, item := range result.Result {
		records = append(records, ExtendedCloudflareDNSRecord{
			Record: &item,
		})
	}

	return records, nil
}

// This function retrieves a DNS record from Cloudflare using its name.
func (cf *CF) GetDNSRecord(ctx context.Context, record ExtendedCloudflareDNSRecord, zoneID string) (*ExtendedCloudflareDNSRecord, error) {
	result, err := cf.Client.DNS.Records.List(ctx, dns.RecordListParams{
		ZoneID: cloudflare.F(zoneID),
		Name:   cloudflare.F(dns.RecordListParamsName{Exact: cloudflare.F(record.Record.Name)}),
	})
	if err != nil {
		return nil, err
	}

	var recordGet *dns.RecordResponse
	for _, item := range result.Result {
		if item.Name == record.Record.Name {
			recordGet = &item
			break
		}
	}

	convertedRecord := ExtendedCloudflareDNSRecord{
		Record: recordGet,
	}

	return &convertedRecord, nil
}

// The function creates a DNS record and logs its details.
func (cf *CF) CreateDNSRecord(ctx context.Context, params dns.RecordNewParams) (*dns.RecordResponse, error) {

	record, err := cf.Client.DNS.Records.New(ctx, params)
	if err != nil {
		return nil, err
	}

	slog.With("params", params).InfoContext(ctx, "Record created",
		slog.String("Name", record.Name),
		slog.String("Content", record.Content),
		slog.Bool("Proxied", record.Proxied),
		slog.Int("TTL", int(record.TTL)),
		slog.String("CreatedOn", record.CreatedOn.String()),
		slog.String("ModifiedOn", record.ModifiedOn.String()),
		slog.Bool("Updated", false),
		slog.Bool("Created", true),
	)

	return record, nil
}

// This function updates a DNS record and logs the changes.
func (cf *CF) UpdateDNSRecord(ctx context.Context, recordId string, params dns.RecordUpdateParams) (*dns.RecordResponse, error) {

	record, err := cf.Client.DNS.Records.Update(ctx, recordId, params)
	if err != nil {
		slog.With("params", params).ErrorContext(ctx, "UpdateDNSRecord error", "err", err)
		return nil, err
	}

	slog.InfoContext(ctx, "Record updated",
		slog.String("Name", record.Name),
		slog.String("Content", record.Content),
		slog.Bool("Proxied", record.Proxied),
		slog.Int("TTL", int(record.TTL)),
		slog.Bool("Updated", true),
		slog.Bool("Created", false),
	)

	return record, nil
}

// This function deletes a DNS record and logs the changes.
func (cf *CF) DeleteDNSRecord(ctx context.Context, record ExtendedCloudflareDNSRecord, zoneID string) (*dns.RecordDeleteResponse, error) {

	response, err := cf.Client.DNS.Records.Delete(ctx, record.Record.ID, dns.RecordDeleteParams{
		ZoneID: cloudflare.F(zoneID),
	})
	if err != nil {
		slog.With("record", record).ErrorContext(ctx, "DeleteDNSRecord error", "msg", err)
		return response, err
	}

	slog.InfoContext(ctx, "Record deleted",
		slog.String("Name", record.Record.Name),
		slog.String("Content", record.Record.Content),
		slog.Bool("Proxied", record.Record.Proxied),
		slog.Int("TTL", int(record.Record.TTL)),
	)
	return response, nil
}
