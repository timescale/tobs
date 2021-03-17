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
	cmd.Flags().BoolP("enable-kube-prometheus", "k", false, "Enable Kube-Prometheus")
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

	kp, err := cmd.Flags().GetBool("enable-kube-prometheus")
	if err != nil {
		return fmt.Errorf("could not install The Observability Stack: %w", err)
	}

	err = installStack(file, ref, dbURI, enableBackUp, kp)

	if err != nil {
		return fmt.Errorf("could not install The Observability Stack: %w", err)
	}
	return nil
}

func installStack(file, ref, dbURI string, enableBackUp, enableKubePrometheus bool) error {
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

	// Note do not change the below order the --set flag is set
	// in above we are appending more flags to it with below check
	cmds = append(cmds, "--set")
	helmValues := "cli=true"
	if dbURI != "" {
		helmValues = appendDBURIValues(dbURI, name, helmValues)
	}

	// If enable Kube-Prometheus is disabled by flag check the backup option
	// from values.yaml as a second option
	if !enableKubePrometheus {
		e, err := utils.ExportValuesFieldValue(ref, []string{"kube-prometheus-stack", "enabled"})
		enableKubePrometheus = e.(bool)
		if err != nil {
			return err
		}

		// If kubePrometheus is enabled from values.yaml amke sure to
		// validate default Prometheus & Grafana from tobs are disabled
		// if not disabled we will end up with duplicate components
		if enableKubePrometheus {
			e, err := utils.ExportValuesFieldValue(ref, []string{"prometheus", "enabled"})
			enablePrometheus := e.(bool)
			if err != nil {
				return err
			}

			e, err = utils.ExportValuesFieldValue(ref, []string{"grafana", "enabled"})
			enableGrafana := e.(bool)
			if err != nil {
				return err
			}

			if enablePrometheus || enableGrafana {
				return fmt.Errorf("kube-prometheus-stack is enabled but prometheus or grafana from default tobs are not disabled")
			}
		}
	} else if enableKubePrometheus {
		helmValues = enableKubePrometheusStack(helmValues)
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
		e, err := utils.ExportValuesFieldValue(ref, []string{"timescaledb-single", "backup", "enabled"})
		enableBackUp = e.(bool)
		if err != nil {
			return err
		}
	} else {
		// update timescaleDB backup in values.yaml
		helmValues = helmValues+",timescaledb-single.backup.enabled=true"
	}

	err = timescaledb_secrets.CreateTimescaleDBSecrets(name, namespace, enableBackUp)
	if err != nil {
		return err
	}

	cmds = append(cmds, helmValues)

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

func appendDBURIValues(dbURI, name string, helmValues string) string {
	helmValues = helmValues + ",timescaledb-single.enabled=false," + "timescaledbExternal.enabled=true," + "timescaledbExternal.db_uri=" + dbURI +
		",promscale.connection.uri.secretTemplate=" + name + "-timescaledb-uri"
	return helmValues
}

func enableKubePrometheusStack(helmValues string) string {
	helmValues = helmValues+",prometheus.enabled=false,"+"grafana.enabled=false,kube-prometheus-stack.enabled=true"
	return helmValues
}