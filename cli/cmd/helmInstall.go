package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/timescale/tobs/cli/pkg/k8s"

	"github.com/spf13/cobra"
)

const DEVEL = false
const REPO_LOCATION = "https://charts.timescale.com"

// helmInstallCmd represents the helm install command
var helmInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Installs The Observability Stack",
	Args:  cobra.ExactArgs(0),
	RunE:  helmInstall,
}

func init() {
	helmCmd.AddCommand(helmInstallCmd)
	addHelmInstallFlags(helmInstallCmd)
}

func addHelmInstallFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("filename", "f", "", "YAML configuration file to load")
	cmd.Flags().StringP("chart-reference", "c", "timescale/tobs", "Helm chart reference")
}

func helmInstall(cmd *cobra.Command, args []string) error {
	var err error

	var file, ref string
	file, err = cmd.Flags().GetString("filename")
	if err != nil {
		return fmt.Errorf("could not install The Observability Stack: %w", err)
	}
	ref, err = cmd.Flags().GetString("chart-reference")
	if err != nil {
		return fmt.Errorf("could not install The Observability Stack: %w", err)
	}

	w := io.Writer(os.Stdout)

	addchart := exec.Command("helm", "repo", "add", "timescale", REPO_LOCATION)

	addchart.Stdout = w
	addchart.Stderr = w
	fmt.Println("Adding Timescale Helm Repository")
	err = addchart.Run()
	if err != nil {
		return fmt.Errorf("could not install The Observability Stack: %w", err)
	}

	update := exec.Command("helm", "repo", "update")

	update.Stdout = w
	update.Stderr = w
	fmt.Println("Fetching updates from repository")
	err = update.Run()
	if err != nil {
		return fmt.Errorf("could not install The Observability Stack: %w", err)
	}

	cmds := []string{"install", name, ref, "--set", "cli=true"}
	if namespace != "default" {
		cmds = append(cmds, "--create-namespace", "--namespace", namespace)
	}
	if file != "" {
		cmds = append(cmds, "--values", file)
	}
	if DEVEL {
		cmds = append(cmds, "--devel")
	}

	install := exec.Command("helm", cmds...)
	fmt.Println("Installing The Observability Stack")
	out, err := install.CombinedOutput()
	if err != nil {
		return fmt.Errorf("could not install The Observability Stack: %w \nOutput: %v", err, string(out))
	}

	time.Sleep(10 * time.Second)

	fmt.Println("Waiting for pods to initialize...")
	pods, err := k8s.KubeGetAllPods(namespace, name)
	if err != nil {
		return err
	}

	for _, pod := range pods {
		err = k8s.KubeWaitOnPod(namespace, pod.Name)
		if err != nil {
			return err
		}
	}

	fmt.Println("The Observability Stack has been installed successfully")
	fmt.Println(string(out))
	return nil
}
