package cmd

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/wasilak/cloudflare-ddns/libs"
)

var oneoffCmd = &cobra.Command{
	Use:   "oneoff",
	Short: "Run once and exit",
	PreRun: func(cmd *cobra.Command, args []string) {
		cmd.SetContext(ctx)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := oneoffFunc(cmd.Context()); err != nil {
			return err
		}
		return nil
	},
}

func oneoffFunc(ctx context.Context) error {
	_, err := libs.Runner(ctx)
	if err != nil {
		return err
	}
	return nil
}
