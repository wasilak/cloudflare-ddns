package cmd

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/wasilak/cloudflare-ddns/libs"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Run once -> delete records -> exit",
	PreRun: func(cmd *cobra.Command, args []string) {
		cmd.SetContext(ctx)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := deleteFunc(cmd.Context()); err != nil {
			return err
		}
		return nil
	},
}

// The function calls the Runner function from the libs package and returns any errors encountered.
func deleteFunc(ctx context.Context) error {
	currentIp, err := libs.GetIP()
	if err != nil {
		return err
	}

	records := libs.PrepareRecords()
	err = libs.Runner(ctx, currentIp, records, true)
	if err != nil {
		return err
	}
	return nil
}
