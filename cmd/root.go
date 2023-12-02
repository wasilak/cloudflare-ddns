package cmd

import (
	"context"
	"fmt"
	"os"

	"log/slog"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wasilak/loggergo"
)

// This code block is defining variables and initializing a root command for a command-line interface
// tool.
var (
	cfgFile string
	rootCmd = &cobra.Command{
		Use:   "cloudflare-ddns",
		Short: "Cloudflare dynamic DNS",
		// Version: version,
		PreRun: func(cmd *cobra.Command, args []string) {
			cmd.SetContext(ctx)
		},
		Run: func(cmd *cobra.Command, args []string) {
		},
	}
	ctx = context.Background()
)

// The function executes a root command and prints any errors to the standard error output.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// This function initializes the command-line interface for a Cloudflare DDNS tool, including setting
// up configuration and logging options, and adding commands for version information, one-off updates,
// and running as a daemon.
func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cloudflare-ddns/config.yml)")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(oneoffCmd)
	rootCmd.AddCommand(daemonCmd)
}

// The function initializes the configuration settings for a Go program, including loading environment
// variables and a YAML config file if present.
func initConfig() {
	godotenv.Load()

	viper.BindEnv("CF.APIKey", "CF_API_KEY")
	viper.BindEnv("CF.APIEmail", "CF_API_EMAIL")

	viper.SetDefault("loglevel", "info")
	viper.SetDefault("logformat", "plain")
	viper.SetDefault("dnsRefreshTime", "60s")
	viper.SetDefault("mail.subject", "Your External IP has changed!")

	// This code block is initializing the configuration settings for a Go program. It checks if a config
	// file path has been provided as a command-line argument, and if so, sets the configuration file to
	// that path. If not, it searches for a YAML config file named "config" in the user's home directory.
	// It also sets the configuration type to YAML and adds the home directory as a search path for the
	// config file.
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".config" (without extension).
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
		viper.AddConfigPath(home)
	}

	viper.AutomaticEnv()

	// This code block is checking if a configuration file has been successfully read by the viper
	// package. If the configuration file has been read successfully, it prints a message to the console
	// indicating which configuration file is being used. If the configuration file could not be read, it
	// logs the error message to the console.
	if err := viper.ReadInConfig(); err == nil {
		slog.DebugContext(ctx, "Using config file", "filename", viper.ConfigFileUsed())
	} else {
		slog.ErrorContext(ctx, "error", err)
	}

	loggerConfig := loggergo.LoggerGoConfig{
		Level:  viper.GetString("loglevel"),
		Format: viper.GetString("logformat"),
	}

	_, err := loggergo.LoggerInit(loggerConfig)
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		os.Exit(1)
	}
}
