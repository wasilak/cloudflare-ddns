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

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig, initLogging)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cloudflare-ddns/config.yml)")
	// viper.BindPFlag("author", rootCmd.PersistentFlags().Lookup("author"))
	// viper.BindPFlag("useViper", rootCmd.PersistentFlags().Lookup("viper"))
	// viper.SetDefault("author", "NAME HERE <EMAIL ADDRESS>")
	// viper.SetDefault("license", "apache")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(oneoffCmd)
}

func initConfig() {
	godotenv.Load()

	viper.BindEnv("CF.APIKey", "CF_API_KEY")
	viper.BindEnv("CF.APIEmail", "CF_API_EMAIL")

	viper.SetDefault("LogFile", "/var/log/cloudflare-dns.log")

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
		// viper.AddConfigPath(".")
		viper.AddConfigPath(home)
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		log.Printf("%+v\n", err)
	}
}

func initLogging() {
	opts := slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	}

	file, err := os.OpenFile(viper.GetString("LogFile"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	mw := io.MultiWriter(os.Stdout, file)

	textHandler := opts.NewTextHandler(mw)
	logger := slog.New(textHandler)

	ctx = context.WithValue(ctx, "logger", logger)
}
