package helm_tests

import (
	"context"
	"testing"

	"github.com/timescale/tobs/cli/cmd"
	"github.com/timescale/tobs/cli/pkg/helm"
	"helm.sh/helm/v3/pkg/repo"
)

var helmClientTest helm.Client

func TestNewHelmClient(t *testing.T) {
	opt := &helm.Options{
		Namespace: cmd.Namespace,
		Linting:   true,
	}

	var err error
	helmClientTest, err = helm.New(opt)
	if err != nil {
		t.Fatal(err)
	}
}

func TestHelmClientAddOrUpdateChartRepoPublic(t *testing.T) {
	// Define a public chart repository
	chartRepo := repo.Entry{
		Name: "timescale",
		URL:  "https://charts.timescale.com",
	}

	// Add a chart-repository to the client
	if err := helmClientTest.AddOrUpdateChartRepo(chartRepo); err != nil {
		t.Fatal(err)
	}
}

func TestHelmClientInstallOrUpgradeChart(t *testing.T) {
	// Define the chart to be installed
	chartSpec := helm.ChartSpec{
		ReleaseName:      "tobs",
		ChartName:        "timescale/tobs",
		Namespace:        "default",
		UpgradeCRDs:      true,
		DependencyUpdate: true,
		Version:          "0.4.0",
	}

	if _, err := helmClientTest.InstallOrUpgradeChart(context.Background(), &chartSpec); err != nil {
		t.Fatal(err)
	}
}

func TestHelmClientUpdateChartRepos(t *testing.T) {
	if err := helmClientTest.UpdateChartRepos(); err != nil {
		t.Fatal(err)
	}
}

func TestHelmClientGetReleaseValues(t *testing.T) {
	if _, err := helmClientTest.GetAllReleaseValues("tobs"); err != nil {
		t.Fatal(err)
	}
}

func TestHelmClientUninstallRelease(t *testing.T) {
	// Define the released chart to be installed
	chartSpec := helm.ChartSpec{
		ReleaseName: "tobs",
		ChartName:   "timescale/tobs",
		Namespace:   "default",
		UpgradeCRDs: true,
	}

	if err := helmClientTest.UninstallRelease(&chartSpec); err != nil {
		t.Fatal(err)
	}
}

func TestHelmClientListDeployedReleases(t *testing.T) {
	if _, err := helmClientTest.ListDeployedReleases(); err != nil {
		t.Fatal(err)
	}
}
