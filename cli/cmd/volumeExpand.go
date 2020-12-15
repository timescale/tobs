package cmd

import (
	"errors"
	"fmt"
	"github.com/timescale/tobs/cli/pkg/k8s"

	"github.com/spf13/cobra"
)

// volumeExpandCmd represents the volume expand command
var volumeExpandCmd = &cobra.Command{
	Use:   "expand",
	Short: "Expand PVC's",
	Args:  cobra.ExactArgs(0),
	RunE:  volumeExpand,
}

func init() {
	volumeCmd.AddCommand(volumeExpandCmd)
	volumeExpandCmd.Flags().StringP("timescaleDB-wal", "w", "", "Expand volume of timescaleDB wal")
	volumeExpandCmd.Flags().StringP("timescaleDB-storage", "s", "", "Expand volume of timescaleDB storage")
	volumeExpandCmd.Flags().StringP("prometheus-storage", "p", "", "Expand volume of prometheus storage")
	volumeExpandCmd.Flags().BoolP("restart-pods", "r", false, "Restarts the pods bound to a PVC on PVC expansion")
}

func volumeExpand(cmd *cobra.Command, args []string) error {
	tsDBWal, err := cmd.Flags().GetString("timescaleDB-wal")
	if err != nil {
		return fmt.Errorf("could not get timescaleDB-wal flag %w", err)
	}

	tsDBStorage, err := cmd.Flags().GetString("timescaleDB-storage")
	if err != nil {
		return fmt.Errorf("could not get timescaleDB-storage flag %w", err)
	}

	promStorage, err := cmd.Flags().GetString("prometheus-storage")
	if err != nil {
		return fmt.Errorf("could not get prometheus-storage flag %w", err)
	}

	restartsPods, err := cmd.Flags().GetBool("restart-pods")
	if err != nil {
		return fmt.Errorf("could not get restart-pods flag %w", err)
	}

	var timescaleDBPodLabels = map[string]string{
		"app": name+"-timescaledb",
		"release": name,
	}

	var prometheusPodLabels = map[string]string{
		"app": "prometheus",
		"component": "server",
		"release": name,
	}

	if promStorage == "" && tsDBStorage == "" && tsDBWal == "" {
		return errors.New("use resource specific flag and provide the desired size for pvc expansion")
	}

	if tsDBStorage != "" {
		pvcPrefix := "storage-volume"
		results, err := k8s.ExpandTimescaleDBPVC(namespace,  tsDBStorage, pvcPrefix , timescaleDBPodLabels)
		if err != nil {
			return fmt.Errorf("could not expand timescaleDB-storage: %w", err)
		}

		expandSuccessPrint(pvcPrefix, results)

		if restartsPods {
			err = restartPods(timescaleDBPodLabels)
			if err != nil {
				return err
			}
		}
	}

	if tsDBWal != "" {
		pvcPrefix := "wal-volume"
		results, err := k8s.ExpandTimescaleDBPVC(namespace, tsDBWal, pvcPrefix, timescaleDBPodLabels)
		if err != nil {
			return fmt.Errorf("could not expand timescaleDB-wal: %w", err)
		}

		expandSuccessPrint(pvcPrefix, results)

		if restartsPods {
			err = restartPods(timescaleDBPodLabels)
			if err != nil {
				return err
			}
		}
	}

	if promStorage != "" {
		pvcPrefix := name+"-prometheus-server"
		err := k8s.ExpandPVC(namespace, pvcPrefix, promStorage)
		if err != nil {
			return fmt.Errorf("could not expand prometheus-storage: %w", err)
		}
		expandSuccessPrint(pvcPrefix, map[string]string{pvcPrefix: promStorage})

		if restartsPods {
			err = restartPods(prometheusPodLabels)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func expandSuccessPrint(pvcPrefix string, results map[string]string) {
	if len(results) == 0 {
		return
	}

	fmt.Printf("PVC's of %s\n", pvcPrefix)
	for pvcName, value := range results {
		fmt.Printf("Successfully expanded PVC: %s to %s\n", pvcName, value)
	}
	fmt.Println()
}

func restartPods(labels map[string]string) error {
	err := k8s.DeletePods(namespace, labels)
	if err != nil {
		return fmt.Errorf("failed to restart pods after PVC expansion: %w", err)
	}
	fmt.Println("Triggered to restart the pods bound by the PVC's.")
	return nil
}