package helm

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/timescale/tobs/cli/cmd"
)

// uninstallCmd represents the uninstall command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Version of tobs",
	Args:  cobra.ExactArgs(0),
	RunE:  version,
}

func init() {
	cmd.RootCmd.AddCommand(versionCmd)
}

const tobsVersion = "0.4.0"

func version(cmd *cobra.Command, args []string) error {
	fmt.Printf("Tobs Version: %s\n", tobsVersion)
	return nil
}

