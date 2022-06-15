package install

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/timescale/tobs/cli/cmd"
	"github.com/timescale/tobs/cli/cmd/common"
	"github.com/timescale/tobs/cli/pkg/helm"
	"github.com/timescale/tobs/cli/pkg/otel"
	"github.com/timescale/tobs/cli/pkg/utils"
	"helm.sh/helm/v3/pkg/release"
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
	cmd.Flags().BoolP("enable-timescaledb-backup", "b", false, "Option to enable TimescaleDB S3 backup")
	cmd.Flags().StringP("version", "", "", "Option to provide tobs helm chart version, if not provided will install the latest tobs chart available")
	cmd.Flags().BoolP("enable-prometheus-ha", "", false, "Option to enable prometheus and promscale high-availability, by default scales to 2 replicas")
	cmd.Flags().StringP("external-timescaledb-uri", "e", "", "Connect to an existing db using the provided URI")
	cmd.Flags().BoolP("confirm", "y", false, "Confirmation for all user input prompts")
	cmd.Flags().StringP("export", "", "", "export kubernetes manifests to a yaml file instead of actually installing")
}

type InstallSpec struct {
	ConfigFile         string
	Ref                string
	dbURI              string
	version            string
	enableBackUp       bool
	enablePrometheusHA bool
	ConfirmActions     bool
	dbPassword         string
	exportFile         string
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
	i.ConfirmActions, err = cmd.Flags().GetBool("confirm")
	if err != nil {
		return fmt.Errorf("could not install The Observability Stack: %w", err)
	}
	i.exportFile, err = cmd.Flags().GetString("export")
	if err != nil {
		return fmt.Errorf("could not install The Observability Stack: %w", err)
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
	err = otel.CreateCertManager(c.ConfirmActions)
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

	if c.exportFile != "" {
		fmt.Println("enabled dry-run mode for helm")
		helmValuesSpec.DryRun = true
		helmValuesSpec.Replace = true
	}

	fmt.Println("Installing The Observability Stack, this can take a few minutes")
	release, err := helmClient.InstallOrUpgradeChart(context.Background(), &helmValuesSpec)
	if err != nil {
		return fmt.Errorf("could not install The Observability Stack: %w", err)
	}

	if c.exportFile != "" {
		export, err := os.Create(c.exportFile)
		if err != nil {
			return fmt.Errorf("failed to create export file: %s, %w", c.exportFile, err)
		}
		defer export.Close()
		fmt.Printf("exporting to %s\n", c.exportFile)
		fmt.Fprintln(export, strings.TrimSpace(release.Manifest))
		for _, m := range release.Hooks {
			if isTestHook(m) {
				continue
			}
			fmt.Fprintf(export, "---\n# Source: %s\n%s\n", m.Path, m.Manifest)
		}
	}

	if release.Info == nil {
		fmt.Println("failed to install tobs completely, release notes generation failed...")
		return nil
	}

	fmt.Println(release.Info.Notes)
	fmt.Println("The Observability Stack has been installed successfully")
	return nil
}

func isTestHook(h *release.Hook) bool {
	for _, e := range h.Events {
		if e == release.HookTest {
			return true
		}
	}
	return false
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
