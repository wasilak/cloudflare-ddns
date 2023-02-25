package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wasilak/cloudflare-ddns/libs"
	"golang.org/x/exp/slog"
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Run as a daemon",
	PreRun: func(cmd *cobra.Command, args []string) {
		cmd.SetContext(ctx)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := daemonFunc(cmd.Context()); err != nil {
			return err
		}
		return nil
	},
}

func daemonFunc(ctx context.Context) error {
	logger := ctx.Value("logger").(*slog.Logger)
	logger.Debug("daemonFunc")

	ip, err := libs.GetIP()
	if err != nil {
		return err
	}

	dnsRefreshTime, err := time.ParseDuration(viper.GetString("dnsRefreshTime"))
	if err != nil {
		panic(err)
	}

	ticker := time.NewTicker(dnsRefreshTime)
	logger.Debug(fmt.Sprintf("%+v", dnsRefreshTime))

	go func() {

		for range ticker.C {
			logger.Debug("Starting DNS refresh...")
			libs.Runner(ctx, ip)
			logger.Debug("DNS refresh completed.")
		}
	}()

	defer ticker.Stop()

	return nil
}
