package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// helmShowValuesCmd represents the helm show-values command
var helmShowValuesCmd = &cobra.Command{
	Use:   "show-values",
	Short: "Prints the default Observability Stack values to console",
	Args:  cobra.ExactArgs(0),
	RunE:  helmShowValues,
}

func init() {
	helmCmd.AddCommand(helmShowValuesCmd)
	addHelmInstallFlags(helmShowValuesCmd)
}

func helmShowValues(cmd *cobra.Command, args []string) error {
	var err error

	chart, err := cmd.Flags().GetString("chart-reference")
	if err != nil {
		return fmt.Errorf("could not install The Observability Stack: %w", err)
	}

	var showvalues *exec.Cmd
	if DEVEL {
		showvalues = exec.Command("helm", "show", "values", chart, "--devel")
	} else {
		showvalues = exec.Command("helm", "show", "values", chart)
	}
	showvalues.Stderr = os.Stderr

	var out []byte
	out, err = showvalues.Output()
	if err != nil {
		return fmt.Errorf("could not get Helm values: %w", err)
	}

	fmt.Print(string(out))

	return nil
}
