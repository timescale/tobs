package helm

import (
	"context"

	"helm.sh/helm/v3/pkg/release"
)

type Client interface {
	AddOrUpdateChartRepo(rehistryName, repoURL string) error
	InstallOrUpgradeChart(ctx context.Context, spec *ChartSpec) (*release.Release, error)
	GetAllReleaseValues(name string) (map[string]interface{}, error)
	GetReleaseValues(name string) (map[string]interface{}, error)
	GetChartValues(name string) ([]byte, error)
	UninstallRelease(spec *ChartSpec) error
	GetDeployedChartMetadata(releaseName string) (*DeployedChartMetadata, error)
	ExportValuesFieldFromRelease(releaseName string, keys []string) (interface{}, error)
	ExportValuesFieldFromChart(chart string, customValuesFile string, keys []string) (interface{}, error)
	GetChartMetadata(chart string) (*ChartMetadata, error)
	GetValuesYamlFromChart(chart, file string) (interface{}, error)
	Close()
}
