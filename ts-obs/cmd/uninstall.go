package cmd

import (
	"github.com/spf13/cobra"
)

// uninstallCmd represents the uninstall command
var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Alias for helm uninstall",
	RunE: uninstall,
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
    uninstallCmd.Flags().StringP("name", "n", "ts-obs", "Release name")
    uninstallCmd.Flags().BoolP("pvc", "", false, "Remove Persistent Volume Claims")
}

func uninstall(cmd *cobra.Command, args []string) error {
   return helmUninstall(cmd, args)
}
