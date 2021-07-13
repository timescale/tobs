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
	"github.com/timescale/tobs/cli/pkg/timescaledb_secrets"
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
	cmd.Flags().BoolP("only-secrets", "", false, "Option to create only TimescaleDB secrets")
	cmd.Flags().BoolP("enable-timescaledb-backup", "b", false, "Option to enable TimescaleDB S3 backup")
	cmd.Flags().StringP("timescaledb-tls-cert", "", "", "Option to provide your own tls certificate for TimescaleDB")
	cmd.Flags().StringP("timescaledb-tls-key", "", "", "Option to provide your own tls key for TimescaleDB")
	cmd.Flags().StringP("version", "", "", "Option to provide tobs helm chart version, if not provided will install the latest tobs chart available")
	cmd.Flags().BoolP("skip-wait", "", false, "Option to do not wait for pods to get into running state (useful for faster tobs installation)")
	cmd.Flags().BoolP("enable-prometheus-ha", "", false, "Option to enable prometheus and promscale high-availability, by default scales to 3 replicas")
	cmd.Flags().StringP("external-timescaledb-uri", "e", "", "Connect to an existing db using the provided URI")
}

type InstallSpec struct {
	ConfigFile         string
	Ref                string
	dbURI              string
	version            string
	enableBackUp       bool
	onlySecrets        bool
	enablePrometheusHA bool
	skipWait           bool
	tsDBTlsCert        []byte
	tsDBTlsKey         []byte
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

func (c *InstallSpec) InstallStack() error {
	var err error
	helmValues := `
cli: true`

	if c.dbURI != "" {
		helmValues = appendDBURIValues(c.dbURI, cmd.HelmReleaseName, helmValues)
	} else {
		// if db-uri is provided we do not need
		// to create DB level secrets
		err = c.createSecrets()
		if err != nil {
			return fmt.Errorf("failed to create secrets %v", err)
		}
		if c.onlySecrets {
			fmt.Println("Skipping tobs installation because of only-secrets flag.")
			fmt.Println("Successfully created secrets for TimescaleDB.")
			return nil
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

	helmClient := helm.NewClient(cmd.Namespace)
	defer helmClient.Close()
	// if custom helm chart is provided there is no point
	// of adding & upgrading the default tobs helm chart
	if c.Ref == utils.DEFAULT_CHART {
		err = helmClient.AddOrUpdateChartRepo(utils.DEFAULT_REGISTRY_NAME, utils.REPO_LOCATION)
		if err != nil {
			return fmt.Errorf("failed to add & update tobs helm chart: %w", err)
		}
	}

	if c.ConfigFile != "" {
		helmValuesSpec.ValuesFiles = []string{c.ConfigFile}
	}

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

	if c.enablePrometheusHA {
		helmValues = appendPrometheusHAValues(helmValues)
	}

	if c.version != "" {
		helmValuesSpec.Version = c.version
	}

	helmValuesSpec.ValuesYaml = helmValues

	fmt.Println("Installing The Observability Stack")
	release, err := helmClient.InstallOrUpgradeChart(context.Background(), &helmValuesSpec)
	if err != nil {
		return fmt.Errorf("could not install The Observability Stack: %w", err)
	}

	fmt.Println("Waiting for helm install to complete...")

	time.Sleep(10 * time.Second)

	k8sClient := k8s.NewClient()
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

	fmt.Println("The Observability Stack has been installed successfully")
	fmt.Println(release.Info.Notes)
	return nil
}

func appendDBURIValues(dbURI, name string, helmValues string) string {
	helmValues = helmValues + fmt.Sprintf(`
timescaledb-single:
  enabled: false
timescaledbExternal:
  enabled: true
  db_uri: %s
promscale:
  connection:
    uri: 
      secretTemplate: %s-timescaledb-uri`, dbURI, name)
	return helmValues
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

promscale:
  replicaCount: 3
  args:
  - --high-availability

kube-prometheus-stack:
  prometheus:
    prometheusSpec:
      replicas: 3
      prometheusExternalLabelName: cluster
      replicaExternalLabelName: __replica__
`
	return helmValues
}

func (c *InstallSpec) createSecrets() error {
	var i int64
	var err error
	if c.version != "" {
		i, err = utils.ParseVersion(c.version, 3)
		if err != nil {
			return fmt.Errorf("failed to parse version %s %v", c.version, err)
		}
	}

	// here 3000 represent version
	// equal to or greater than 0.3.0
	// if version isn't provided new
	// installations needs secrets
	if i > 3000 || c.version == "" {
		t := timescaledb_secrets.TSDBSecretsInfo{
			ReleaseName:    cmd.HelmReleaseName,
			Namespace:      cmd.Namespace,
			EnableS3Backup: c.enableBackUp,
			TlsCert:        c.tsDBTlsCert,
			TlsKey:         c.tsDBTlsKey,
			K8sClient:      k8s.NewClient(),
		}
		err := t.CreateTimescaleDBSecrets()
		if err != nil {
			return err
		}
	}

	return nil
}
