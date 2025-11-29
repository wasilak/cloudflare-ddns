package api

import (
	"context"
	"fmt"
	"log/slog"

	cloudflare "github.com/cloudflare/cloudflare-go/v4"
	"github.com/cloudflare/cloudflare-go/v4/dns"
	"github.com/wasilak/cloudflare-ddns/libs/cf"
	"github.com/wasilak/cloudflare-ddns/libs/ip"
)

var CfAPI cf.CF
var Records = &[]cf.ExtendedCloudflareDNSRecord{}

// The function updates a DNS record in Cloudflare by either creating a new record or updating an
// existing one.
func RunDNSUpdate(ctx context.Context, record cf.ExtendedCloudflareDNSRecord) error {

	zoneID, err := CfAPI.GetZonesList(ctx, record.ZoneName)
	if err != nil {
		slog.With("record", record).ErrorContext(ctx, "Error", "error", err)
		return err
	}

	r, err := CfAPI.GetDNSRecord(ctx, record, zoneID)
	if err != nil {
		return err
	}

	if r == nil || r.Record == nil || r.Record.ID == "" {
		AddRecord(ctx, &record)
	} else {
		UpdateRecord(ctx, &record)
	}

	return nil
}

func FindDNSRecordByName(recordName string) *cf.ExtendedCloudflareDNSRecord {
	for _, r := range *Records {
		if r.Record.Name == recordName {
			return &r
		}
	}
	return nil
}

func DeleteRecord(ctx context.Context, recordName string, zoneName string) (*cf.ExtendedCloudflareDNSRecord, error) {
	var err error

	zoneID, err := CfAPI.GetZonesList(ctx, zoneName)
	if err != nil {
		return nil, err
	}

	record := &cf.ExtendedCloudflareDNSRecord{
		Record: &dns.RecordResponse{
			Name: recordName,
		},
	}

	record, err = CfAPI.GetDNSRecord(ctx, *record, zoneID)
	if err != nil {
		return nil, err
	}

	if record == nil || record.Record == nil || record.Record.ID == "" {
		return nil, fmt.Errorf("record not found")
	}

	response, err := CfAPI.DeleteDNSRecord(ctx, *record, zoneID)
	if err != nil {
		return nil, err
	}

	slog.With("record", record).DebugContext(ctx, "Record deleted", "response", response)

	for i, r := range *Records {
		if r.Record.Name == record.Record.Name {
			*Records = (*Records)[:i]
			*Records = append(*Records, (*Records)[i+1:]...)
			break
		}
	}
	return record, nil
}

func AddRecord(ctx context.Context, record *cf.ExtendedCloudflareDNSRecord) (*cf.ExtendedCloudflareDNSRecord, error) {
	zoneID, err := CfAPI.GetZonesList(ctx, record.ZoneName)
	if err != nil {
		return nil, err
	}

	createParams := dns.RecordNewParams{
		ZoneID: cloudflare.F(zoneID),
		Body: dns.ARecordParam{
			Name:    cloudflare.F(record.Record.Name),
			Type:    cloudflare.F(dns.ARecordTypeA),
			Content: cloudflare.F(ip.CurrentIp.IP),
			TTL:     cloudflare.F(dns.TTL(record.Record.TTL)),
			Proxied: cloudflare.F(record.Record.Proxied),
		},
	}

	response, err := CfAPI.CreateDNSRecord(ctx, createParams)
	if err != nil {
		return nil, err
	}

	slog.With("record", record).DebugContext(ctx, "Record create", "response", response)

	*Records = append(*Records, *record)
	return record, nil
}

func UpdateRecord(ctx context.Context, updatedRecord *cf.ExtendedCloudflareDNSRecord) (*cf.ExtendedCloudflareDNSRecord, error) {
	zoneID, err := CfAPI.GetZonesList(ctx, updatedRecord.ZoneName)
	if err != nil {
		return nil, err
	}

	record, err := CfAPI.GetDNSRecord(ctx, *updatedRecord, zoneID)
	if err != nil {
		return nil, err
	}

	if record == nil || record.Record == nil || record.Record.ID == "" {
		return nil, fmt.Errorf("record not found")
	}

	slog.With(PrepareRecordForLoggiong("record", record)).DebugContext(ctx, "Updating record", PrepareRecordForLoggiong("updatedRecord", updatedRecord))

	record.ZoneName = updatedRecord.ZoneName

	updateParams := dns.RecordUpdateParams{
		ZoneID: cloudflare.F(zoneID),
		Body: dns.ARecordParam{
			Name:    cloudflare.F(updatedRecord.Record.Name),
			Type:    cloudflare.F(dns.ARecordTypeA),
			Content: cloudflare.F(ip.CurrentIp.IP),
			TTL:     cloudflare.F(dns.TTL(updatedRecord.Record.TTL)),
			Proxied: cloudflare.F(updatedRecord.Record.Proxied),
		},
	}

	response, err := CfAPI.UpdateDNSRecord(ctx, record.Record.ID, updateParams)
	if err != nil {
		for i, r := range *Records {
			if r.Record.Name == record.Record.Name {
				records := *Records
				*Records = append(records[:i], records[i+1:]...)
				break
			}
		}
		return nil, err
	}

	for i, r := range *Records {
		if r.Record.Name == record.Record.Name {
			(*Records)[i] = cf.ExtendedCloudflareDNSRecord{
				Record:   response,
				ZoneName: record.ZoneName,
				CNAME:    record.CNAME,
			}
			break
		}
	}

	return record, nil
}

func PrepareRecordForLoggiong(name string, record *cf.ExtendedCloudflareDNSRecord) slog.Attr {
	return slog.Group(name,
		slog.String("name", record.Record.Name),
		slog.String("content", record.Record.Content),
		slog.Bool("proxied", record.Record.Proxied),
		slog.Int("ttl", int(record.Record.TTL)),
		slog.String("zoneName", record.ZoneName),
		slog.String("type", string(record.Record.Type)),
	)
}
