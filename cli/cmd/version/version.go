package version

import (
	"fmt"

	"github.com/spf13/cobra"
	root "github.com/timescale/tobs/cli/cmd"
	"github.com/timescale/tobs/cli/pkg/helm"
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

const tobsVersion = "0.5.0"

func version(cmd *cobra.Command, args []string) error {
	d, err := cmd.Flags().GetBool("deployed-chart")
	if err != nil {
		return fmt.Errorf("could not get deployed tobs helm chart version: %w", err)
	}

	var chartVersion string
	helmClient := helm.NewClient(root.Namespace)
	defer helmClient.Close()
	if d {
		deployedChart, err := helmClient.GetDeployedChartMetadata(root.HelmReleaseName)
		if err != nil {
			chartVersion = fmt.Errorf("failed to get the deployed chart version: %v", err).Error()
		} else {
			chartVersion = fmt.Sprintf("deployed tobs helm chart version: %s", deployedChart.Version)
		}
	} else {
		err = helmClient.AddOrUpdateChartRepo(utils.DEFAULT_REGISTRY_NAME, utils.REPO_LOCATION)
		if err != nil {
			return fmt.Errorf("failed to add and update the tobs helm chart version %v", err)
		}
		latestChart, err := helmClient.GetChartMetadata(utils.DEFAULT_CHART)
		if err != nil {
			return fmt.Errorf("failed to get latest tobs helm chart version %v", err)
		}
		chartVersion = fmt.Sprintf("latest tobs helm chart version: %s", latestChart.Version)
	}

	fmt.Printf("Tobs CLI Version: %s, %s \n", tobsVersion, chartVersion)
	return nil
}
