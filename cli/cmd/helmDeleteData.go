package cmd

import (
	"fmt"

	"github.com/timescale/tobs/cli/pkg/k8s"

	"github.com/spf13/cobra"
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
	pvcnames, err := k8s.KubeGetPVCNames(namespace, map[string]string{"release": name})
	if err != nil {
		return fmt.Errorf("could not delete PVCs: %w", err)
	}

	fmt.Println("Removing Persistent Volume Claims")
	for _, s := range pvcnames {
		err = k8s.KubeDeletePVC(namespace, s)
		if err != nil {
			return fmt.Errorf("could not delete PVCs: %w", err)
		}
	}

	return nil
}
