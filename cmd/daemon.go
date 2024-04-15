package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"log/slog"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wasilak/cloudflare-ddns/libs"
)

// This code defines a Cobra command called `daemon` with a `Use` string of "daemon" and a `Short`
// string of "Run as a daemon". It also sets a `PreRun` function that sets the command's context to the
// provided context `ctx`. The `RunE` function is the main function that will be executed when the
// command is run. It calls the `daemonFunc` function with the command's context and returns any errors
// that occur.
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

// This function runs a daemon that periodically refreshes DNS and notifies if the IP address has
// changed.
func daemonFunc(ctx context.Context) error {

	// This code is parsing the value of the `dnsRefreshTime` configuration parameter from the Viper
	// configuration object as a duration using the `time.ParseDuration` function. If there is an error
	// parsing the duration, it will panic with the error message. The parsed duration value is then
	// stored in the `dnsRefreshTime` variable.
	dnsRefreshTime, err := time.ParseDuration(viper.GetString("dnsRefreshTime"))
	if err != nil {
		panic(err)
	}

	slog.DebugContext(ctx, "Refresh Time", "dnsRefreshTime", dnsRefreshTime)

	ctx, cancel := context.WithCancel(ctx)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGHUP)

	// `defer func() { signal.Stop(signalChan); cancel() }()` is a deferred function call that will be
	// executed when the `daemonFunc` function returns. It stops the signal channel from receiving any
	// more signals and cancels the context, which will cause any child contexts to also be cancelled.
	// This ensures that any resources used by the daemon are properly cleaned up when the function
	// returns.
	defer func() {
		signal.Stop(signalChan)
		cancel()
	}()

	// This code defines an anonymous function that runs as a goroutine. It listens for signals on the
	// `signalChan` channel and takes different actions depending on the signal received. If the signal is
	// `syscall.SIGHUP`, it reloads the configuration file using the `viper.ReadInConfig()` function and
	// logs a message indicating that the new configuration file is being used. If the signal is
	// `os.Interrupt`, it logs a message indicating that the daemon is stopping, cancels the context, and
	// exits the program with an exit code of 1. If the context is cancelled, it logs a message indicating
	// that the function is done and exits the program with an exit code of 1. This goroutine runs
	// concurrently with the main loop of the `daemonFunc` function and allows the daemon to respond to
	// signals while it is running.
	go func() {
		for {
			select {
			case s := <-signalChan:
				switch s {
				case syscall.SIGHUP:
					slog.DebugContext(ctx, "Reloading config...")
					if err := viper.ReadInConfig(); err == nil {
						slog.DebugContext(ctx, "Using config file", "filename", viper.ConfigFileUsed())
					} else {
						slog.ErrorContext(ctx, "Error", err)
					}
				case os.Interrupt:
					slog.DebugContext(ctx, "Stopping...")
					cancel()
					os.Exit(1)
				}
			case <-ctx.Done():
				slog.DebugContext(ctx, "Done.")
				os.Exit(1)
			}
		}
	}()

	currentIp := runRunner("")

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.Tick(dnsRefreshTime):
			runRunner(currentIp)
		}
	}
}

func runRunner(currentIp string) string {
	slog.DebugContext(ctx, "Starting DNS refresh...")

	records := libs.PrepareRecords()
	ip, err := libs.Runner(ctx, records)
	if err != nil {
		slog.ErrorContext(ctx, "Error", err)
	} else if ip != "" && currentIp != ip {
		currentIp = ip
		libs.Notify(ctx, ip)
	}

	slog.DebugContext(ctx, "DNS refresh completed.")

	return currentIp
}
