package uninstall

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	root "github.com/timescale/tobs/cli/cmd"
	"github.com/timescale/tobs/cli/pkg/helm"
	"github.com/timescale/tobs/cli/pkg/k8s"
	"github.com/timescale/tobs/cli/pkg/otel"
	"github.com/timescale/tobs/cli/pkg/utils"
	k8sApiErrors "k8s.io/apimachinery/pkg/api/errors"
)

// helmUninstallCmd represents the helm uninstall command
var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstalls The Observability Stack",
	Args:  cobra.ExactArgs(0),
	RunE:  helmUninstall,
}

func init() {
	root.RootCmd.AddCommand(uninstallCmd)
	uninstallCmd.Flags().BoolP("delete-data", "", false, "Delete persistent volume claims")
}

func helmUninstall(cmd *cobra.Command, args []string) error {
	var err error

	var deleteData bool
	deleteData, err = cmd.Flags().GetBool("delete-data")
	if err != nil {
		return fmt.Errorf("could not uninstall The Observability Stack: %w", err)
	}

	helmClient := helm.NewClient(root.Namespace)
	defer helmClient.Close()

	fmt.Println("Uninstalling The Observability Stack")

	k8sClient := k8s.NewClient()
	// If chart is upgraded to 0.4.0 & performing uninstall
	// we should manually delete the 0.4.0 upgrade job
	err = delete040UpgradeJob(helmClient, k8sClient)
	if err != nil {
		return err
	}

	spec := &helm.ChartSpec{
		ReleaseName: root.HelmReleaseName,
		Namespace:   root.Namespace,
	}

	otelCol := otel.OtelCol{
		ReleaseName: root.HelmReleaseName,
		Namespace:   root.Namespace,
		K8sClient:   k8sClient,
		HelmClient:  helmClient,
	}
	isOtelEnabled, err := otelCol.IsOtelOperatorEnabledInRelease()
	if err != nil {
		return err
	}

	if isOtelEnabled {
		// delete default otel collector
		fmt.Println("deleting the default opentelemetry collector...")
		err = otelCol.DeleteDefaultOtelCollector()
		if err != nil {
			if !k8sApiErrors.IsNotFound(err) {
				return err
			}
			fmt.Println("couldn't find the default otel collector.")
		}

		certManagerInstalled, err := otelCol.IsCertManagerInstalledByTobs()
		if err != nil {
			fmt.Printf("ERROR: couldn't find the cert-manager status %v", err)
		}
		if certManagerInstalled {
			cmVersion, err := otelCol.GetCertManagerVersion()
			if err != nil {
				return err
			}
			fmt.Printf(`
WARNING: Cert-manager, version: %s  is installed by tobs, 
As cert-manager can be used by other components running in the cluster.
We are leaving the task of uninstalling cert-manager to the user.
Steps to uninstall cert-manager is documented here: https://cert-manager.io/docs/installation/kubectl/#uninstalling

`, cmVersion)
		}
	}

	helmClient = helm.NewClient(root.Namespace)
	err = helmClient.UninstallRelease(spec)
	if err != nil {
		return fmt.Errorf("could not uninstall The Observability Stack: %w", err)
	}
	fmt.Println("Waiting for pods to terminate...")
	for i := 0; i < 1000; i++ {
		pods, err := k8sClient.KubeGetAllPods(root.Namespace, root.HelmReleaseName)
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

	// As helm uninstall doesn't remove CRD, we are manually deleting
	// the otelcol CRD post helm successful uninstall by now otel operator is down
	if isOtelEnabled {
		fmt.Println("Deleting opentelemetry collector CRD...")
		if err = otel.DeleteOtelColCRD(); err != nil {
			return err
		}
	}

	fmt.Println("Deleting remaining artifacts")
	err = k8sClient.KubeDeleteService(root.Namespace, root.HelmReleaseName+"-config")
	if err != nil {
		fmt.Println(err, ", skipping")
	}

	err = k8sClient.KubeDeleteEndpoint(root.Namespace, root.HelmReleaseName)
	if err != nil {
		fmt.Println(err, ", skipping")
	}

	err = k8sClient.DeleteSecret("tobs-kube-prometheus-admission", root.Namespace)
	if err != nil {
		fmt.Println(err, ", failed to delete kube-prometheus-admission secret")
	}

	if deleteData {
		err = deletePVCData(&cobra.Command{}, []string{})
		if err != nil {
			fmt.Println(err, ", failed to delete pvc's")
		}
	} else {
		fmt.Println("Data still remains. To delete data as well, run 'tobs uninstall delete-data'")
	}

	return nil
}

func delete040UpgradeJob(helmClient helm.Client, k8sClient k8s.Client) error {
	deployedChart, err := helmClient.GetDeployedChartMetadata(root.HelmReleaseName, root.Namespace)
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
		upgradeJob, err := k8sClient.GetJob(utils.UpgradeJob_040, root.Namespace)
		if err != nil {
			ok := k8sApiErrors.IsNotFound(err)
			if !ok {
				return fmt.Errorf("failed to delete %s job %v", utils.UpgradeJob_040, err)
			}
		}

		if upgradeJob.Name != "" {
			fmt.Println("deleting the 0.4.0 upgrade job...")
			err = k8sClient.DeleteJob(utils.UpgradeJob_040, root.Namespace)
			if err != nil {
				return fmt.Errorf("failed to delete job %s %v\n", utils.UpgradeJob_040, err)
			}
		}
	}
	return nil
}
