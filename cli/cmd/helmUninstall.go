package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/timescale/tobs/cli/pkg/k8s"

	"github.com/spf13/cobra"
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

	var uninstall *exec.Cmd
	if namespace == "default" {
		uninstall = exec.Command("helm", "uninstall", name)
	} else {
		uninstall = exec.Command("helm", "uninstall", name, "--namespace", namespace)
	}

	uninstall.Stdout = mw
	uninstall.Stderr = mw
	fmt.Println("Uninstalling The Observability Stack")
	err = uninstall.Run()
	if err != nil {
		return fmt.Errorf("could not uninstall The Observability Stack: %w", err)
	}

	fmt.Println("Waiting for pods to terminate...")
	for i := 0; i < 1000; i++ {
		pods, err := k8s.KubeGetAllPods(namespace, name)
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
	err = k8s.KubeDeleteService(namespace, name+"-config")
	if err != nil {
		fmt.Println(err, ", skipping")
	}

	err = k8s.KubeDeleteEndpoint(namespace, name)
	if err != nil {
		fmt.Println(err, ", skipping")
	}

	if deleteData {
		fmt.Println("Checking Persistent Volume Claims")
		pvcnames, err := k8s.KubeGetPVCNames(namespace, map[string]string{"release": name})
		if err != nil {
			return fmt.Errorf("could not uninstall The Observability Stack: %w", err)
		}

		fmt.Println("Removing Persistent Volume Claims")
		for _, s := range pvcnames {
			err = k8s.KubeDeletePVC(namespace, s)
			if err != nil {
				return fmt.Errorf("could not uninstall The Observability Stack: %w", err)
			}
		}
	} else {
		fmt.Println("Data still remains. To delete data as well, run 'tobs helm delete-data'")
	}

	return nil
}
