package helm_tests

import (
	"context"
	"github.com/timescale/tobs/cli/pkg/utils"
	"testing"

	"github.com/timescale/tobs/cli/cmd"
	"github.com/timescale/tobs/cli/pkg/helm"
)

var helmClientTest helm.Client

func TestNewHelmClient(t *testing.T) {
	opt := &helm.ClientOptions{
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
	// Add a chart-repository to the client
	if err := helmClientTest.AddOrUpdateChartRepo(utils.DEFAULT_REGISTRY_NAME, utils.REPO_LOCATION); err != nil {
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
