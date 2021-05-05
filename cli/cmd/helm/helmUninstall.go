package helm

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/spf13/cobra"
	root "github.com/timescale/tobs/cli/cmd"
	"github.com/timescale/tobs/cli/pkg/k8s"
	"github.com/timescale/tobs/cli/pkg/timescaledb_secrets"
	"github.com/timescale/tobs/cli/pkg/utils"
	errors2 "k8s.io/apimachinery/pkg/api/errors"
)

// helmUninstallCmd represents the helm uninstall command
var helmUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstalls The Observability Stack",
	Args:  cobra.ExactArgs(0),
	RunE:  helmUninstall,
}

func init() {
	helmCmd.AddCommand(helmUninstallCmd)
	helmUninstallCmd.Flags().BoolP("delete-data", "", false, "Delete persistent volume claims")
}

func helmUninstall(cmd *cobra.Command, args []string) error {
	var err error

	var deleteData bool
	deleteData, err = cmd.Flags().GetBool("delete-data")
	if err != nil {
		return fmt.Errorf("could not uninstall The Observability Stack: %w", err)
	}

	var stdbuf bytes.Buffer
	mw := io.MultiWriter(os.Stdout, &stdbuf)

	e, err := utils.ExportValuesFieldFromRelease(root.HelmReleaseName, root.Namespace, TimescaleDBBackUpKeyForValuesYaml)
	if err != nil {
		return err
	}

	enableBackUp, ok := e.(bool)
	if !ok {
		return fmt.Errorf("enable Backup was not a bool")
	}

	var uninstall *exec.Cmd
	helmArgs := []string{"uninstall", root.HelmReleaseName}
	if root.Namespace != "default" {
		helmArgs = append(helmArgs, "--namespace", root.Namespace)
	}
	uninstall = exec.Command("helm", helmArgs...)

	uninstall.Stdout = mw
	uninstall.Stderr = mw
	fmt.Println("Uninstalling The Observability Stack")

	// If chart is upgraded to 0.4.0 & performing uninstall
	// we should manually delete the 0.4.0 upgrade job
	err = delete040UpgradeJob()
	if err != nil {
		return err
	}

	err = uninstall.Run()
	if err != nil {
		return fmt.Errorf("could not uninstall The Observability Stack: %w", err)
	}

	timescaledb_secrets.DeleteTimescaleDBSecrets(root.HelmReleaseName, root.Namespace, enableBackUp)
	fmt.Println("Waiting for pods to terminate...")
	for i := 0; i < 1000; i++ {
		pods, err := k8s.KubeGetAllPods(root.Namespace, root.HelmReleaseName)
		if err != nil {
			return fmt.Errorf("could not uninstall The Observability Stack: %w", err)
		}
		if len(pods) == 0 {
			break
		} else if i == 999 {
			fmt.Println("WARNING: pods did not terminate in 100 seconds")
		}
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Println("Deleting remaining artifacts")
	err = k8s.KubeDeleteService(root.Namespace, root.HelmReleaseName+"-config")
	if err != nil {
		fmt.Println(err, ", skipping")
	}

	err = k8s.KubeDeleteEndpoint(root.Namespace, root.HelmReleaseName)
	if err != nil {
		fmt.Println(err, ", skipping")
	}

	if deleteData {
		fmt.Println("Checking Persistent Volume Claims")
		pvcnames, err := k8s.KubeGetPVCNames(root.Namespace, map[string]string{"release": root.HelmReleaseName})
		if err != nil {
			return fmt.Errorf("could not uninstall The Observability Stack: %w", err)
		}

		fmt.Println("Removing Persistent Volume Claims")
		for _, s := range pvcnames {
			err = k8s.KubeDeletePVC(root.Namespace, s)
			if err != nil {
				return fmt.Errorf("could not uninstall The Observability Stack: %w", err)
			}
		}

	} else {
		fmt.Println("Data still remains. To delete data as well, run 'tobs helm delete-data'")
	}

	return nil
}

func delete040UpgradeJob() error {
	deployedChart, err := utils.GetDeployedChartMetadata(root.HelmReleaseName, root.Namespace)
	if err != nil {
		return err
	}

	dVersion, err := utils.ParseVersion(deployedChart.Version, 3)
	if err != nil {
		fmt.Printf("failed to parse version %v\n", err)
	}
	version0_4_0, err := utils.ParseVersion(utils.Version_040, 3)
	if err != nil {
		return fmt.Errorf("failed to parse 0.4.0 version %w", err)
	}
	if dVersion >= version0_4_0 {
		upgradeJob, err := k8s.GetJob(utils.UpgradeJob_040, root.Namespace)
		if err != nil {
			ok := errors2.IsNotFound(err)
			if !ok {
				return fmt.Errorf("failed to delete %s job %v", utils.UpgradeJob_040, err)
			}
		}

		if upgradeJob.Name != "" {
			fmt.Println("deleting the 0.4.0 upgrade job...")
			err = k8s.DeleteJob(utils.UpgradeJob_040, root.Namespace)
			if err != nil {
				return fmt.Errorf("failed to delete job %s %v\n", utils.UpgradeJob_040, err)
			}
		}
	}
	return nil
}
