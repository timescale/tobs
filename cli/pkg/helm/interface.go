package helm

import (
	"context"

	"helm.sh/helm/v3/pkg/release"
)

type Client interface {
	AddOrUpdateChartRepo(registryName, repoURL string) error
	UpdateChartRepos() error
	InstallOrUpgradeChart(ctx context.Context, spec *ChartSpec) (*release.Release, error)
	ListDeployedReleases() ([]*release.Release, error)
	GetAllReleaseValues(name string) (map[string]interface{}, error)
	GetReleaseValues(name string) (map[string]interface{}, error)
	GetChartValues(name string) ([]byte, error)
	InspectChartYaml(name string) ([]byte, error)
	UninstallRelease(spec *ChartSpec) error
}
