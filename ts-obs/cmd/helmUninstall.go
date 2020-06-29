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
}

func helmUninstall(cmd *cobra.Command, args []string) error {
	var err error

	var stdbuf bytes.Buffer
	mw := io.MultiWriter(os.Stdout, &stdbuf)

	var uninstall *exec.Cmd
	if namespace == "default" {
		uninstall = exec.Command("helm", "uninstall", name)
	} else {
		uninstall = exec.Command("helm", "uninstall", name, "-n", namespace)
	}

	uninstall.Stdout = mw
	uninstall.Stderr = mw
	fmt.Println("Uninstalling Timescale Observability")
	err = uninstall.Run()
	if err != nil {
		return fmt.Errorf("could not uninstall Timescale Observability: %w", err)
	}

	fmt.Println("Waiting for pods to terminate...")
	for i := 0; i < 1000; i++ {
		pods, err := KubeGetAllPods(namespace, name)
		if err != nil {
			return fmt.Errorf("could not uninstall Timescale Observability: %w", err)
		}
		if len(pods) == 0 {
			break
		} else if i == 999 {
			fmt.Println("WARNING: pods did not terminate in 100 seconds")
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

	fmt.Println("Data was not deleted. To delete data as well, run 'ts-obs helm delete-data'")

	return nil
}
