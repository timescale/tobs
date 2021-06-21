package helm

import (
	"context"
	"fmt"
	"log"

	"github.com/timescale/tobs/cli/cmd"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/repo"
)

func getClient() Client {
	opt := &Options{
		Namespace: cmd.Namespace,
		Linting:   true,
	}

	helmClient, err := New(opt)
	if err != nil {
		log.Fatalf("failed to create helm client: %v", err)
	}
	return helmClient
}

func AddUpdateChart(chartName, repoURL string) error {
	c := getClient()
	r := repo.Entry{
		Name: chartName,
		URL:  repoURL,
	}
	err := c.AddOrUpdateChartRepo(r)
	if err != nil {
		return fmt.Errorf("failed to add or update chart repo %w", err)
	}
	return err
}

func InstallUpgradeChart(chartSpec *ChartSpec) (*release.Release, error) {
	c := getClient()
	release, err := c.InstallOrUpgradeChart(context.Background(), chartSpec)
	if err != nil {
		return release, fmt.Errorf("failed to install chart %w", err)
	}
	return release, nil
}

func UninstallChart(chartSpec *ChartSpec) error {
	c := getClient()
	if err := c.UninstallRelease(chartSpec); err != nil {
		return fmt.Errorf("failed to uninstall chart %w", err)
	}
	return nil
}

func ListReleases() ([]*release.Release, error) {
	c := getClient()
	return c.ListDeployedReleases()
}

func GetAllValuesFromRelease(releaseName string) (map[string]interface{}, error) {
	c := getClient()
	values, err := c.GetAllReleaseValues(releaseName)
	if err != nil {
		return values, fmt.Errorf("failed to get the values from the release %w", err)
	}
	return values, err
}

func GetValuesFromRelease(releaseName string) (map[string]interface{}, error) {
	c := getClient()
	values, err := c.GetReleaseValues(releaseName)
	if err != nil {
		return values, fmt.Errorf("failed to get the values from the release %w", err)
	}
	return values, err
}

func GetValuesFromChart(chartName string) ([]byte, error) {
	c := getClient()
	values, err := c.GetChartValues(chartName)
	if err != nil {
		return values, fmt.Errorf("failed to get the values from the release %w", err)
	}
	return values, err
}

func InspectChartYaml(chartName string) ([]byte, error) {
	c := getClient()
	values, err := c.InspectChartYaml(chartName)
	if err != nil {
		return values, fmt.Errorf("failed to get the values from the release %w", err)
	}
	return values, err
}
