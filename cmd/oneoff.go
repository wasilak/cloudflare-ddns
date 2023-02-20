package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"golang.org/x/exp/slog"
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
	logger := ctx.Value("logger").(*slog.Logger)
	logger.Debug(fmt.Sprintf("oneoff!!"))
	return nil
}
