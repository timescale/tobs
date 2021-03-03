package helm

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"time"

	"github.com/spf13/cobra"
	"github.com/timescale/tobs/cli/cmd"
	"github.com/timescale/tobs/cli/pkg/k8s"
	"github.com/timescale/tobs/cli/pkg/timescaledb_secrets"
	"github.com/timescale/tobs/cli/pkg/utils"
)

const DEVEL = false

var TimescaleDBBackUpKeyForValuesYaml = []string{"timescaledb-single", "backup", "enabled"}

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
	cmd.Flags().StringP("timescaledb-tls-cert", "", "", "Option to provide your own tls certificate for TimescaleDB")
	cmd.Flags().StringP("timescaledb-tls-key", "", "", "Option to provide your own tls key for TimescaleDB")
}

type installSpec struct {
	configFile   string
	ref          string
	dbURI        string
	enableBackUp bool
	tsDBTlsCert  []byte
	tsDBTlsKey   []byte
}

func helmInstall(cmd *cobra.Command, args []string) error {
	var err error

	var i installSpec
	i.configFile, err = cmd.Flags().GetString("filename")
	if err != nil {
		return fmt.Errorf("could not install The Observability Stack: %w", err)
	}
	i.ref, err = cmd.Flags().GetString("chart-reference")
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

	err = i.installStack()
	if err != nil {
		return fmt.Errorf("could not install The Observability Stack: %w", err)
	}
	return nil
}

func (c *installSpec) installStack() error {
	var err error
	// if custom helm chart is provided there is no point
	// of adding & upgrading the default tobs helm chart
	if c.ref == utils.DEFAULT_CHART {
		err = utils.AddTobsHelmChart()
		if err != nil {
			return err
		}

		err = utils.UpdateTobsHelmChart(false)
		if err != nil {
			return err
		}
	}

	cmds := []string{"install", cmd.HelmReleaseName, c.ref}

	// Note do not change the below order the --set flag is set
	// in above we are appending more flags to it with below check
	helmValues := "cli=true"

	// If enable backup is disabled by flag check the backup option
	// from values.yaml as a second option
	if !c.enableBackUp {
		e, err := utils.ExportValuesFieldFromChart(c.ref, TimescaleDBBackUpKeyForValuesYaml)
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
		helmValues = helmValues + ",timescaledb-single.backup.enabled=true"
	}

	t := timescaledb_secrets.TSDBSecretsInfo{
		ReleaseName:    cmd.HelmReleaseName,
		Namespace:      cmd.Namespace,
		EnableS3Backup: c.enableBackUp,
		TlsCert:        c.tsDBTlsCert,
		TlsKey:         c.tsDBTlsKey,
	}

	if c.dbURI != "" {
		helmValues = appendDBURIValues(c.dbURI, cmd.HelmReleaseName, helmValues)
	} else {
		// if db-uri is provided we do not need
		// to create DB level secrets
		err = t.CreateTimescaleDBSecrets()
		if err != nil {
			return err
		}
	}

	if cmd.Namespace != "default" {
		cmds = append(cmds, "--create-namespace", "--namespace", cmd.Namespace)
	}
	if c.configFile != "" {
		cmds = append(cmds, "--values", c.configFile)
	}
	if DEVEL {
		cmds = append(cmds, "--devel")
	}

	cmds = append(cmds, "--set", helmValues)
	install := exec.Command("helm", cmds...)
	fmt.Println("Installing The Observability Stack")
	out, err := install.CombinedOutput()
	if err != nil {
		return fmt.Errorf("could not install The Observability Stack: %w \nOutput: %v", err, string(out))
	}

	fmt.Println("Waiting for helm install to complete...")

	time.Sleep(10 * time.Second)

	fmt.Println("Waiting for pods to initialize...")
	pods, err := k8s.KubeGetAllPods(cmd.Namespace, cmd.HelmReleaseName)
	if err != nil {
		return err
	}

	for _, pod := range pods {
		err = k8s.KubeWaitOnPod(cmd.Namespace, pod.Name)
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
