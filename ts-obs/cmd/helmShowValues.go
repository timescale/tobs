package cmd

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

// helmShowValuesCmd represents the helm show-values command
var helmShowValuesCmd = &cobra.Command{
	Use:   "show-values",
	Short: "Prints the default Timescale Observability values to console",
	Args:  cobra.ExactArgs(0),
	RunE:  helmShowValues,
}

func init() {
	helmCmd.AddCommand(helmShowValuesCmd)
}

func helmShowValues(cmd *cobra.Command, args []string) error {
	var err error

	var getyaml *exec.Cmd
	if DEVEL {
		getyaml = exec.Command("helm", "show", "values", "timescale/timescale-observability", "--devel")
	} else {
		getyaml = exec.Command("helm", "show", "values", "timescale/timescale-observability")
	}

	var out []byte
	out, err = getyaml.CombinedOutput()
	if err != nil {
		return fmt.Errorf("could not get Helm values: %w", err)
	}

	fmt.Print(string(out))

	return nil
}
