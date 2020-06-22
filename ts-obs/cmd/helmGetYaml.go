package cmd

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

// helmGetYamlCmd represents the helm get-yaml command
var helmGetYamlCmd = &cobra.Command{
	Use:   "get-yaml",
	Short: "Prints the default timescale-obserability values to console",
	Args:  cobra.ExactArgs(0),
	RunE:  helmGetYaml,
}

func init() {
	helmCmd.AddCommand(helmGetYamlCmd)
}

func helmGetYaml(cmd *cobra.Command, args []string) error {
	var err error

	getyaml := exec.Command("helm", "show", "values", "timescale/timescale-observability", "--devel")

	var out []byte
	out, err = getyaml.CombinedOutput()
	if err != nil {
		return err
	}

	fmt.Print(string(out))

	return nil
}
