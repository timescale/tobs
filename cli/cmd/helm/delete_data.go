package helm

import (
	"fmt"

	"github.com/spf13/cobra"
	root "github.com/timescale/tobs/cli/cmd"
	"github.com/timescale/tobs/cli/cmd/common"
	"github.com/timescale/tobs/cli/pkg/k8s"
)

// helmDeleteDataCmd represents the helm delete-data command
var helmDeleteDataCmd = &cobra.Command{
	Use:   "delete-data",
	Short: "Deletes Persistent Volume Claims",
	Args:  cobra.ExactArgs(0),
	RunE:  deletePVCData,
}

func init() {
	helmCmd.AddCommand(helmDeleteDataCmd)
}

func deletePVCData(cmd *cobra.Command, args []string) error {
	var err error

	fmt.Println("Getting Persistent Volume Claims")
	k8sClient := k8s.NewClient()
	pvcnames, err := k8sClient.KubeGetPVCNames(root.Namespace, map[string]string{"release": root.HelmReleaseName})
	if err != nil {
		return fmt.Errorf("could not delete PVCs: %w", err)
	}

	// Prometheus PVC's doesn't hold the release labelSet
	prometheusPvcNames, err := k8sClient.KubeGetPVCNames(root.Namespace, common.GetPrometheusLabels())
	if err != nil {
		return fmt.Errorf("could not uninstall The Observability Stack: %w", err)
	}
	pvcnames = append(pvcnames, prometheusPvcNames...)

	fmt.Println("Removing Persistent Volume Claims")
	for _, s := range pvcnames {
		err = k8sClient.KubeDeletePVC(root.Namespace, s)
		if err != nil {
			return fmt.Errorf("could not delete PVCs: %w", err)
		}
	}

	return nil
}
