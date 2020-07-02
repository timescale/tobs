package cmd

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

// helmGetYamlCmd represents the helm get-yaml command
var helmGetYamlCmd = &cobra.Command{
	Use:   "get-yaml",
	Short: "Prints the default Timescale Observability values to console",
	Args:  cobra.ExactArgs(0),
	RunE:  helmGetYaml,
}

func init() {
	helmCmd.AddCommand(helmGetYamlCmd)
}

func helmGetYaml(cmd *cobra.Command, args []string) error {
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
