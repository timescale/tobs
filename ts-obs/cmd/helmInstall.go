package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/spf13/cobra"
)

const DEVEL = true

// helmInstallCmd represents the helm install command
var helmInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Installs Timescale Observability",
	Args:  cobra.ExactArgs(0),
	RunE:  helmInstall,
}

func init() {
	helmCmd.AddCommand(helmInstallCmd)
	helmInstallCmd.Flags().StringP("filename", "f", "", "YAML configuration file to load")
}

func helmInstall(cmd *cobra.Command, args []string) error {
	var err error

	var name string
	name, err = cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	var file string
	file, err = cmd.Flags().GetString("filename")
	if err != nil {
		return err
	}

	var namespace string
	namespace, err = cmd.Flags().GetString("namespace")
	if err != nil {
		return err
	}

	w := io.Writer(os.Stdout)

	addchart := exec.Command("helm", "repo", "add", "timescale", "https://charts.timescale.com")

	addchart.Stdout = w
	addchart.Stderr = w
	fmt.Println("Adding Timescale Helm Repository")
	err = addchart.Run()
	if err != nil {
		return err
	}

	update := exec.Command("helm", "repo", "update")

	update.Stdout = w
	update.Stderr = w
    fmt.Println("Fetching updates from repository")
	err = update.Run()
	if err != nil {
		return err
	}

	var install *exec.Cmd
	if DEVEL {
		if file == "" {
			install = exec.Command("helm", "install", name, "timescale/timescale-observability", "--devel")
		} else {
			install = exec.Command("helm", "upgrade", "--install", name, "--values", file, "timescale/timescale-observability", "--devel")
		}
	} else {
		if file == "" {
			install = exec.Command("helm", "install", name, "timescale/timescale-observability")
		} else {
			install = exec.Command("helm", "upgrade", "--install", name, "--values", file, "timescale/timescale-observability")
		}
	}

	install.Stdout = w
	install.Stderr = w
	fmt.Println("Installing Timescale Observability")
	err = install.Run()
	if err != nil {
		return err
	}

	time.Sleep(10 * time.Second)

	fmt.Println("Waiting for pods to initialize...")
	pods, err := KubeGetAllPods(namespace)
	if err != nil {
		return err
	}

	for _, pod := range pods {
		err = KubeWaitOnPod(namespace, pod.Name)
		if err != nil {
			return err
		}
	}

	return nil
}
