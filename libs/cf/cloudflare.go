package cf

import (
	"context"
	"os"

	"log/slog"

	cloudflare "github.com/cloudflare/cloudflare-go/v4"
	"github.com/cloudflare/cloudflare-go/v4/dns"
	"github.com/cloudflare/cloudflare-go/v4/option"
	"github.com/cloudflare/cloudflare-go/v4/zones"
)

type CF struct {
	Client *cloudflare.Client
	CTX    context.Context
}

type ExtendedCloudflareDNSRecord struct {
	Record          *dns.RecordResponse `mapstructure:"record"`
	KeepAfterDelete bool                `mapstructure:"keep_after_delete,omitempty"`
	CNAME           string              `mapstructure:"CNAME,omitempty"`
	ZoneName        string              `mapstructure:"zone_name"`
}

// The function initializes a Cloudflare API client with the provided API key, email, and context.
func (cf *CF) Init(CFAPIKey, CFAPIEmail string, ctx context.Context) {
	cf.CTX = ctx

	cf.Client = cloudflare.NewClient(
		option.WithAPIKey(CFAPIKey),     // defaults to os.LookupEnv("CLOUDFLARE_API_KEY")
		option.WithAPIEmail(CFAPIEmail), // defaults to os.LookupEnv("CLOUDFLARE_EMAIL")
	)
}

func (cf *CF) GetZonesList(zoneName string) string {
	zones, err := cf.Client.Zones.List(cf.CTX, zones.ZoneListParams{
		Name: cloudflare.F(zoneName),
	})
	if err != nil {
		slog.With("zoneName", zoneName).ErrorContext(cf.CTX, "Error GetZoneID", "error", err)
		os.Exit(1)
	}

	return zones.Result[0].ID
}

// This function retrieves a DNS record from Cloudflare using its ID.
func (cf *CF) GetDNSRecord(record ExtendedCloudflareDNSRecord) (ExtendedCloudflareDNSRecord, error) {

	recordGet, err := cf.Client.DNS.Records.Get(cf.CTX, record.Record.ID, dns.RecordGetParams{})
	if err != nil {
		return ExtendedCloudflareDNSRecord{}, err
	}

	convertedRecord := ExtendedCloudflareDNSRecord{
		Record: recordGet,
	}

	return convertedRecord, nil
}

// The function creates a DNS record and logs its details.
func (cf *CF) createDNSRecord(params dns.RecordNewParams) error {

	record, err := cf.Client.DNS.Records.New(cf.CTX, params)
	if err != nil {
		return err
	}

	slog.With("params", params).InfoContext(cf.CTX, "Record created",
		slog.String("Name", record.Name),
		slog.String("Content", record.Content),
		slog.Bool("Proxied", record.Proxied),
		slog.Int("TTL", int(record.TTL)),
		slog.String("CreatedOn", record.CreatedOn.String()),
		slog.String("ModifiedOn", record.ModifiedOn.String()),
		slog.Bool("Updated", false),
		slog.Bool("Created", true),
	)

	return nil
}

// This function updates a DNS record and logs the changes.
func (cf *CF) updateDNSRecord(recordId string, params dns.RecordUpdateParams) {

	record, err := cf.Client.DNS.Records.Update(cf.CTX, recordId, params)
	if err != nil {
		slog.With("params", params).ErrorContext(cf.CTX, "UpdateDNSRecord error")
	}

	slog.InfoContext(cf.CTX, "Record updated",
		slog.String("Name", record.Name),
		slog.String("Content", record.Content),
		slog.Bool("Proxied", record.Proxied),
		slog.Int("TTL", int(record.TTL)),
		// slog.String("CreatedOn", record.CreatedOn.String()),
		// slog.String("ModifiedOn", record.ModifiedOn.String()),
		slog.Bool("Updated", true),
		slog.Bool("Created", false),
	)
}

// This function deletes a DNS record and logs the changes.
func (cf *CF) deleteDNSRecord(record ExtendedCloudflareDNSRecord) {

	_, err := cf.Client.DNS.Records.Delete(cf.CTX, record.Record.ID, dns.RecordDeleteParams{})
	if err != nil {
		slog.With("record", record).ErrorContext(cf.CTX, "DeleteDNSRecord error", "msg", err)
	}

	slog.InfoContext(cf.CTX, "Record deleted",
		slog.String("Name", record.Record.Name),
		slog.String("Content", record.Record.Content),
		slog.Bool("Proxied", record.Record.Proxied),
		slog.Int("TTL", int(record.Record.TTL)),
	)
}

// The function updates a DNS record in Cloudflare by either creating a new record or updating an
// existing one.
func (cf *CF) RunDNSUpdate(record ExtendedCloudflareDNSRecord, deleteRecords bool) {

	zoneID := cf.GetZonesList(record.ZoneName)

	listParams := dns.RecordListParams{
		ZoneID: cloudflare.F(zoneID),
		Name: cloudflare.F(dns.RecordListParamsName{
			Exact: cloudflare.F(record.Record.Name),
		}),
	}

	// listing records, because we might not have their IDs
	recs, err := cf.Client.DNS.Records.List(cf.CTX, listParams)
	if err != nil {
		slog.With("record", record).ErrorContext(cf.CTX, "Error", "error", err)
	}

	if len(recs.Result) == 0 {
		createParams := dns.RecordNewParams{
			ZoneID: cloudflare.F(zoneID),
			Record: dns.RecordParam{
				Name:    cloudflare.F(record.Record.Name),
				Type:    cloudflare.F(dns.RecordType(record.Record.Type)),
				Proxied: cloudflare.F(record.Record.Proxied),
				TTL:     cloudflare.F(record.Record.TTL),
				Content: cloudflare.F(record.Record.Content),
			},
		}
		cf.createDNSRecord(createParams)
		if err != nil {
			slog.With("params", createParams).ErrorContext(cf.CTX, err.Error())
		}
	} else {
		for _, item := range recs.Result {

			if deleteRecords && !record.KeepAfterDelete {
				record.Record.ID = item.ID
				cf.deleteDNSRecord(record)
			} else {
				updateParams := dns.RecordUpdateParams{
					ZoneID: cloudflare.F(zoneID),
					Record: dns.RecordParam{
						Name:    cloudflare.F(record.Record.Name),
						Type:    cloudflare.F(dns.RecordType(record.Record.Type)),
						Proxied: cloudflare.F(record.Record.Proxied),
						TTL:     cloudflare.F(record.Record.TTL),
						Content: cloudflare.F(record.Record.Content),
					},
				}

				cf.updateDNSRecord(item.ID, updateParams)
			}
		}
	}
}
