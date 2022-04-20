package install

import (
	"context"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/spf13/cobra"
	"github.com/timescale/tobs/cli/cmd"
	"github.com/timescale/tobs/cli/cmd/common"
	"github.com/timescale/tobs/cli/pkg/helm"
	"github.com/timescale/tobs/cli/pkg/k8s"
	"github.com/timescale/tobs/cli/pkg/otel"
	"github.com/timescale/tobs/cli/pkg/utils"
)

// helmInstallCmd represents the helm install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Installs The Observability Stack",
	Args:  cobra.ExactArgs(0),
	RunE:  helmInstall,
}

func init() {
	cmd.RootCmd.AddCommand(installCmd)
	cmd.AddRootFlags(installCmd)
	addInstallUtilitiesFlags(installCmd)

}

func addInstallUtilitiesFlags(cmd *cobra.Command) {
	cmd.Flags().BoolP("only-secrets", "", false, "[DEPRECATED] Option to create only TimescaleDB secrets")
	cmd.Flags().BoolP("enable-timescaledb-backup", "b", false, "Option to enable TimescaleDB S3 backup")
	cmd.Flags().StringP("timescaledb-tls-cert", "", "", "Option to provide your own tls certificate for TimescaleDB")
	cmd.Flags().StringP("timescaledb-tls-key", "", "", "Option to provide your own tls key for TimescaleDB")
	cmd.Flags().StringP("version", "", "", "Option to provide tobs helm chart version, if not provided will install the latest tobs chart available")
	cmd.Flags().BoolP("skip-wait", "", false, "Option to do not wait for pods to get into running state (useful for faster tobs installation)")
	cmd.Flags().BoolP("enable-prometheus-ha", "", false, "Option to enable prometheus and promscale high-availability, by default scales to 3 replicas")
	cmd.Flags().BoolP("tracing", "", false, "Option to enable OpenTelemetry and Jaeger components")
	cmd.Flags().StringP("external-timescaledb-uri", "e", "", "Connect to an existing db using the provided URI")
	cmd.Flags().BoolP("confirm", "y", false, "Confirmation for all user input prompts")
}

type InstallSpec struct {
	ConfigFile         string
	Ref                string
	dbURI              string
	version            string
	enableBackUp       bool
	enableOtel         bool
	onlySecrets        bool
	enablePrometheusHA bool
	skipWait           bool
	tsDBTlsCert        []byte
	tsDBTlsKey         []byte
	confirmActions     bool
	dbPassword         string
}

func helmInstall(cmd *cobra.Command, args []string) error {
	var err error

	var i InstallSpec
	i.ConfigFile, err = cmd.Flags().GetString("filename")
	if err != nil {
		return fmt.Errorf("could not install The Observability Stack: %w", err)
	}
	i.Ref, err = cmd.Flags().GetString("chart-reference")
	if err != nil {
		return fmt.Errorf("could not install The Observability Stack: %w", err)
	}
	i.dbURI, err = cmd.Flags().GetString("external-timescaledb-uri")
	if err != nil {
		return fmt.Errorf("could not install The Observability Stack: %w", err)
	}
	i.enableBackUp, err = cmd.Flags().GetBool("enable-timescaledb-backup")
	if err != nil {
		return fmt.Errorf("could not install The Observability Stack: %w", err)
	}
	i.enableOtel, err = cmd.Flags().GetBool("tracing")
	if err != nil {
		return fmt.Errorf("could not install The Observability Stack: %w", err)
	}
	i.version, err = cmd.Flags().GetString("version")
	if err != nil {
		return fmt.Errorf("could not install The Observability Stack: %w", err)
	}
	i.onlySecrets, err = cmd.Flags().GetBool("only-secrets")
	if err != nil {
		return fmt.Errorf("could not install The Observability Stack: %w", err)
	}
	i.skipWait, err = cmd.Flags().GetBool("skip-wait")
	if err != nil {
		return fmt.Errorf("could not install The Observability Stack: %w", err)
	}
	i.enablePrometheusHA, err = cmd.Flags().GetBool("enable-prometheus-ha")
	if err != nil {
		return fmt.Errorf("could not install The Observability Stack: %w", err)
	}
	i.confirmActions, err = cmd.Flags().GetBool("confirm")
	if err != nil {
		return fmt.Errorf("could not install The Observability Stack: %w", err)
	}

	certFile, err := cmd.Flags().GetString("timescaledb-tls-cert")
	if err != nil {
		return fmt.Errorf("could not install The Observability Stack: %w", err)
	}

	keyFile, err := cmd.Flags().GetString("timescaledb-tls-key")
	if err != nil {
		return fmt.Errorf("could not install The Observability Stack: %w", err)
	}

	if certFile != "" && keyFile != "" {
		i.tsDBTlsCert, err = ioutil.ReadFile(certFile)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %v", certFile, err)
		}

		i.tsDBTlsKey, err = ioutil.ReadFile(keyFile)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %v", keyFile, err)
		}
	} else if certFile != "" && keyFile == "" {
		return fmt.Errorf("receieved only TLS certificate, please provide TLS key in --timescaledb-tls-key")
	} else if certFile == "" && keyFile != "" {
		return fmt.Errorf("receieved only TLS key, please provide TLS certificate in --timescaledb-tls-cert")
	}

	err = i.InstallStack()
	if err != nil {
		return fmt.Errorf("could not install The Observability Stack: %w", err)
	}
	return nil
}

var helmClient helm.Client
var k8sClient k8s.Client
var helmValues = `
cli: true`

func (c *InstallSpec) InstallStack() error {
	var err error

	helmClient = helm.NewClient(cmd.Namespace)
	defer helmClient.Close()

	// if custom helm chart is provided there is no point
	// of adding & upgrading the default tobs helm chart
	if c.Ref == utils.DEFAULT_CHART {
		err = helmClient.AddOrUpdateChartRepo(utils.DEFAULT_REGISTRY_NAME, utils.REPO_LOCATION)
		if err != nil {
			return fmt.Errorf("failed to add & update tobs helm chart: %w", err)
		}
	}

	helmValuesSpec := helm.ChartSpec{
		ReleaseName: cmd.HelmReleaseName,
		ChartName:   c.Ref,
		Namespace:   cmd.Namespace,
		// by default prior to helm install
		// we create namespace using kubeClient to
		// create TimescaleDB secrets prior to the
		// actual installation, the below CreateNamespace
		// option is useful for backward compatibility
		// i.e. if a user wants to install tobs helm chart < 0.3.0
		// this option creates the namespace.
		CreateNamespace: true,
	}

	if c.ConfigFile != "" {
		helmValuesSpec.ValuesFiles = []string{c.ConfigFile}
	}

	err = c.enableTimescaleDBBackup()
	if err != nil {
		return err
	}

	if c.enablePrometheusHA {
		helmValues = appendPrometheusHAValues(helmValues)
	}

	err = c.enableOtelOperator()
	if err != nil {
		return err
	}

	if c.enableOtel {
		// opentelemetry operator needs cert-manager as a dependency as adding cert-manager isn't good practice and
		// not recommended by the cert-manager maintainers. We are explicitly creating cert-manager with kubectl
		// for more details on this refer: https://github.com/jetstack/cert-manager/issues/3616
		err = otel.CreateCertManager(c.confirmActions)
		if err != nil {
			return fmt.Errorf("failed to create cert-manager %v", err)
		}
	}

	if c.version != "" {
		helmValuesSpec.Version = c.version
	}

	e, err := helmClient.ExportValuesFieldFromChart(c.Ref, c.ConfigFile, []string{"timescaledb-single", "enabled"})
	if err != nil {
		return err
	}
	enableTimescaleDB, err := utils.InterfaceToBool(e)
	if err != nil {
		return fmt.Errorf("cannot convert timescaledb-single.enabled to bool, %v", err)
	}

	// As multiple times we are appending Promscale values the below func
	// helps us to append by overriding the previous configs field by field
	promscaleConfig := appendPromscaleValues(c.enableOtel, enableTimescaleDB, c.enablePrometheusHA, c.dbURI, c.dbPassword, c.version)
	helmValues = helmValues + promscaleConfig

	helmValuesSpec.ValuesYaml = helmValues
	fmt.Println("Installing The Observability Stack")
	release, err := helmClient.InstallOrUpgradeChart(context.Background(), &helmValuesSpec)
	if err != nil {
		return fmt.Errorf("could not install The Observability Stack: %w", err)
	}

	fmt.Println("Waiting for helm install to complete...")

	time.Sleep(10 * time.Second)

	k8sClient = k8s.NewClient()
	err = c.waitForPods()
	if err != nil {
		return err
	}

	// create the default otelcol CR as operator should be is up & running by now....
	err = c.deployOtelCollectorCR()
	if err != nil {
		return err
	}

	if release.Info == nil {
		fmt.Println("failed to install tobs completely, release notes generation failed...")
		return nil
	}

	fmt.Println(release.Info.Notes)
	fmt.Println("The Observability Stack has been installed successfully")
	return nil
}

func (c *InstallSpec) waitForPods() error {
	if !c.skipWait {
		fmt.Println("Waiting for pods to initialize...")
		pods, err := k8sClient.KubeGetAllPods(cmd.Namespace, cmd.HelmReleaseName)
		if err != nil {
			return err
		}

		for _, pod := range pods {
			err = k8sClient.KubeWaitOnPod(cmd.Namespace, pod.Name)
			if err != nil {
				return err
			}
		}
	} else {
		fmt.Println("skipping the wait for pods to come to a running state because --skip-wait is enabled.")
	}

	return nil
}

func (c *InstallSpec) enableTimescaleDBBackup() error {
	// If enable backup is disabled by flag check the backup option
	// from values.yaml as a second option
	if !c.enableBackUp {
		e, err := helmClient.ExportValuesFieldFromChart(c.Ref, c.ConfigFile, common.TimescaleDBBackUpKeyForValuesYaml)
		if err != nil {
			return err
		}
		var ok bool
		c.enableBackUp, ok = e.(bool)
		if !ok {
			return fmt.Errorf("enable Backup was not a bool")
		}
	} else {
		// update timescaleDB backup in values.yaml
		helmValues = helmValues + `
timescaledb-single:
  backup:
    enabled: true`
	}

	return nil
}

func (c *InstallSpec) enableOtelOperator() error {
	if c.enableOtel {
		helmValues = enableOtelInValues(helmValues)
	} else {
		e, err := helmClient.ExportValuesFieldFromChart(c.Ref, c.ConfigFile, []string{"opentelemetryOperator", "enabled"})
		if err != nil {
			return err
		}
		var ok bool
		c.enableOtel, ok = e.(bool)
		if !ok {
			return fmt.Errorf("opentelemetryOperator.enabled is not a bool")
		}
	}
	return nil
}

func (c *InstallSpec) deployOtelCollectorCR() error {
	if c.enableOtel {
		otelCol := otel.OtelCol{
			ReleaseName: cmd.HelmReleaseName,
			Namespace:   cmd.Namespace,
			K8sClient:   k8sClient,
			HelmClient:  helmClient,
		}
		config, err := helmClient.ExportValuesFieldFromChart(c.Ref, c.ConfigFile, []string{"opentelemetryOperator", "collector", "config"})
		if err != nil {
			return err
		}
		otelColConfig, ok := config.(string)
		if !ok {
			return fmt.Errorf("opentelemetryOperator.collector.config is not a string")
		}
		if err = otelCol.CreateDefaultCollector(otelColConfig); err != nil {
			return err
		}
	}

	return nil
}

func enableOtelInValues(helmValues string) string {
	helmValues = helmValues + `
opentelemetryOperator:
  enabled: true
`
	return helmValues
}

func appendPromscaleValues(enableOtel, timescaledb, promHA bool, dbURI, dbPassword, version string) string {
	var args string
	config := fmt.Sprintf(`
promscale:  
  extraEnv:
  - name: "TOBS_TELEMETRY_INSTALLED_BY"
    value: "cli"
  - name: "TOBS_TELEMETRY_VERSION"
    value: "%s"
  - name: "TOBS_TELEMETRY_TRACING_ENABLED"
    value: "%t"
  - name: "TOBS_TELEMETRY_TIMESCALEDB_ENABLED"
    value: "%t"`, version, enableOtel, timescaledb)

	if enableOtel {
		config = config + `
  openTelemetry:
    enabled: true`
	}

	if promHA {
		config = config + `
  replicaCount: 3`
		args = args + `
  - --metrics.high-availability`
	}

	if dbURI != "" {
		config = config + fmt.Sprintf(`
  connection:
    uri: %s`, dbURI)
	} else {
		config = config + fmt.Sprintf(`
  connection:
    password: %s
    host: %s.%s.svc.cluster.local`, dbPassword, cmd.HelmReleaseName, cmd.Namespace)
	}

	if args != "" {
		args = `
  extraArgs:` + args
	}

	return config + args
}

func appendPrometheusHAValues(helmValues string) string {
	helmValues = helmValues + `
timescaledb-single:
  patroni:
    bootstrap:
      dcs:
        postgresql:
          parameters:
            max_connections: 400

kube-prometheus-stack:
  prometheus:
    prometheusSpec:
      replicas: 2
      prometheusExternalLabelName: cluster
      replicaExternalLabelName: __replica__
`
	return helmValues
}
