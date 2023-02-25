package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"
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

	ip, err := libs.GetIP()
	if err != nil {
		return err
	}

	dnsRefreshTime, err := time.ParseDuration(viper.GetString("dnsRefreshTime"))
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(ctx)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGHUP)

	defer func() {
		signal.Stop(signalChan)
		cancel()
	}()

	go func() {
		for {
			select {
			case s := <-signalChan:
				switch s {
				case syscall.SIGHUP:
					logger.Debug("Reloading config...")
				case os.Interrupt:
					logger.Debug("Stopping...")
					cancel()
					os.Exit(1)
				}
			case <-ctx.Done():
				logger.Debug("Done.")
				os.Exit(1)
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.Tick(dnsRefreshTime):
			logger.Debug("Starting DNS refresh...")
			libs.Runner(ctx, ip)
			logger.Debug("DNS refresh completed.")
		}
	}
}
