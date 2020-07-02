package cmd

import (
	"github.com/spf13/cobra"
)

// retentionCmd represents the retention command
var retentionCmd = &cobra.Command{
	Use:   "retention",
	Short: "Subcommand for operations that change retention period",
}

func init() {
	metricsCmd.AddCommand(retentionCmd)
}
