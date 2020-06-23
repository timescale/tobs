package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/spf13/cobra"
)

// helmUninstallCmd represents the helm uninstall command
var helmUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstalls Timescale Observability",
	Args:  cobra.ExactArgs(0),
	RunE:  helmUninstall,
}

func init() {
	helmCmd.AddCommand(helmUninstallCmd)
	helmUninstallCmd.Flags().BoolP("pvc", "", false, "Remove Persistent Volume Claims")
}

func helmUninstall(cmd *cobra.Command, args []string) error {
	var err error

	var name string
	name, err = cmd.Flags().GetString("name")
	if err != nil {
		return fmt.Errorf("could not uninstall Timescale Observability: %w", err)
	}

	var pvc bool
	pvc, err = cmd.Flags().GetBool("pvc")
	if err != nil {
		return fmt.Errorf("could not uninstall Timescale Observability: %w", err)
	}

	var namespace string
	namespace, err = cmd.Flags().GetString("namespace")
	if err != nil {
		return fmt.Errorf("could not uninstall Timescale Observability: %w", err)
	}

	var stdbuf bytes.Buffer
	mw := io.MultiWriter(os.Stdout, &stdbuf)

	uninstall := exec.Command("helm", "uninstall", name)

	uninstall.Stdout = mw
	uninstall.Stderr = mw
	fmt.Println("Uninstalling Timescale Observability")
	err = uninstall.Run()
	if err != nil {
		return fmt.Errorf("could not uninstall Timescale Observability: %w", err)
	}

	fmt.Println("Waiting for pods to terminate...")
	for i := 0; i < 1000; i++ {
		pods, err := KubeGetAllPods(namespace)
		if err != nil {
			return fmt.Errorf("could not uninstall Timescale Observability: %w", err)
		}
		if len(pods) == 0 {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Println("Deleting remaining artifacts")
	err = KubeDeleteService(namespace, name+"-config")
	if err != nil {
		fmt.Println(err, ", skipping")
	}

	err = KubeDeleteEndpoint(namespace, name)
	if err != nil {
		fmt.Println(err, ", skipping")
	}

	if !pvc {
		return nil
	}

	fmt.Println("Getting Persistent Volume Claims")
	pvcnames, err := KubeGetPVCNames(namespace, map[string]string{"release": name})
	if err != nil {
		return fmt.Errorf("could not uninstall Timescale Observability: %w", err)
	}

	fmt.Println("Removing Persistent Volume Claims")
	for _, s := range pvcnames {
		err = KubeDeletePVC(namespace, s)
		if err != nil {
			return fmt.Errorf("could not uninstall Timescale Observability: %w", err)
		}
	}

	return nil
}
