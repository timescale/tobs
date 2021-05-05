package helm

import (
	"fmt"
	"github.com/spf13/cobra"
	root "github.com/timescale/tobs/cli/cmd"
	"github.com/timescale/tobs/cli/pkg/utils"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Version of tobs",
	Args:  cobra.ExactArgs(0),
	RunE:  version,
}

func init() {
	root.RootCmd.AddCommand(versionCmd)
	versionCmd.Flags().BoolP("deployed-chart", "d", false, "Option to show deployed tobs helm chart version")
}

const tobsVersion = "0.4.0"

func version(cmd *cobra.Command, args []string) error {
	d, err := cmd.Flags().GetBool("deployed-chart")
	if err != nil {
		return fmt.Errorf("could not get deployed tobs helm chart version: %w", err)
	}

	var chartVersion string
	if d {
		deployedChart, err := utils.GetDeployedChartMetadata(root.HelmReleaseName, root.Namespace)
		if err != nil {
			chartVersion = fmt.Errorf("failed to get the deployed chart version: %v", err).Error()
		} else {
			chartVersion = fmt.Sprintf("deployed tobs helm chart version: %s", deployedChart.Version)
		}
	} else {
		err = utils.AddUpdateTobsChart(false)
		if err != nil {
			return fmt.Errorf("failed to add and update the tobs helm chart version %v", err)
		}
		latestChart, err := utils.GetTobsChartMetadata(utils.DEFAULT_CHART)
		if err != nil {
			return fmt.Errorf("failed to get latest tobs helm chart version %v", err)
		}
		chartVersion = fmt.Sprintf("latest tobs helm chart version: %s", latestChart.Version)
	}

	fmt.Printf("Tobs CLI Version: %s, %s \n", tobsVersion, chartVersion)
	return nil
}
