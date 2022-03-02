package otel

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/open-telemetry/opentelemetry-operator/apis/v1alpha1"
	"github.com/timescale/tobs/cli/pkg/helm"
	"github.com/timescale/tobs/cli/pkg/k8s"
	"github.com/timescale/tobs/cli/pkg/utils"
	errors2 "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	CertManagerVersion    = "v1.6.1"
	CertManagerNamespace  = "cert-manager"
	otelColKind           = "OpenTelemetryCollector"
	otelColApiVersion     = "opentelemetry.io/v1alpha1"
	otelOperatorNamespace = "opentelemetry-operator-system"
	otelColResourceName   = "opentelemetrycollectors"
)

var (
	certManagerManifests = map[string]string{
		"cert-manager": fmt.Sprintf("https://github.com/jetstack/cert-manager/releases/download/%s/cert-manager.yaml", CertManagerVersion),
	}

	otelColCRD = fmt.Sprintf("%s.opentelemetry.io", otelColResourceName)

	openTelemetryCRDVersion     = "v0.41.1"
	openTelemetryCRDsPathPrefix = fmt.Sprintf("https://raw.githubusercontent.com/open-telemetry/opentelemetry-operator/%s/bundle/manifests/", openTelemetryCRDVersion)
	OpenTelemetryCRDs           = map[string]string{
		"instrumentations.opentelemetry.io":        openTelemetryCRDsPathPrefix + "opentelemetry.io_instrumentations.yaml",
		"opentelemetrycollectors.opentelemetry.io": openTelemetryCRDsPathPrefix + "opentelemetry.io_opentelemetrycollectors.yaml",
	}
)

type OtelCol struct {
	ReleaseName string
	Namespace   string
	K8sClient   k8s.Client
	HelmClient  helm.Client
	UpgradeCM   bool
}

func CreateCertManager(confirmActions bool) error {
	apiClient := k8s.NewAPIClient()
	crd, err := apiClient.GetCRD("certificates.cert-manager.io")
	if err != nil {
		if errors2.IsNotFound(err) {
			fmt.Println("Couldn't find the cert-manager. The cert-manager is required to deploy OpenTelemetry. Do you want to install the cert-manager?")
			if !confirmActions {
				utils.ConfirmAction()
			}

			err = createUpgradeCertManager()
			if err != nil {
				return fmt.Errorf("failed to create cert-manager %v", err)
			}
			fmt.Println("Successfully created cert-manager")
			return nil
		}
	}

	// validate the version of cer-manager in cluster
	certManagerVersion := crd.Labels["app.kubernetes.io/version"]
	ok, err := isDesiredVersion(certManagerVersion, CertManagerVersion)
	if err != nil {
		return err
	}

	if !ok {
		return fmt.Errorf("existing cert-manager in the cluster with version %s doesn't comply with the"+
			" OpenTelemetry Operator, requires cert-manager version %s", certManagerVersion, CertManagerVersion)
	}

	fmt.Printf("cert-manager version %s exists in the cluster, skipping the installation...\n", certManagerVersion)
	return err
}

// this func check whether the current version >= desired version
func isDesiredVersion(current, desired string) (bool, error) {
	cV, err := utils.ParseVersion(strings.Replace(current, "v", "", 1), 3)
	if err != nil {
		return false, fmt.Errorf("failed to parse %s version %w", current, err)
	}

	dV, err := utils.ParseVersion(strings.Replace(desired, "v", "", 1), 3)
	if err != nil {
		return false, fmt.Errorf("failed to parse %s version %w", desired, err)
	}

	if cV < dV {
		return false, nil
	}

	return true, nil
}

func (c *OtelCol) ValidateCertManager() error {
	certMInstalled, err := c.IsCertManagerInstalledByTobs()
	if err != nil {
		return fmt.Errorf("ERROR: couldn't find the cert-manager status %v", err)
	}

	cmVersion, err := c.GetCertManagerVersion()
	if err != nil {
		return err
	}
	ok, err := isDesiredVersion(cmVersion, CertManagerVersion)
	if err != nil {
		return err
	}

	// cert-manager is not installed by tobs
	// ask user to upgrade their cert-manager isn't upto
	// the version we expect...
	if !certMInstalled {
		fmt.Println("Cert-manager isn't installed or managed by tobs.")
		if !ok {
			return fmt.Errorf("existing cert-manager in the cluster with version %s doesn't comply with the"+
				" OpenTelemetry Operator, requires cert-manager version %s", cmVersion, CertManagerVersion)
		}
	} else if !ok {
		// cert-manager is installed by tobs, upgrade cert-manager,
		// if others apps are using resources created by cert-manager
		// notify users that XYZ are the resources managed by existing
		// cert-manager, and error out stating these resources need a manual
		// upgrade before completing the upgrade

		certManagerResources, err := c.K8sClient.ListCertManagerDeprecatedCRs()
		if err != nil {
			return fmt.Errorf("failed to list certificate custom resources %v", err)
		}

		var onlyCertMResources []k8s.ResourceDetails
		for _, resource := range certManagerResources {
			// the listCMDeprecated resources gets them from otel-operator
			// namespace this needs to be ignored as otel-operator helm-chart upgrades them
			// Mote: Resource with name 'opentelemetry-operator-selfsigned-issuer` isn't part of
			// any namespace so adding a check.
			if resource.Namespace != "opentelemetry-operator-system" {
				if resource.Name != "opentelemetry-operator-selfsigned-issuer" {
					onlyCertMResources = append(onlyCertMResources, resource)
				}
			}
		}

		if len(onlyCertMResources) > 0 {
			fmt.Printf("\n!!! TOBS UPGRADE NEEDS MANUAL UPGRADE OF CERT-MANAGER RESOURCES: \n\n")

			fmt.Printf("As tracing is enabled in tobs, upgrading tobs requires upgrading cert-manager to %s, "+
				"The below listed cert-manager certificate resources are not managed by tobs and needs to be manually upgraded "+
				"using: the cmctl utility or by using the kubectl cert-manager plugin to convert old manifests to v1.\n"+
				"Follow this link for more details: https://cert-manager.io/docs/installation/upgrading/upgrading-1.5-1.6/ \n\n", CertManagerVersion)

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Name", "Namespace", "APIVersion", "Resource Type"})

			for _, v := range onlyCertMResources {
				table.Append([]string{v.Name, v.Namespace, v.APIVersion, v.ResourceType})
			}
			table.Render()

			fmt.Println()
			return fmt.Errorf("cert-manager resources not owned by tobs needs manual upgrade")
		}

		fmt.Printf("Upgrading the cert-manager to %s\n", CertManagerVersion)
		utils.ConfirmAction()
		c.UpgradeCM = true
	}

	return nil
}

func UpgradeCertManager() error {
	err := createUpgradeCertManager()
	if err != nil {
		return fmt.Errorf("failed to upgrade cert-manager %v", err)
	}

	fmt.Printf("Successfully upgraded cert-manager to %s\n", CertManagerVersion)
	return err
}

func createUpgradeCertManager() error {
	k8sClient := k8s.NewClient()
	err := k8sClient.ApplyManifests(certManagerManifests)
	if err != nil {
		return fmt.Errorf("failed to apply cert-manager %v", err)
	}

	// verify cert-manager is up & running
	pods, err := k8sClient.KubeGetPods(CertManagerNamespace, map[string]string{"app.kubernetes.io/instance": "cert-manager",
		"app.kubernetes.io/version": CertManagerVersion})
	if err != nil {
		return err
	}

	for _, pod := range pods {
		if err = k8sClient.KubeWaitOnPod(CertManagerNamespace, pod.Name); err != nil {
			return err
		}
	}

	if err = k8sClient.UpdateNamespaceLabels(CertManagerNamespace, map[string]string{"app.kubernetes.io/created-by": "tobs-cli"}); err != nil {
		return err
	}
	return nil
}

func DeleteOtelColCRD() error {
	apiClient := k8s.NewAPIClient()
	err := apiClient.DeleteCRD(otelColCRD)
	if err != nil {
		return fmt.Errorf("failed to delete %s CRD with error: %v", otelColCRD, err)
	}
	return nil
}

func (c *OtelCol) IsCertManagerInstalledByTobs() (bool, error) {
	labels, err := c.K8sClient.GetNamespaceLabels(CertManagerNamespace)
	if err != nil {
		return false, err
	}

	if val, ok := labels["app.kubernetes.io/created-by"]; ok {
		if val == "tobs-cli" {
			return true, nil
		}
	}
	return false, nil
}

func (c *OtelCol) GetCertManagerVersion() (string, error) {
	certManagerDeployment, err := c.K8sClient.GetDeployment("cert-manager", "cert-manager")
	if err != nil {
		return "", err
	}

	if val, ok := certManagerDeployment.Labels["app.kubernetes.io/version"]; ok {
		return val, nil
	}
	return "", nil
}

func (c *OtelCol) CreateDefaultCollector(otelColConfig string) error {
	// check the status of otel operator as CR creation needs webhooks
	// validation from operator
	otelOperatorPod, err := c.K8sClient.KubeGetPodName(otelOperatorNamespace, map[string]string{"control-plane": "controller-manager"})
	if err != nil {
		return fmt.Errorf("failed to find otel operator: %v", err)
	}
	err = c.K8sClient.KubeWaitOnPod(otelOperatorNamespace, otelOperatorPod)
	if err != nil {
		return err
	}

	// As collector config is string we should template .ReleaseName.Name and .Release.Namespace
	otelColConfig = strings.Replace(otelColConfig, "{{ .Release.Name }}", c.ReleaseName, -1)
	otelColConfig = strings.Replace(otelColConfig, "{{ .Release.Namespace }}", c.Namespace, -1)

	defaultOtelCol := &v1alpha1.OpenTelemetryCollector{
		ObjectMeta: metav1.ObjectMeta{
			Name: c.ReleaseName + "-opentelemetry",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       otelColKind,
			APIVersion: otelColApiVersion,
		},
		Spec: v1alpha1.OpenTelemetryCollectorSpec{
			Config: otelColConfig,
		},
	}
	body, err := json.Marshal(defaultOtelCol)
	if err != nil {
		return err
	}

	return c.K8sClient.CreateCustomResource(c.Namespace, otelColApiVersion, otelColResourceName, body)
}

func (c *OtelCol) DeleteDefaultOtelCollector() error {
	return c.K8sClient.DeleteCustomResource(c.Namespace, otelColApiVersion, otelColResourceName, c.ReleaseName+"-opentelemetry")
}

func (c *OtelCol) IsOtelOperatorEnabledInRelease() (bool, error) {
	var isEnabled bool
	e, err := c.HelmClient.ExportValuesFieldFromRelease(c.ReleaseName, []string{"opentelemetryOperator", "enabled"})
	if err != nil {
		return isEnabled, err
	}
	var ok bool
	isEnabled, ok = e.(bool)
	if !ok {
		return isEnabled, fmt.Errorf("opentelemetryOperator.enabled is not a bool")
	}

	return isEnabled, nil
}
