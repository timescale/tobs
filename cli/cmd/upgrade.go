package cmd

import (
	"errors"
	"fmt"
	"os/exec"
	"reflect"
	"strings"

	"github.com/spf13/cobra"
	"github.com/timescale/tobs/cli/pkg/utils"
)

// upgradeCmd represents the upgrade command
var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Alias for helm upgrade",
	Args:  cobra.ExactArgs(0),
	RunE:  upgrade,
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
	addHelmInstallFlags(upgradeCmd)
	upgradeCmd.Flags().BoolP("reset-values", "", false, "Reset helm chart to default values of the helm chart. This is same flag that exists in helm upgrade")
	upgradeCmd.Flags().BoolP("reuse-values", "", false, "Reuse the last release's values and merge in any overrides from the command line via --set and -f. If '--reset-values' is specified, this is ignored.\nThis is same flag that exists in helm upgrade ")
	upgradeCmd.Flags().BoolP("same-chart", "", false, "Use the same helm chart do not upgrade helm chart but upgrade the existing chart with new values")
	upgradeCmd.Flags().BoolP("confirm", "y", false, "Confirmation flag for upgrading")
}

func upgrade(cmd *cobra.Command, args []string) error {
	return upgradeTobs(cmd, args)
}

func upgradeTobs(cmd *cobra.Command, args []string) error {
	file, err := cmd.Flags().GetString("filename")
	if err != nil {
		return fmt.Errorf("couldn't get the filename flag value: %w", err)
	}

	ref, err := cmd.Flags().GetString("chart-reference")
	if err != nil {
		return fmt.Errorf("couldn't get the chart-reference flag value: %w", err)
	}

	reset, err := cmd.Flags().GetBool("reset-values")
	if err != nil {
		return fmt.Errorf("couldn't get the reset-values flag value: %w", err)
	}

	reuse, err := cmd.Flags().GetBool("reuse-values")
	if err != nil {
		return fmt.Errorf("couldn't get the reuse-values flag value: %w", err)
	}

	confirm, err := cmd.Flags().GetBool("confirm")
	if err != nil {
		return fmt.Errorf("couldn't get the confirm flag value: %w", err)
	}

	sameChart, err := cmd.Flags().GetBool("same-chart")
	if err != nil {
		return fmt.Errorf("couldn't get the reuse-values flag value: %w", err)
	}

	cmds := []string{"upgrade", name, ref, "--namespace", namespace}

	if file != "" {
		cmds = append(cmds, "--values", file)
	}

	if reset {
		cmds = append(cmds, "--reset-values")
	}

	if reuse {
		cmds = append(cmds, "--reuse-values")
	}

	latestChart, err := utils.GetTobsChartMetadata(ref)
	if err != nil {
		return err
	}

	deployedChart, err := utils.GetDeployedChartMetadata(name, namespace)
	if err != nil {
		return err
	}

	if deployedChart.Name == "" {
		fmt.Println("couldn't find the existing tobs deployment. Deploying tobs...")
		if !confirm {
			utils.ConfirmAction()
		}
		err = installStack(file, ref, "")
		if err != nil {
			return err
		}
		return nil
	}

	// add & update helm chart only if it's default chart
	// if same-chart upgrade is disabled
	if ref == utils.DEFAULT_CHART && !sameChart {
		err = utils.AddTobsHelmChart()
		if err != nil {
			return err
		}

		err = utils.UpdateTobsHelmChart(true)
		if err != nil {
			return err
		}
	}

	lVersion, err := utils.ParseVersion(latestChart.Version, 3)
	if err != nil {
		return fmt.Errorf("failed to parse latest helm chart version %w", err)
	}

	s := strings.Split(deployedChart.Chart, "-")
	dVersion, err := utils.ParseVersion(s[1], 3)
	if err != nil {
		return fmt.Errorf("failed to parse deployed helm chart version %w", err)
	}

	var foundNewChart bool
	if lVersion <= dVersion {
		dValues, err := utils.DeployedValuesYaml(ref, name, namespace)
		if err != nil {
			return err
		}

		nValues, err := utils.NewValuesYaml(ref, file)
		if err != nil {
			return err
		}

		if ok := reflect.DeepEqual(dValues, nValues); ok {
			err = errors.New("failed to upgrade there is no latest helm chart available and existing helm deployment values are same as the provided values")
			return err
		}
	} else {
		foundNewChart = true
		if sameChart {
			err = errors.New("provided helm chart is newer compared to existing deployed helm chart cannot upgrade as --same-chart flag is provided")
			return err
		}
	}

	if foundNewChart {
		fmt.Printf("Upgrading to latest helm chart version: %s\n", latestChart.Version)
	} else {
		fmt.Println("Upgrading the existing helm chart with values.yaml file")
	}

	if !confirm {
		utils.ConfirmAction()
	}

	out := exec.Command("helm", cmds...)
	k, err := out.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to upgrade %v", string(k))
	}

	fmt.Printf("Successfully upgraded %s.\n", name)
	return nil
}
