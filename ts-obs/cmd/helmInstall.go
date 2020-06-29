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
const REPO_LOCATION = "https://charts.timescale.com"

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

	var file string
	file, err = cmd.Flags().GetString("filename")
	if err != nil {
		return fmt.Errorf("could not install Timescale Observability: %w", err)
	}

	w := io.Writer(os.Stdout)

	addchart := exec.Command("helm", "repo", "add", "timescale", REPO_LOCATION)

	addchart.Stdout = w
	addchart.Stderr = w
	fmt.Println("Adding Timescale Helm Repository")
	err = addchart.Run()
	if err != nil {
		return fmt.Errorf("could not install Timescale Observability: %w", err)
	}

	update := exec.Command("helm", "repo", "update")

	update.Stdout = w
	update.Stderr = w
	fmt.Println("Fetching updates from repository")
	err = update.Run()
	if err != nil {
		return fmt.Errorf("could not install Timescale Observability: %w", err)
	}

	var install *exec.Cmd
	if DEVEL {
		if namespace == "default" {
			if file == "" {
				install = exec.Command("helm", "install", name, "timescale/timescale-observability", "--devel")
			} else {
				install = exec.Command("helm", "install", name, "--values", file, "timescale/timescale-observability", "--devel")
			}
		} else {
			if file == "" {
				install = exec.Command("helm", "install", name, "timescale/timescale-observability", "--create-namespace", "-n", namespace, "--devel")
			} else {
				install = exec.Command("helm", "install", name, "--values", file, "timescale/timescale-observability", "--create-namespace", "-n", namespace, "--devel")
			}
		}
	} else {
		if namespace == "default" {
			if file == "" {
				install = exec.Command("helm", "install", name, "timescale/timescale-observability")
			} else {
				install = exec.Command("helm", "install", name, "--values", file, "timescale/timescale-observability")
			}
		} else {
			if file == "" {
				install = exec.Command("helm", "install", name, "timescale/timescale-observability", "--create-namespace", "-n", namespace)
			} else {
				install = exec.Command("helm", "install", name, "--values", file, "timescale/timescale-observability", "--create-namespace", "-n", namespace)
			}
		}
	}

	fmt.Println("Installing Timescale Observability")
    out, err := install.CombinedOutput()
	if err != nil {
		return fmt.Errorf("could not install Timescale Observability: %w", err)
	}

	time.Sleep(10 * time.Second)

	fmt.Println("Waiting for pods to initialize...")
	pods, err := KubeGetAllPods(namespace, name)
	if err != nil {
		return err
	}

	for _, pod := range pods {
		err = KubeWaitOnPod(namespace, pod.Name)
		if err != nil {
			return err
		}
	}

	fmt.Println("Timescale Observability has been installed successfully")
    fmt.Println(string(out))
	return nil
}
