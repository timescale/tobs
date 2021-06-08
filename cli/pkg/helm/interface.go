package helm

import (
	"context"

	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/repo"
)

//go:generate mockgen -source=interface.go -package mockhelmclient -destination=./mock/interface.go -self_package=. Client

type Client interface {
	AddOrUpdateChartRepo(entry repo.Entry) error
	UpdateChartRepos() error
	InstallOrUpgradeChart(ctx context.Context, spec *ChartSpec) (*release.Release, error)
	ListDeployedReleases() ([]*release.Release, error)
	GetReleaseValues(name string, allValues bool) (map[string]interface{}, error)
	GetChartValues(name string) ([]byte, error)
	InspectChartYaml(name string) ([]byte, error)
	UninstallRelease(spec *ChartSpec) error
}
