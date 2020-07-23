package cmd

import (
	"github.com/spf13/cobra"
)

// chunkIntervalCmd represents the chunk-interval command
var chunkIntervalCmd = &cobra.Command{
	Use:   "chunk-interval",
	Short: "Subcommand for operations that change chunk interval",
}

func init() {
	metricsCmd.AddCommand(chunkIntervalCmd)
}
