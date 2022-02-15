package helm

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/timescale/tobs/cli/pkg/utils"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/repo"
	"helm.sh/helm/v3/pkg/storage/driver"
	apiextensionsV1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// we use .helmcache to maintain tobs helm chart cache data
// .helmrepo file contains repo metadata i.e. similar to repositories.yaml in helm CLI
// by using our own below specified paths we do not interfere with helm CLI if any installed
const (
	defaultCachePath            = "/tmp/.helmcache"
	defaultRepositoryConfigPath = "/tmp/.helmrepo"
)

// Client defines the values of a helm client
type clientImpl struct {
	settings          *cli.EnvSettings
	providers         getter.Providers
	storage           *repo.File
	actionConfig      *action.Configuration
	linting           bool
	namespace         string
	oldNamespaceValue string
}

// Options defines the options of a client
type ClientOptions struct {
	Namespace        string
	RepositoryConfig string
	RepositoryCache  string
	Debug            bool
	Linting          bool
	DebugLog         action.DebugLog
}

func NewClient(namespace string) Client {
	opt := &ClientOptions{
		Namespace:        namespace,
		RepositoryConfig: defaultRepositoryConfigPath,
		RepositoryCache:  defaultCachePath,
		Linting:          true,
	}

	// set helm namespace in env variable as
	// a workaround for Promscale & TimescaleDB
	// helm charts as they do not include namespace
	// in the template configuration & helm client
	// expects the namespace set in env if not
	// configured in template level
	namespaceFromEnv := setNamespaceInEnv(namespace)

	helmClient, err := New(opt)
	if err != nil {
		log.Fatalf("failed to create helm client: %v", err)
	}
	helmClient.oldNamespaceValue = namespaceFromEnv
	return helmClient
}

// New returns a new Helm client with the provided options
func New(options *ClientOptions) (*clientImpl, error) {
	settings := cli.New()
	return newClient(options, settings.RESTClientGetter(), settings)
}

// setEnvSettings sets the client's environment settings
func setNamespaceInEnv(namespace string) string {
	// capture the old namespace from env vars
	oldNamespaceValue := os.Getenv("HELM_NAMESPACE")

	// set the namespace with this ugly workaround because cli.EnvSettings.namespace is private
	// thank you helm!
	err := os.Setenv("HELM_NAMESPACE", namespace)
	if err != nil {
		log.Fatalf("failed to set HELM_NAMESPACE env variable %v", err)
	}
	return oldNamespaceValue
}

// after install/upgrade we need to unset helm namespace set by us
// to make sure we don't leave unwanted envs set.
func (c *clientImpl) unSetHelmNamespaceEnv() {
	err := os.Setenv("HELM_NAMESPACE", c.oldNamespaceValue)
	if err != nil {
		log.Fatalf("Error: failed to unset HELM_NAMESPACE env variable to '' : %v\n", err)
	}
}

// newClient returns a new Helm client via the provided options and REST config
func newClient(options *ClientOptions, clientGetter genericclioptions.RESTClientGetter, settings *cli.EnvSettings) (*clientImpl, error) {
	debugLog := func(_ string, _ ...interface{}) {}
	if options.DebugLog != nil {
		debugLog = options.DebugLog
	}

	actionConfig := new(action.Configuration)
	err := actionConfig.Init(
		clientGetter,
		options.Namespace,
		os.Getenv("HELM_DRIVER"),
		debugLog,
	)
	if err != nil {
		return nil, err
	}

	settings.RepositoryCache = defaultCachePath
	settings.RepositoryConfig = defaultRepositoryConfigPath

	return &clientImpl{
		settings:     settings,
		providers:    getter.All(settings),
		actionConfig: actionConfig,
		storage:      &repo.File{},
		linting:      options.Linting,
		namespace:    options.Namespace,
	}, nil
}

func (c *clientImpl) GetChartMetadata(chart string) (*ChartMetadata, error) {
	chartDetails := &ChartMetadata{}
	client := action.NewInstall(c.actionConfig)
	helmChart, _, err := c.getChart(chart, &client.ChartPathOptions)
	if err != nil {
		return nil, err
	}

	for _, k := range helmChart.Raw {
		if k.Name == "Chart.yaml" {
			err = yaml.Unmarshal(k.Data, chartDetails)
			if err != nil {
				return chartDetails, fmt.Errorf("failed to unmarshal helm chart metadata %w", err)
			}

			return chartDetails, nil
		}
	}

	return nil, fmt.Errorf("failed to get Chart.yaml from the provided chart")
}

func (c *clientImpl) GetDeployedChartMetadata(releaseName, namespace string) (*DeployedChartMetadata, error) {
	var charts []DeployedChartMetadata
	l, err := c.listDeployedReleases(namespace)
	if err != nil {
		return nil, err
	}
	for _, k := range l {
		d := DeployedChartMetadata{
			Name:       k.Name,
			Namespace:  k.Namespace,
			Updated:    k.Info.LastDeployed.String(),
			Status:     k.Info.Status.String(),
			Chart:      k.Chart.Name(),
			AppVersion: k.Chart.Metadata.AppVersion,
			Version:    k.Chart.Metadata.Version,
		}
		charts = append(charts, d)
	}

	for _, c := range charts {
		if c.Name == releaseName {
			return &c, nil
		}
	}

	return nil, utils.ErrorTobsDeploymentNotFound(releaseName)
}

func (c *clientImpl) ExportValuesFieldFromChart(chart string, customValuesFile string, keys []string) (interface{}, error) {
	res, err := c.GetValuesYamlFromChart(chart, customValuesFile)
	if err != nil {
		return nil, fmt.Errorf("failed to do helm show values on the helm chart %w", err)
	}
	return FetchValue(res, keys)
}

func (c *clientImpl) ExportValuesFieldFromRelease(releaseName string, keys []string) (interface{}, error) {
	res, err := c.GetAllReleaseValues(releaseName)
	if err != nil {
		return nil, fmt.Errorf("failed to do helm get values from the helm release %w", err)
	}
	return FetchValue(res, keys)
}

func (c *clientImpl) GetValuesYamlFromChart(chart, file string) (interface{}, error) {
	var res []byte
	var err error
	if file != "" {
		res, err = ioutil.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("unable to read values from provided file %w", err)
		}
	} else {
		res, err = c.GetChartValues(chart)
		if err != nil {
			return nil, err
		}
	}

	var i interface{}
	err = yaml.Unmarshal(res, &i)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal existing values.yaml file %w", err)
	}

	values := ConvertMapI2MapS(i)
	return values, nil
}

func (c *clientImpl) Close() {
	c.unSetHelmNamespaceEnv()
}

// AddOrUpdateChartRepo adds or updates the provided helm chart repository
func (c *clientImpl) AddOrUpdateChartRepo(registryName, repoURL string) error {
	entry := repo.Entry{
		Name: registryName,
		URL:  repoURL,
	}
	chartRepo, err := repo.NewChartRepository(&entry, c.providers)
	if err != nil {
		return err
	}

	chartRepo.CachePath = c.settings.RepositoryCache
	_, err = chartRepo.DownloadIndexFile()
	if err != nil {
		return err
	}

	if c.storage.Has(entry.Name) {
		log.Printf("WARNING: repository name %q already exists", entry.Name)
		return nil
	}

	c.storage.Update(&entry)
	err = c.storage.WriteFile(c.settings.RepositoryConfig, 0o644)
	if err != nil {
		return err
	}

	return nil
}

// InstallOrUpgradeChart triggers the installation of the provided chart.
// If the chart is already installed, trigger an upgrade instead
func (c *clientImpl) InstallOrUpgradeChart(ctx context.Context, spec *ChartSpec) (*release.Release, error) {
	releaseInfo := &release.Release{}
	installed, err := c.chartIsInstalled(spec.ReleaseName)
	if err != nil {
		return releaseInfo, err
	}

	if installed {
		return releaseInfo, c.upgrade(ctx, spec)
	}

	return c.install(spec)
}

// GetReleaseValues returns the all computed values for the specified release.
func (c *clientImpl) GetAllReleaseValues(name string) (map[string]interface{}, error) {
	getReleaseValuesClient := action.NewGetValues(c.actionConfig)
	getReleaseValuesClient.AllValues = true
	return getReleaseValuesClient.Run(name)
}

// GetReleaseValues returns the values for the specified release.
func (c *clientImpl) GetReleaseValues(name string) (map[string]interface{}, error) {
	getReleaseValuesClient := action.NewGetValues(c.actionConfig)
	return getReleaseValuesClient.Run(name)
}

// GetChartValues returns the values from chart.
func (c *clientImpl) GetChartValues(name string) ([]byte, error) {
	client := action.NewInstall(c.actionConfig)
	helmChart, _, err := c.getChart(name, &client.ChartPathOptions)
	if err != nil {
		return nil, err
	}

	for _, k := range helmChart.Raw {
		if k.Name == "values.yaml" {
			return k.Data, nil
		}
	}

	return nil, fmt.Errorf("failed to get values from the provided chart")
}

// UninstallRelease uninstalls the provided release
func (c *clientImpl) UninstallRelease(spec *ChartSpec) error {
	client := action.NewUninstall(c.actionConfig)

	setUninstallReleaseOptions(spec, client)

	_, err := client.Run(spec.ReleaseName)
	if err != nil {
		return err
	}
	return nil
}

// install lints and installs the provided chart
func (c *clientImpl) install(spec *ChartSpec) (*release.Release, error) {
	release := &release.Release{}
	client := action.NewInstall(c.actionConfig)
	setInstallOptions(spec, client)
	helmChart, chartPath, err := c.getChart(spec.ChartName, &client.ChartPathOptions)
	if err != nil {
		return release, err
	}

	if helmChart.Metadata.Type != "" && helmChart.Metadata.Type != "application" {
		return release, fmt.Errorf(
			"chart %q has an unsupported type and is not installable: %q",
			helmChart.Metadata.Name,
			helmChart.Metadata.Type,
		)
	}

	values, err := spec.GetValuesMap()
	if err != nil {
		return release, err
	}

	if c.linting {
		err = c.lint(chartPath, values)
		if err != nil {
			return release, err
		}
	}

	release, err = client.Run(helmChart, values)
	if err != nil {
		return release, err
	}

	log.Printf("release installed successfully: %s/%s-%s", release.Name, release.Name, release.Chart.Metadata.Version)

	return release, nil
}

// upgrade upgrades a chart and CRDs
func (c *clientImpl) upgrade(ctx context.Context, spec *ChartSpec) error {
	client := action.NewUpgrade(c.actionConfig)
	setUpgradeOptions(spec, client)
	helmChart, chartPath, err := c.getChart(spec.ChartName, &client.ChartPathOptions)
	if err != nil {
		return err
	}

	values, err := spec.GetValuesMap()
	if err != nil {
		return err
	}

	if c.linting {
		err = c.lint(chartPath, values)
		if err != nil {
			return err
		}
	}

	if !spec.SkipCRDs && spec.UpgradeCRDs {
		log.Printf("updating crds")
		err = c.upgradeCRDs(ctx, helmChart)
		if err != nil {
			return err
		}
	}

	rel, err := client.Run(spec.ReleaseName, helmChart, values)
	if err != nil {
		return err
	}

	log.Printf("release upgrade successfully: %s/%s-%s", rel.Name, rel.Name, rel.Chart.Metadata.Version)

	return nil
}

// lint lints a chart's values
func (c *clientImpl) lint(chartPath string, values map[string]interface{}) error {
	client := action.NewLint()

	result := client.Run([]string{chartPath}, values)

	for _, err := range result.Errors {
		log.Printf("Error %s", err)
	}

	if len(result.Errors) > 0 {
		return fmt.Errorf("linting for chartpath %q failed", chartPath)
	}

	return nil
}

// upgradeCRDs upgrades the CRDs of the provided chart
func (c *clientImpl) upgradeCRDs(ctx context.Context, chartInstance *chart.Chart) error {
	cfg, err := c.settings.RESTClientGetter().ToRESTConfig()
	if err != nil {
		return err
	}

	k8sClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		return err
	}

	for _, crd := range chartInstance.CRDObjects() {
		// use this ugly detour to parse the crdYaml to a CustomResourceDefinitions-Object because direct
		// yaml-unmarshalling does not find the correct keys
		// this code is taken from https://github.com/mittwald/go-helm-client/blob/master/client.go#L517
		jsonCRD, err := yaml.ToJSON(crd.File.Data)
		if err != nil {
			return err
		}

		var meta metaV1.TypeMeta
		err = json.Unmarshal(jsonCRD, &meta)
		if err != nil {
			return err
		}

		switch meta.APIVersion {

		case "apiextensions.k8s.io/apiextensionsV1":
			var crdObj apiextensionsV1.CustomResourceDefinition
			err = json.Unmarshal(jsonCRD, &crdObj)
			if err != nil {
				return err
			}
			existingCRDObj, err := k8sClient.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, crdObj.Name, metaV1.GetOptions{})
			if err != nil {
				return err
			}
			crdObj.ResourceVersion = existingCRDObj.ResourceVersion
			_, err = k8sClient.ApiextensionsV1().CustomResourceDefinitions().Update(ctx, &crdObj, metaV1.UpdateOptions{})
			if err != nil {
				return err
			}

		case "apiextensions.k8s.io/v1beta1":
			var crdObj v1beta1.CustomResourceDefinition
			err = json.Unmarshal(jsonCRD, &crdObj)
			if err != nil {
				return err
			}
			existingCRDObj, err := k8sClient.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, crdObj.Name, metaV1.GetOptions{})
			if err != nil {
				return err
			}
			crdObj.ResourceVersion = existingCRDObj.ResourceVersion
			_, err = k8sClient.ApiextensionsV1beta1().CustomResourceDefinitions().Update(ctx, &crdObj, metaV1.UpdateOptions{})
			if err != nil {
				return err
			}

		case "apiextensions.k8s.io/v1":
			var crdObj apiextensionsV1.CustomResourceDefinition
			err = json.Unmarshal(jsonCRD, &crdObj)
			if err != nil {
				return err
			}
			existingCRDObj, err := k8sClient.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, crdObj.Name, metaV1.GetOptions{})
			if err != nil {
				return err
			}
			crdObj.ResourceVersion = existingCRDObj.ResourceVersion
			_, err = k8sClient.ApiextensionsV1().CustomResourceDefinitions().Update(ctx, &crdObj, metaV1.UpdateOptions{})
			if err != nil {
				return err
			}

		default:
			return fmt.Errorf("failed to update crd %q: unsupported api-version %q", crd.Name, meta.APIVersion)
		}
	}

	return nil
}

// getChart returns a chart matching the provided chart name and options
func (c *clientImpl) getChart(chartName string, chartPathOptions *action.ChartPathOptions) (*chart.Chart, string, error) {
	chartPath, err := chartPathOptions.LocateChart(chartName, c.settings)
	if err != nil {
		return nil, "", err
	}

	helmChart, err := loader.Load(chartPath)
	if err != nil {
		return nil, "", err
	}

	if helmChart.Metadata.Deprecated {
		log.Printf("WARNING: This chart (%q) is deprecated", helmChart.Metadata.Name)
	}

	return helmChart, chartPath, err
}

// chartIsInstalled checks whether a chart is already installed or not by the provided release name
func (c *clientImpl) chartIsInstalled(release string) (bool, error) {
	histClient := action.NewHistory(c.actionConfig)
	histClient.Max = 1
	if _, err := histClient.Run(release); err != nil {
		if err == driver.ErrReleaseNotFound {
			err = nil
		}
		return false, err
	}

	return true, nil
}

// mergeInstallOptions merges values of the provided chart to helm install options used by the client
func setInstallOptions(chartSpec *ChartSpec, installOptions *action.Install) {
	installOptions.DisableHooks = chartSpec.DisableHooks
	installOptions.Replace = chartSpec.Replace
	installOptions.Wait = chartSpec.Wait
	installOptions.Timeout = chartSpec.Timeout
	installOptions.Namespace = chartSpec.Namespace
	installOptions.ReleaseName = chartSpec.ReleaseName
	installOptions.Version = chartSpec.Version
	installOptions.GenerateName = chartSpec.GenerateName
	installOptions.NameTemplate = chartSpec.NameTemplate
	installOptions.Atomic = chartSpec.Atomic
	installOptions.SkipCRDs = chartSpec.SkipCRDs
	installOptions.DryRun = chartSpec.DryRun
	installOptions.SubNotes = chartSpec.SubNotes
	installOptions.CreateNamespace = chartSpec.CreateNamespace
}

// mergeUpgradeOptions merges values of the provided chart to helm upgrade options used by the client
func setUpgradeOptions(chartSpec *ChartSpec, upgradeOptions *action.Upgrade) {
	upgradeOptions.Version = chartSpec.Version
	upgradeOptions.Namespace = chartSpec.Namespace
	upgradeOptions.Timeout = chartSpec.Timeout
	upgradeOptions.Wait = chartSpec.Wait
	upgradeOptions.DisableHooks = chartSpec.DisableHooks
	upgradeOptions.Force = chartSpec.Force
	upgradeOptions.ResetValues = chartSpec.ResetValues
	upgradeOptions.ReuseValues = chartSpec.ReuseValues
	upgradeOptions.Recreate = chartSpec.Recreate
	upgradeOptions.MaxHistory = chartSpec.MaxHistory
	upgradeOptions.Atomic = chartSpec.Atomic
	upgradeOptions.CleanupOnFail = chartSpec.CleanupOnFail
	upgradeOptions.DryRun = chartSpec.DryRun
	upgradeOptions.SubNotes = chartSpec.SubNotes
}

// mergeUninstallReleaseOptions merges values of the provided chart to helm uninstall options used by the client
func setUninstallReleaseOptions(chartSpec *ChartSpec, uninstallReleaseOptions *action.Uninstall) {
	uninstallReleaseOptions.DisableHooks = chartSpec.DisableHooks
	uninstallReleaseOptions.Timeout = chartSpec.Timeout
}

// ListDeployedReleases lists all deployed releases.
// Namespace and other context is provided via the Options struct when instantiating a client.
func (c *clientImpl) listDeployedReleases(namespace string) ([]*release.Release, error) {
	err := c.actionConfig.Init(c.settings.RESTClientGetter(), namespace, "", func(_ string, _ ...interface{}) {})
	if err != nil {
		return nil, fmt.Errorf("failed to list deployed releases %v", err)
	}
	listClient := action.NewList(c.actionConfig)
	listClient.StateMask = action.ListDeployed | action.ListFailed | action.ListPendingInstall | action.ListPendingUpgrade
	return listClient.Run()
}
