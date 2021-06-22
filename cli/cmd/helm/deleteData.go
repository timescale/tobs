package helm

import (
	"fmt"

	"github.com/spf13/cobra"
	root "github.com/timescale/tobs/cli/cmd"
)

// helmDeleteDataCmd represents the helm delete-data command
var helmDeleteDataCmd = &cobra.Command{
	Use:   "delete-data",
	Short: "Deletes Persistent Volume Claims",
	Args:  cobra.ExactArgs(0),
	RunE:  helmDeleteData,
}

func init() {
	helmCmd.AddCommand(helmDeleteDataCmd)
}

func helmDeleteData(cmd *cobra.Command, args []string) error {
	var err error

	fmt.Println("Getting Persistent Volume Claims")
	pvcnames, err := kubeClient.KubeGetPVCNames(root.Namespace, map[string]string{"release": root.HelmReleaseName})
	if err != nil {
		return fmt.Errorf("could not delete PVCs: %w", err)
	}

	fmt.Println("Removing Persistent Volume Claims")
	for _, s := range pvcnames {
		err = kubeClient.KubeDeletePVC(root.Namespace, s)
		if err != nil {
			return fmt.Errorf("could not delete PVCs: %w", err)
		}
	}

	return nil
}
