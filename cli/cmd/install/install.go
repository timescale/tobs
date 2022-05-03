package install

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/timescale/tobs/cli/cmd"
	"github.com/timescale/tobs/cli/cmd/common"
	"github.com/timescale/tobs/cli/pkg/helm"
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
	cmd.Flags().BoolP("only-secrets", "", false, "[DEPRECATED] Defunct flag historically used to create TimescaleDB secrets")
	cmd.Flags().StringP("timescaledb-tls-cert", "", "", "[DEPRECATED] Use helm values file to configure TLS certificate. This option can be configured with either 'timescaledb-single.secrets.certificate' or 'timescaledb-single.secrets.certificateSecretName'")
	cmd.Flags().StringP("timescaledb-tls-key", "", "", "[DEPRECATED] Use helm values file to configure TLS certificate. This option can be configured with either 'timescaledb-single.secrets.certificate' or 'timescaledb-single.secrets.certificateSecretName'")
	cmd.Flags().BoolP("enable-timescaledb-backup", "b", false, "Option to enable TimescaleDB S3 backup")
	cmd.Flags().StringP("version", "", "", "Option to provide tobs helm chart version, if not provided will install the latest tobs chart available")
	cmd.Flags().BoolP("skip-wait", "", false, "[DEPRECATED] flag is not functional as tobs installation requires waiting for pods to be in running state due to opentelemetry prerequisities")
	cmd.Flags().BoolP("enable-prometheus-ha", "", false, "Option to enable prometheus and promscale high-availability, by default scales to 2 replicas")
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
	enablePrometheusHA bool
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
	i.version, err = cmd.Flags().GetString("version")
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

	// TODO(paulfantom): Remove deprecated flags post 0.11.0 release
	if cmd.Flags().Changed("tracing") {
		fmt.Println("DEPRECATED flag used: 'tracing'. This flag will be removed in future versions of tobs. Feature is now enabled by default and if you want to disable deployment of opentelemetry-operator, change value of opentelemetryOperator.enabled")
	}
	if cmd.Flags().Changed("skip-wait") {
		fmt.Println("DEPRECATED flag used: 'skip-wait'. This flag will be removed in future versions of tobs. Feature is now disabled by default to allow smooth installation of opentelemetry components")
	}
	if cmd.Flags().Changed("only-secrets") {
		fmt.Println("DEPRECATED flag used: 'only-secrets'. This flag will be removed in future versions of tobs.")
	}
	if cmd.Flags().Changed("timescaledb-tls-cert") {
		fmt.Println("DEPRECATED flag used: 'timescaledb-tls-cert'. This flag will be removed in future versions of tobs. Use helm values file to configure TLS certificate. This option can be configured with either 'timescaledb-single.secrets.certificate' or 'timescaledb-single.secrets.certificateSecretName'")
	}
	if cmd.Flags().Changed("timescaledb-tls-key") {
		fmt.Println("DEPRECATED flag used: 'timescaledb-tls-key'. This flag will be removed in future versions of tobs. Use helm values file to configure TLS certificate. This option can be configured with either 'timescaledb-single.secrets.certificate' or 'timescaledb-single.secrets.certificateSecretName'")
	}

	err = i.InstallStack()
	if err != nil {
		return fmt.Errorf("could not install The Observability Stack: %w", err)
	}
	return nil
}

var helmClient helm.Client
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
		Wait:            true,
		Timeout:         15 * time.Minute,
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

	// opentelemetry operator needs cert-manager as a dependency as adding cert-manager isn't good practice and
	// not recommended by the cert-manager maintainers. We are explicitly creating cert-manager with kubectl
	// for more details on this refer: https://github.com/jetstack/cert-manager/issues/3616
	err = otel.CreateCertManager(c.confirmActions)
	if err != nil {
		return fmt.Errorf("failed to create cert-manager %v", err)
	}

	if c.version != "" {
		helmValuesSpec.Version = c.version
	}

	v, err := helmClient.ExportValuesFieldFromChart(c.Ref, c.ConfigFile, []string{"promscale", "openTelemetry", "enabled"})
	if err != nil {
		return err
	}
	enabledOTEL, err := utils.InterfaceToBool(v)
	if err != nil {
		return fmt.Errorf("cannot convert promscale.openTelemetry.enabled to bool, %v", err)
	}
	// waiting for helm install completion is necessary only for applying otel CRs. This is an opentelemetry-operator prerequiste
	//helmValuesSpec.Wait = enabledOTEL

	v, err = helmClient.ExportValuesFieldFromChart(c.Ref, c.ConfigFile, []string{"timescaledb-single", "enabled"})
	if err != nil {
		return err
	}
	enableTimescaleDB, err := utils.InterfaceToBool(v)
	if err != nil {
		return fmt.Errorf("cannot convert timescaledb-single.enabled to bool, %v", err)
	}

	// As multiple times we are appending Promscale values the below func
	// helps us to append by overriding the previous configs field by field
	promscaleConfig := appendPromscaleValues(enabledOTEL, enableTimescaleDB, c.enablePrometheusHA, c.dbURI, c.dbPassword, c.version)
	helmValues = helmValues + promscaleConfig

	helmValuesSpec.ValuesYaml = helmValues
	fmt.Println("Installing The Observability Stack, this can take a few minutes")
	release, err := helmClient.InstallOrUpgradeChart(context.Background(), &helmValuesSpec)
	if err != nil {
		return fmt.Errorf("could not install The Observability Stack: %w", err)
	}

	if release.Info == nil {
		fmt.Println("failed to install tobs completely, release notes generation failed...")
		return nil
	}

	fmt.Println(release.Info.Notes)
	fmt.Println("The Observability Stack has been installed successfully")
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
    host: %s.%s.svc`, dbPassword, cmd.HelmReleaseName, cmd.Namespace)
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
