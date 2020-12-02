package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/timescale/tobs/cli/pkg/k8s"
)

// volumeGetCmd represents the volume expand command
var volumeGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get PVC's volume",
	Args:  cobra.ExactArgs(0),
	RunE:  volumeGet,
}

func init() {
	volumeCmd.AddCommand(volumeGetCmd)
	volumeGetCmd.Flags().BoolP("timescaleDB-wal", "w", false, "Get volume of timescaleDB wal")
	volumeGetCmd.Flags().BoolP("timescaleDB-storage", "s", false, "Get volume of timescaleDB storage")
	volumeGetCmd.Flags().BoolP("prometheus-storage", "p", false, "Get volume of prometheus storage")
}

func volumeGet(cmd *cobra.Command, args []string) error {
	tsDBWal, err := cmd.Flags().GetBool("timescaleDB-wal")
	if err != nil {
		return fmt.Errorf("could not get timescaleDB-wal flag %w", err)
	}

	tsDBStorage, err := cmd.Flags().GetBool("timescaleDB-storage")
	if err != nil {
		return fmt.Errorf("could not get timescaleDB-storage flag %w", err)
	}

	promStorage, err := cmd.Flags().GetBool("prometheus-storage")
	if err != nil {
		return fmt.Errorf("could not get prometheus-storage flag %w", err)
	}

	if tsDBStorage {
		pvcPrefix := "storage-volume"
		results, err := k8s.GetPVCSizes(namespace, pvcPrefix, map[string]string{"app": name+"-timescaledb"})
		if err != nil {
			return fmt.Errorf("could not get timescaleDB-storage: %w", err)
		}
		volumeGetPrint(pvcPrefix, results)
	}

	if tsDBWal {
		pvcPrefix := "wal-volume"
		results, err := k8s.GetPVCSizes(namespace, pvcPrefix, map[string]string{"app": name+"-timescaledb"})
		if err != nil {
			return fmt.Errorf("could not get timescaleDB-wal: %w", err)
		}
		volumeGetPrint(pvcPrefix, results)
	}

	if promStorage {
		pvcPrefix := name+"-prometheus-server"
		results, err := k8s.GetPVCSizes(namespace, pvcPrefix, nil)
		if err != nil {
			return fmt.Errorf("could not get prometheus-storage: %w", err)
		}
		volumeGetPrint(pvcPrefix, results)
	}

	return nil
}

func volumeGetPrint(pvcPrefix string, results map[string]string) {
	if len(results) == 0 {
		return
	}

	fmt.Printf("PVC's of %s\n", pvcPrefix)
	for pvcName, value := range results {
		fmt.Printf("Existing size of PVC: %s is %s\n", pvcName, value)
	}
	fmt.Println()
}