package dependency_tests

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"testing"

	"github.com/timescale/tobs/cli/cmd/upgrade"
	"github.com/timescale/tobs/cli/pkg/helm"
	"sigs.k8s.io/yaml"
)

func TestMain(m *testing.M) {
	validateKubePrometheusVersions()
}

func validateKubePrometheusVersions() {
	// Get existing tobs helm chart metadata
	b, err := ioutil.ReadFile("./../../../chart/Chart.yaml")
	if err != nil {
		log.Fatal(err)
	}
	existingTobsChart := &helm.ChartMetadata{}
	err = yaml.Unmarshal(b, existingTobsChart)
	if err != nil {
		log.Fatal(err)
	}
	var kubePrometheusVersion string
	for _, i := range existingTobsChart.Dependencies {
		if i.Name == "kube-prometheus-stack" {
			kubePrometheusVersion = i.Version
			break
		}
	}

	// Get upstream kube-prometheus chart metadata using kube-prometheus version used in tobs local Chart.yaml
	resp, err := http.Get("https://raw.githubusercontent.com/prometheus-community/helm-charts/kube-prometheus-stack-" + kubePrometheusVersion + "/charts/kube-prometheus-stack/Chart.yaml")
	if err != nil {
		log.Fatalf("failed to get the kube-prometheus CHart.yaml info %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	upstreamKPChart := &helm.ChartMetadata{}
	err = yaml.Unmarshal(bodyBytes, upstreamKPChart)
	if err != nil {
		log.Fatal(err)
	}

	upstreamKPChart.AppVersion = "v" + upstreamKPChart.AppVersion
	// validate existing tobs kube-prometheus helm chart version & CRDs version with upstream version and CRDs that are being used
	if upstreamKPChart.Version != kubePrometheusVersion || upstreamKPChart.AppVersion != upgrade.KubePrometheusCRDVersion {
		log.Fatalf("failed to validate tobs kube-prometheus helm chart version and CRDs version with upstream versions."+
			"Mismatch in validation, tobs Kube-Prometheus version: %s, tobs kube-prometheus CRD version: %s, "+
			"upstream kube-prometheus CRD version: %s", kubePrometheusVersion, upgrade.KubePrometheusCRDVersion, upstreamKPChart.AppVersion)
	}
	fmt.Println("successfully validated kube-prometheus CRD versions with upstream versions.")
}
