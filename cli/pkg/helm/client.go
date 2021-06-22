package helm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/spf13/pflag"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/downloader"
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

var storage = repo.File{}

func NewClient(namespace string) *ClientInfo {
	opt := &ClientOptions{
		Namespace: namespace,
		Linting:   true,
	}

	helmClient, err := New(opt)
	if err != nil {
		log.Fatalf("failed to create helm client: %v", err)
	}
	return helmClient
}

// Client defines the values of a helm client
type ClientInfo struct {
	settings     *cli.EnvSettings
	providers    getter.Providers
	storage      *repo.File
	actionConfig *action.Configuration
	linting      bool
}

// ClientOptions defines the options of a client
type ClientOptions struct {
	Namespace        string
	RepositoryConfig string
	RepositoryCache  string
	Debug            bool
	Linting          bool
	DebugLog         action.DebugLog
}

// New returns a new Helm client with the provided options
func New(options *ClientOptions) (*ClientInfo, error) {
	settings := cli.New()
	return newClient(options, settings.RESTClientGetter(), settings)
}

// newClient returns a new Helm client via the provided options and REST config
func newClient(options *ClientOptions, clientGetter genericclioptions.RESTClientGetter, settings *cli.EnvSettings) (*ClientInfo, error) {
	err := setEnvSettings(options, settings)
	if err != nil {
		return nil, err
	}

	debugLog := func(_ string, _ ...interface{}) {}
	if options.DebugLog != nil {
		debugLog = options.DebugLog
	}

	actionConfig := new(action.Configuration)
	err = actionConfig.Init(
		clientGetter,
		settings.Namespace(),
		os.Getenv("HELM_DRIVER"),
		debugLog,
	)
	if err != nil {
		return nil, err
	}
	return &ClientInfo{
		settings:     settings,
		providers:    getter.All(settings),
		storage:      &storage,
		actionConfig: actionConfig,
		linting:      options.Linting,
	}, nil
}

// setEnvSettings sets the client's environment settings based on the provided client configuration
func setEnvSettings(options *ClientOptions, settings *cli.EnvSettings) error {
	if options == nil {
		options = &ClientOptions{}
	}

	// set the namespace with this ugly workaround because cli.EnvSettings.namespace is private
	// thank you helm!
	if options.Namespace != "" {
		pflags := pflag.NewFlagSet("", pflag.ContinueOnError)
		settings.AddFlags(pflags)
		err := pflags.Parse([]string{"-n", options.Namespace})
		if err != nil {
			return err
		}
	}

	return nil
}

// AddOrUpdateChartRepo adds or updates the provided helm chart repository
func (c *ClientInfo) AddOrUpdateChartRepo(registryName, repoURL string) error {
	entry := repo.Entry{
		Name:                  registryName,
		URL:                   repoURL,
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
		fmt.Println("hello")
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

// UpdateChartRepos updates the list of chart repositories stored in the client's cache
func (c *ClientInfo) UpdateChartRepos() error {
	for _, entry := range c.storage.Repositories {
		chartRepo, err := repo.NewChartRepository(entry, c.providers)
		if err != nil {
			return err
		}

		chartRepo.CachePath = c.settings.RepositoryCache
		_, err = chartRepo.DownloadIndexFile()
		if err != nil {
			return err
		}

		c.storage.Update(entry)
	}

	return c.storage.WriteFile(c.settings.RepositoryConfig, 0o644)
}

// InstallOrUpgradeChart triggers the installation of the provided chart.
// If the chart is already installed, trigger an upgrade instead
func (c *ClientInfo) InstallOrUpgradeChart(ctx context.Context, spec *ChartSpec) (*release.Release, error) {
	release := &release.Release{}
	installed, err := c.chartIsInstalled(spec.ReleaseName)
	if err != nil {
		return release, err
	}

	if installed {
		return release, c.upgrade(ctx, spec)
	}
	return c.install(spec)
}

// install lints and installs the provided chart
func (c *ClientInfo) install(spec *ChartSpec) (*release.Release, error) {
	release := &release.Release{}
	client := action.NewInstall(c.actionConfig)
	mergeInstallOptions(spec, client)

	if client.Version == "" {
		client.Version = ">0.0.0-0"
	}

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

	d := &updateDeps{
		helmChart:        helmChart,
		spec:             spec,
		chartPath:        &chartPath,
		chartPathOptions: &client.ChartPathOptions,
	}
	err = c.updateDependencies(d)
	if err != nil {
		return release, fmt.Errorf("failed to update helm chart dependencies %v", err)
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

	client.Namespace = c.settings.Namespace()
	// create namespace if it doesn't exist
	client.CreateNamespace = spec.CreateNamespace
	release, err = client.Run(helmChart, values)
	if err != nil {
		return release, err
	}

	log.Printf("release installed successfully: %s/%s-%s", release.Name, release.Name, release.Chart.Metadata.Version)

	return release, nil
}

type updateDeps struct {
	helmChart        *chart.Chart
	spec             *ChartSpec
	chartPath        *string
	chartPathOptions *action.ChartPathOptions
}

func (c *ClientInfo) updateDependencies(details *updateDeps) error {
	if req := details.helmChart.Metadata.Dependencies; req != nil {
		if err := action.CheckDependencies(details.helmChart, req); err != nil {
			if !details.spec.DependencyUpdate {
				return err
			} else {
				man := &downloader.Manager{
					ChartPath:        *details.chartPath,
					Keyring:          details.chartPathOptions.Keyring,
					SkipUpdate:       false,
					Getters:          c.providers,
					RepositoryConfig: c.settings.RepositoryConfig,
					RepositoryCache:  c.settings.RepositoryCache,
					Out:              os.Stdout,
				}
				if err := man.Update(); err != nil {
					return err
				}

				// as chart dependencies are updated fetch the chart again
				details.helmChart, *details.chartPath, err = c.getChart(details.spec.ChartName, details.chartPathOptions)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// upgrade upgrades a chart and CRDs
func (c *ClientInfo) upgrade(ctx context.Context, spec *ChartSpec) error {
	client := action.NewUpgrade(c.actionConfig)
	mergeUpgradeOptions(spec, client)

	if client.Version == "" {
		fmt.Println("hii-hello....")
		client.Version = ">0.0.0-0"
	}

	helmChart, chartPath, err := c.getChart(spec.ChartName, &client.ChartPathOptions)
	if err != nil {
		return err
	}

	d := &updateDeps{
		helmChart:        helmChart,
		spec:             spec,
		chartPath:        &chartPath,
		chartPathOptions: &client.ChartPathOptions,
	}
	err = c.updateDependencies(d)
	if err != nil {
		return fmt.Errorf("failed to update helm chart dependencies %v", err)
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

// uninstallRelease uninstalls the provided release
func (c *ClientInfo) UninstallRelease(spec *ChartSpec) error {
	client := action.NewUninstall(c.actionConfig)

	mergeUninstallReleaseOptions(spec, client)

	_, err := client.Run(spec.ReleaseName)
	if err != nil {
		return err
	}

	return nil
}

// lint lints a chart's values
func (c *ClientInfo) lint(chartPath string, values map[string]interface{}) error {
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
func (c *ClientInfo) upgradeCRDs(ctx context.Context, chartInstance *chart.Chart) error {
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
func (c *ClientInfo) getChart(chartName string, chartPathOptions *action.ChartPathOptions) (*chart.Chart, string, error) {
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
func (c *ClientInfo) chartIsInstalled(release string) (bool, error) {
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

// ListDeployedReleases lists all deployed releases.
// Namespace and other context is provided via the ClientOptions struct when instantiating a client.
func (c *ClientInfo) ListDeployedReleases() ([]*release.Release, error) {
	err := c.actionConfig.Init(c.settings.RESTClientGetter(), "", "", func(_ string, _ ...interface{}) {})
	if err != nil {
		return nil, fmt.Errorf("failed to list deployed releases %v", err)
	}
	listClient := action.NewList(c.actionConfig)
	listClient.StateMask = action.ListDeployed
	return listClient.Run()
}

// GetReleaseValues returns the all computed values for the specified release.
func (c *ClientInfo) GetAllReleaseValues(name string) (map[string]interface{}, error) {
	getReleaseValuesClient := action.NewGetValues(c.actionConfig)
	getReleaseValuesClient.AllValues = true
	return getReleaseValuesClient.Run(name)
}

// GetReleaseValues returns the values for the specified release.
func (c *ClientInfo) GetReleaseValues(name string) (map[string]interface{}, error) {
	getReleaseValuesClient := action.NewGetValues(c.actionConfig)
	return getReleaseValuesClient.Run(name)
}

// GetChartValues returns the values from chart.
func (c *ClientInfo) GetChartValues(name string) ([]byte, error) {
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

func (c *ClientInfo) InspectChartYaml(name string) ([]byte, error) {
	client := action.NewInstall(c.actionConfig)
	helmChart, _, err := c.getChart(name, &client.ChartPathOptions)
	if err != nil {
		return nil, err
	}

	for _, k := range helmChart.Raw {
		if k.Name == "Chart.yaml" {
			return k.Data, nil
		}
	}

	return nil, fmt.Errorf("failed to get Chart.yaml from the provided chart")
}

// mergeInstallOptions merges values of the provided chart to helm install options used by the client
func mergeInstallOptions(chartSpec *ChartSpec, installOptions *action.Install) {
	installOptions.DisableHooks = chartSpec.DisableHooks
	installOptions.Replace = chartSpec.Replace
	installOptions.Wait = chartSpec.Wait
	installOptions.DependencyUpdate = chartSpec.DependencyUpdate
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
}

// mergeUpgradeOptions merges values of the provided chart to helm upgrade options used by the client
func mergeUpgradeOptions(chartSpec *ChartSpec, upgradeOptions *action.Upgrade) {
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
func mergeUninstallReleaseOptions(chartSpec *ChartSpec, uninstallReleaseOptions *action.Uninstall) {
	uninstallReleaseOptions.DisableHooks = chartSpec.DisableHooks
	uninstallReleaseOptions.Timeout = chartSpec.Timeout
}
