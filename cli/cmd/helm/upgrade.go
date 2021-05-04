package helm

import (
	"errors"
	"fmt"
	"os/exec"
	"reflect"
	"strings"

	"github.com/spf13/cobra"
	root "github.com/timescale/tobs/cli/cmd"
	"github.com/timescale/tobs/cli/pkg/k8s"
	"github.com/timescale/tobs/cli/pkg/utils"
	v1 "k8s.io/api/core/v1"
	errors2 "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// upgradeCmd represents the upgrade command
var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Alias for helm upgrade",
	Args:  cobra.ExactArgs(0),
	RunE:  upgrade,
}

func init() {
	root.RootCmd.AddCommand(upgradeCmd)
	addHelmInstallFlags(upgradeCmd)
	upgradeCmd.Flags().BoolP("reset-values", "", false, "Reset helm chart to default values of the helm chart. This is same flag that exists in helm upgrade")
	upgradeCmd.Flags().BoolP("reuse-values", "", false, "Reuse the last release's values and merge in any overrides from the command line via --set and -f. If '--reset-values' is specified, this is ignored.\nThis is same flag that exists in helm upgrade ")
	upgradeCmd.Flags().BoolP("same-chart", "", false, "Use the same helm chart do not upgrade helm chart but upgrade the existing chart with new values")
	upgradeCmd.Flags().BoolP("confirm", "y", false, "Confirmation flag for upgrading")
	upgradeCmd.Flags().BoolP("skip-crds", "", false, "Option to skip creating CRDs on upgrade")
}

func upgrade(cmd *cobra.Command, args []string) error {
	return upgradeTobs(cmd, args)
}

type upgradeSpec struct {
	deployedChartVersion string
	newChartVersion      string
	skipCrds             bool
}

func upgradeTobs(cmd *cobra.Command, args []string) error {
	file, err := cmd.Flags().GetString("filename")
	if err != nil {
		return fmt.Errorf("couldn't get the filename flag value: %w", err)
	}

	ref, err := cmd.Flags().GetString("chart-reference")
	if err != nil {
		return fmt.Errorf("couldn't get the chart-reference flag value: %w", err)
	}

	reset, err := cmd.Flags().GetBool("reset-values")
	if err != nil {
		return fmt.Errorf("couldn't get the reset-values flag value: %w", err)
	}

	reuse, err := cmd.Flags().GetBool("reuse-values")
	if err != nil {
		return fmt.Errorf("couldn't get the reuse-values flag value: %w", err)
	}

	confirm, err := cmd.Flags().GetBool("confirm")
	if err != nil {
		return fmt.Errorf("couldn't get the confirm flag value: %w", err)
	}

	sameChart, err := cmd.Flags().GetBool("same-chart")
	if err != nil {
		return fmt.Errorf("couldn't get the reuse-values flag value: %w", err)
	}

	enableBackUp, err := cmd.Flags().GetBool("enable-timescaledb-backup")
	if err != nil {
		return fmt.Errorf("could not install The Observability Stack: %w", err)
	}

	skipCrds, err := cmd.Flags().GetBool("skip-crds")
	if err != nil {
		return fmt.Errorf("could not install The Observability Stack: %w", err)
	}

	cmds := []string{"upgrade", root.HelmReleaseName, ref, "--namespace", root.Namespace}

	if file != "" {
		cmds = append(cmds, "--values", file)
	}

	if reset {
		cmds = append(cmds, "--reset-values")
	}

	if reuse {
		cmds = append(cmds, "--reuse-values")
	}

	latestChart, err := utils.GetTobsChartMetadata(ref)
	if err != nil {
		return err
	}

	deployedChart, err := utils.GetDeployedChartMetadata(root.HelmReleaseName, root.Namespace)
	if err != nil {
		return err
	}

	if deployedChart.Name == "" {
		fmt.Println("couldn't find the existing tobs deployment. Deploying tobs...")
		if !confirm {
			utils.ConfirmAction()
		}
		s := installSpec{
			configFile:   file,
			ref:          ref,
			dbURI:        "",
			enableBackUp: enableBackUp,
		}
		err = s.installStack()
		if err != nil {
			return err
		}
		return nil
	}

	// add & update helm chart only if it's default chart
	// if same-chart upgrade is disabled
	if ref == utils.DEFAULT_CHART && !sameChart {
		err = utils.AddTobsHelmChart()
		if err != nil {
			return err
		}

		err = utils.UpdateTobsHelmChart(true)
		if err != nil {
			return err
		}
	}

	lVersion, err := utils.ParseVersion(latestChart.Version, 3)
	if err != nil {
		return fmt.Errorf("failed to parse latest helm chart version %w", err)
	}

	deployedVersion := strings.Split(deployedChart.Chart, "-")[1]
	dVersion, err := utils.ParseVersion(deployedVersion, 3)
	if err != nil {
		return fmt.Errorf("failed to parse deployed helm chart version %w", err)
	}

	var foundNewChart bool
	if lVersion <= dVersion {
		dValues, err := utils.DeployedValuesYaml(ref, root.HelmReleaseName, root.Namespace)
		if err != nil {
			return err
		}

		nValues, err := utils.NewValuesYaml(ref, file)
		if err != nil {
			return err
		}

		if ok := reflect.DeepEqual(dValues, nValues); ok {
			err = errors.New("failed to upgrade there is no latest helm chart available and existing helm deployment values are same as the provided values")
			return err
		}
	} else {
		foundNewChart = true
		if sameChart {
			err = errors.New("provided helm chart is newer compared to existing deployed helm chart cannot upgrade as --same-chart flag is provided")
			return err
		}
	}

	if foundNewChart {
		fmt.Printf("Upgrading to latest helm chart version: %s\n", latestChart.Version)
	} else {
		fmt.Println("Upgrading the existing helm chart with values.yaml file")
	}

	if !confirm {
		utils.ConfirmAction()
	}

	upgradeDetails := &upgradeSpec{
		deployedChartVersion: deployedVersion,
		newChartVersion:      latestChart.Version,
		skipCrds:             skipCrds,
	}

	err = upgradeDetails.UpgradePathBasedOnVersion()
	if err != nil {
		return err
	}

	out := exec.Command("helm", cmds...)
	output, err := out.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to upgrade %s: %w", output, err)
	}

	fmt.Printf("Successfully upgraded %s to version: %s\n", root.HelmReleaseName, latestChart.Version)
	return nil
}

func (c *upgradeSpec) UpgradePathBasedOnVersion() error {
	nVersion, err := utils.ParseVersion(c.newChartVersion, 3)
	if err != nil {
		return fmt.Errorf("failed to parse latest helm chart version %w", err)
	}

	dVersion, err := utils.ParseVersion(c.deployedChartVersion, 3)
	if err != nil {
		return fmt.Errorf("failed to parse deployed helm chart version %w", err)
	}

	version0_2_2, err := utils.ParseVersion("0.2.2", 3)
	if err != nil {
		return fmt.Errorf("failed to parse 0.2.2 version %w", err)
	}

	version0_4_0, err := utils.ParseVersion("0.4.0", 3)
	if err != nil {
		return fmt.Errorf("failed to parse 0.2.2 version %w", err)
	}

	if nVersion >= version0_4_0 && dVersion < version0_4_0  {
		if !c.skipCrds {
			err = createCRDS()
			if err != nil {
				return err
			}
		}

		prometheusNodeExporter := "-prometheus-node-exporter"
		err = k8s.DeleteDaemonset(root.HelmReleaseName+prometheusNodeExporter, root.Namespace)
		if err != nil {
			return err
		}
		err = k8s.KubeDeleteService(root.Namespace, root.HelmReleaseName+prometheusNodeExporter)
		if err != nil {
			return err
		}
	}

	switch {
	// The below case if for upgrade from any versions <= 0.2.2 to greater versions
	case dVersion <= version0_2_2 && nVersion > version0_2_2:
		err = copyOldSecretsToNewSecrets()
		if err != nil {
			return fmt.Errorf("failed to perform copying of old secrets to new secrets %v", err)
		}

		// in this release of tobs the grafana-db-job has been updated with spec
		// the upgrade fails to patch the spec so we are deleting & the upgrade will re-create it.
		err := k8s.DeleteJob(root.HelmReleaseName+"-grafana-db", root.Namespace)
		if err != nil {
			return err
		}
	default:
		// if the upgrade doesn't match the above condition
		// that means we do not have an upgrade path for the base version to new version
		// Note: This is helpful when someone wants to upgrade with just values.yaml (not between versions)
		return nil
	}

	return nil
}

func copyOldSecretsToNewSecrets() error {
	err := copyDBSecret()
	if err != nil {
		return err
	}

	err = copyDBCertificate()
	if err != nil {
		return err
	}

	err = copyDBBackup()
	if err != nil {
		return err
	}

	return nil
}

func copyDBSecret() error {
	fmt.Println("Migrating the credentials")
	existingDBSecret := root.HelmReleaseName + "-timescaledb-passwords"
	newDBsecret := root.HelmReleaseName + "-credentials"

	s, err := k8s.KubeGetSecret(root.Namespace, existingDBSecret)
	if err != nil {
		return fmt.Errorf("failed to get the secret existing secret during the upgrade %s: %v", existingDBSecret, err)
	}

	var admin, postgres, standby []byte

	if bytepass, exists := s.Data["admin"]; exists {
		admin = bytepass
	} else {
		return fmt.Errorf("could not get TimescaleDB password: %w", errors.New("user not found"))
	}

	if bytepass, exists := s.Data["postgres"]; exists {
		postgres = bytepass
	} else {
		return fmt.Errorf("could not get TimescaleDB password: %w", errors.New("user not found"))
	}

	if bytepass, exists := s.Data["standby"]; exists {
		standby = bytepass
	} else {
		return fmt.Errorf("could not get TimescaleDB password: %w", errors.New("user not found"))
	}

	sec := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      newDBsecret,
			Namespace: root.Namespace,
			Labels:    utils.GetTimescaleDBsecretLabels(),
		},
		Data: map[string][]byte{
			"PATRONI_REPLICATION_PASSWORD": standby,
			"PATRONI_admin_PASSWORD":       admin,
			"PATRONI_SUPERUSER_PASSWORD":   postgres,
		},
		Type: "Opaque",
	}

	err = k8s.CreateSecret(sec)
	if err != nil {
		return fmt.Errorf("failed to create secret during the upgrade %s: %v", sec.Name, err)
	}

	fmt.Printf("secret/%s created\n\n", newDBsecret)

	return nil
}

func copyDBCertificate() error {
	fmt.Println("Migrating the certificate")
	existingDBCertificate := root.HelmReleaseName + "-timescaledb-certificate"
	newDBCertificate := root.HelmReleaseName + "-certificate"

	s, err := k8s.KubeGetSecret(root.Namespace, existingDBCertificate)
	if err != nil {
		return fmt.Errorf("failed to get the secret existing secret during the upgrade %s: %v", existingDBCertificate, err)
	}

	sec := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      newDBCertificate,
			Namespace: root.Namespace,
			Labels:    utils.GetTimescaleDBsecretLabels(),
		},
		Data: s.Data,
		Type: "Opaque",
	}

	err = k8s.CreateSecret(sec)
	if err != nil {
		return fmt.Errorf("failed to create secret during the upgrade %s: %v", sec.Name, err)
	}

	fmt.Printf("secret/%s created\n\n", newDBCertificate)

	return nil
}

func copyDBBackup() error {
	existingDBBackup := root.HelmReleaseName + "-timescaledb-pgbackrest"
	newDBBackUp := root.HelmReleaseName + "-pgbackrest"

	s, err := k8s.KubeGetSecret(root.Namespace, existingDBBackup)
	if err != nil {
		// backup is optional is disabled skip backup secret copying
		ok := errors2.IsNotFound(err)
		if !ok {
			return fmt.Errorf("failed to get the secret existing secret during the upgrade %s: %v", existingDBBackup, err)
		}
		return nil
	}

	fmt.Println("Migrating backup configuration")

	var pgBackRestConf string
	if bytepass, exists := s.Data["pgbackrest.conf"]; exists {
		pgBackRestConf = string(bytepass)
	} else {
		return fmt.Errorf("could not get TimescaleDB pgbackrest.conf in secret %s as backup is enabled", existingDBBackup)
	}

	s3Dataset := parsePgBackRestConf(pgBackRestConf)
	newData := make(map[string][]byte)
	for key, value := range s3Dataset {
		newKey := strings.Replace(key, "-", "_", -1)
		newKey = strings.ToUpper(newKey)
		newKey = "PGBACKREST_" + newKey
		newData[newKey] = []byte(value)
	}

	sec := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      newDBBackUp,
			Namespace: root.Namespace,
			Labels:    utils.GetTimescaleDBsecretLabels(),
		},
		Data: newData,
		Type: "Opaque",
	}

	err = k8s.CreateSecret(sec)
	if err != nil {
		return fmt.Errorf("failed to create secret during the upgrade %s: %v", sec.Name, err)
	}

	fmt.Printf("secret/%s created\n\n", newDBBackUp)

	return nil
}

// in older version of timescaleDB pgbackrest conf is set
// in string we need to break the string into key/value only
// *-s3-* keys
func parsePgBackRestConf(data string) map[string]string {
	newData := make(map[string]string)
	dataSets := strings.Split(data, "\n")
	for _, s := range dataSets {
		if strings.Contains(s, "-s3-") {
			d := strings.Split(s, "=")
			// we only care key/value pairs
			if len(d) == 2 {
				newData[d[0]] = d[1]
			}
		}
	}

	return newData
}

var crdURLs = []string{
"https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/v0.47.0/example/prometheus-operator-crd/monitoring.coreos.com_alertmanagerconfigs.yaml",
"https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/v0.47.0/example/prometheus-operator-crd/monitoring.coreos.com_alertmanagers.yaml",
"https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/v0.47.0/example/prometheus-operator-crd/monitoring.coreos.com_podmonitors.yaml",
"https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/v0.47.0/example/prometheus-operator-crd/monitoring.coreos.com_probes.yaml",
"https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/v0.47.0/example/prometheus-operator-crd/monitoring.coreos.com_prometheuses.yaml",
"https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/v0.47.0/example/prometheus-operator-crd/monitoring.coreos.com_servicemonitors.yaml",
"https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/v0.47.0/example/prometheus-operator-crd/monitoring.coreos.com_thanosrulers.yaml",
"https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/v0.47.0/example/prometheus-operator-crd/monitoring.coreos.com_prometheusrules.yaml",
}

var crdNames = []string{
	"alertmanagerconfigs.monitoring.coreos.com",
	"alertmanagers.monitoring.coreos.com",
	"podmonitors.monitoring.coreos.com",
	"probes.monitoring.coreos.com",
	"prometheuses.monitoring.coreos.com",
	"prometheusrules.monitoring.coreos.com",
	"servicemonitors.monitoring.coreos.com",
	"thanosrulers.monitoring.coreos.com",
}

func createCRDS() error {
	err := k8s.CreateCRDS(crdURLs)
	if err != nil {
		return err
	}
	fmt.Println("Successfully created CRDs: ", crdNames)
	return nil
}