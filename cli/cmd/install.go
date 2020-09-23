package cmd

import (
	"github.com/spf13/cobra"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Alias for helm install",
	Args:  cobra.ExactArgs(0),
	RunE:  install,
}

func init() {
	rootCmd.AddCommand(installCmd)
	addHelmInstallFlags(installCmd)
}

func install(cmd *cobra.Command, args []string) error {
	return helmInstall(cmd, args)
}
