package cmd

import (
	"context"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/wasilak/cloudflare-ddns/libs"
	"github.com/wasilak/cloudflare-ddns/libs/ip"
)

// This code defines a Cobra command called "oneoff" which can be executed from the command line. The
// command has a short description "Run once and exit". It also has a PreRun function that sets the
// context of the command to the context passed in as an argument. The RunE function executes the
// oneoffFunc function from the libs package, passing in the context of the command. If there are any
// errors encountered, they are returned.
var oneoffCmd = &cobra.Command{
	Use:   "oneoff",
	Short: "Run once and exit",
	PreRun: func(cmd *cobra.Command, args []string) {
		cmd.SetContext(ctx)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := oneOffFunc(cmd.Context()); err != nil {
			return err
		}
		return nil
	},
}

// The function calls the Runner function from the libs package and returns any errors encountered.
func oneOffFunc(ctx context.Context) error {
	var err error
	ip.CurrentIp, err = ip.GetIP(ctx)
	if err != nil {
		return err
	} else {
		slog.DebugContext(ctx, "External IP", "ip", ip.CurrentIp.IP, "ip_source", ip.CurrentIp.Source.GetName())
	}

	records := libs.PrepareRecords()
	err = libs.Runner(ctx, records)
	if err != nil {
		return err
	}
	return nil
}
