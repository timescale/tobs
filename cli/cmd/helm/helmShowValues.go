package helm

import (
	"fmt"
	"github.com/spf13/cobra"
	root "github.com/timescale/tobs/cli/cmd"
	"github.com/timescale/tobs/cli/pkg/helm"
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
	addChartDetailsFlags(helmShowValuesCmd)
}

func helmShowValues(cmd *cobra.Command, args []string) error {
	var err error

	chart, err := cmd.Flags().GetString("chart-reference")
	if err != nil {
		return fmt.Errorf("could not install The Observability Stack: %w", err)
	}

	helmClient = helm.NewClient(root.Namespace)
	res, err := helmClient.GetChartValues(chart)
	if err != nil {
		return fmt.Errorf("failed to get helm values: %w", err)
	}

	fmt.Println(string(res))

	return nil
}
