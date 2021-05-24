package helm

import (
	"github.com/spf13/cobra"
	"github.com/timescale/tobs/cli/cmd"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Alias for helm install",
	Args:  cobra.ExactArgs(0),
	RunE:  install,
}

func init() {
	cmd.RootCmd.AddCommand(installCmd)
	addChartDetailsFlags(installCmd)
	addInstallUtilitiesFlags(installCmd)
}

func install(cmd *cobra.Command, args []string) error {
	return helmInstall(cmd, args)
}
