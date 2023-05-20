package cmd

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/exp/slog"
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
			logger := cmd.Context().Value("logger").(*slog.Logger)
			logger.Debug(fmt.Sprintf("%+v", viper.AllSettings()))
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
	cobra.OnInitialize(initConfig, initLogging)

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

	viper.SetDefault("LogFile", "/var/log/cloudflare-dns.log")
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
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		log.Printf("%+v\n", err)
	}
}

// The function initializes logging with a debug level and writes logs to a file and standard output.
func initLogging() {
	opts := slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	}

	// This code block is initializing a file for logging. It uses the `os.OpenFile` function to create or
	// open a file with the path specified in the configuration settings (`viper.GetString("LogFile")`).
	// The `os.O_CREATE|os.O_WRONLY|os.O_APPEND` flags are used to create the file if it does not exist,
	// open it for writing, and append to the end of the file if it already exists. The `0666` permission
	// bits are used to set the file permissions to read and write for all users. If an error occurs while
	// opening or creating the file, the function logs the error message and exits the program. The
	// `io.MultiWriter` function is used to create a writer that writes to both standard output and the
	// log file.
	file, err := os.OpenFile(viper.GetString("LogFile"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	mw := io.MultiWriter(os.Stdout, file)

	textHandler := opts.NewTextHandler(mw)
	logger := slog.New(textHandler)

	ctx = context.WithValue(ctx, "logger", logger)
}
