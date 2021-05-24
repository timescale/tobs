package helm

import (
	"github.com/spf13/cobra"
	"github.com/timescale/tobs/cli/cmd"
)

// uninstallCmd represents the uninstall command
var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Alias for helm uninstall",
	Args:  cobra.ExactArgs(0),
	RunE:  uninstall,
}

func init() {
	cmd.RootCmd.AddCommand(uninstallCmd)
	uninstallCmd.Flags().BoolP("delete-data", "", false, "Delete persistent volume claims")
}

func uninstall(cmd *cobra.Command, args []string) error {
	return helmUninstall(cmd, args)
}
