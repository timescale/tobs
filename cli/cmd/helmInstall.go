package cmd

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/timescale/tobs/cli/pkg/timescaledb_secrets"

	"github.com/timescale/tobs/cli/pkg/utils"

	"github.com/timescale/tobs/cli/pkg/k8s"

	"github.com/spf13/cobra"
)

const DEVEL = false

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
	cmd.Flags().StringP("external-timescaledb-uri", "e", "", "Connect to an existing db using the provided URI")
	cmd.Flags().BoolP("enable-timescaledb-backup", "b", false, "Enable TimescaleDB S3 backup")
}

func helmInstall(cmd *cobra.Command, args []string) error {
	var err error

	var file, ref, dbURI string
	file, err = cmd.Flags().GetString("filename")
	if err != nil {
		return fmt.Errorf("could not install The Observability Stack: %w", err)
	}
	ref, err = cmd.Flags().GetString("chart-reference")
	if err != nil {
		return fmt.Errorf("could not install The Observability Stack: %w", err)
	}
	dbURI, err = cmd.Flags().GetString("external-timescaledb-uri")
	if err != nil {
		return fmt.Errorf("could not install The Observability Stack: %w", err)
	}

	enableBackUp, err := cmd.Flags().GetBool("enable-timescaledb-backup")
	if err != nil {
		return fmt.Errorf("could not install The Observability Stack: %w", err)
	}

	err = installStack(file, ref, dbURI, enableBackUp)
	if err != nil {
		return fmt.Errorf("could not install The Observability Stack: %w", err)
	}
	return nil
}

func installStack(file, ref, dbURI string, enableBackUp bool) error {
	var err error
	// if custom helm chart is provided there is no point
	// of adding & upgrading the default tobs helm chart
	if ref == utils.DEFAULT_CHART {
		err = utils.AddTobsHelmChart()
		if err != nil {
			return err
		}

		err = utils.UpdateTobsHelmChart(false)
		if err != nil {
			return err
		}
	}

	cmds := []string{"install", name, ref}
	cmds = append(cmds, "--set", "cli=true")
	if dbURI != "" {
		cmds, err = appendDBURIValues(dbURI, name, cmds)
		if err != nil {
			return err
		}
	}

	if namespace != "default" {
		cmds = append(cmds, "--create-namespace", "--namespace", namespace)
	}
	if file != "" {
		cmds = append(cmds, "--values", file)
	}
	if DEVEL {
		cmds = append(cmds, "--devel")
	}

	// If enable backup is disabled by flag check the backup option
	// from values.yaml as a second option
	if !enableBackUp {
		enableBackUp, err = utils.ExportBackUpEnabledField(ref)
		if err != nil {
			return err
		}
	} else {
		fmt.Println("")
		// TODO to self
		// set timescaledb.single.backup.enabled=true if backup option is enabled from install flag
		// waiting for other PR on this as it holds changes around this.
	}

	err = timescaledb_secrets.CreateTimescaleDBSecrets(name, namespace, enableBackUp)
	if err != nil {
		return err
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

func appendDBURIValues(dbURI, name string, cmds []string) ([]string, error) {
	cmds = append(cmds, "--set",
		"timescaledb-single.enabled=false,"+
			"timescaledbExternal.enabled=true,"+
			"timescaledbExternal.db_uri="+dbURI+
			",promscale.connection.uri.secretTemplate="+name+"-timescaledb-uri,")
	return cmds, nil
}
