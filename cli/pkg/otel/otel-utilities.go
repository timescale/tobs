package otel

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/open-telemetry/opentelemetry-operator/api/v1alpha1"
	"github.com/timescale/tobs/cli/pkg/helm"
	"github.com/timescale/tobs/cli/pkg/k8s"
	"github.com/timescale/tobs/cli/pkg/utils"
	errors2 "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	CertManagerVersion    = "v1.5.2"
	CertManagerNamespace  = "cert-manager"
	otelColKind           = "OpenTelemetryCollector"
	otelColApiVersion     = "opentelemetry.io/v1alpha1"
	otelOperatorNamespace = "opentelemetry-operator-system"
	otelColResourceName   = "opentelemetrycollectors"
)

var (
	CertManagerManifests = fmt.Sprintf("https://github.com/jetstack/cert-manager/releases/download/%s/cert-manager.yaml", CertManagerVersion)
	otelColCRD           = fmt.Sprintf("%s.opentelemetry.io", otelColResourceName)
)

type OtelCol struct {
	ReleaseName string
	Namespace   string
	K8sClient   k8s.Client
	HelmClient  helm.Client
}

func CreateCertManager(confirmActions bool) error {
	apiClient := k8s.NewAPIClient()
	_, err := apiClient.GetCRD("certificates.cert-manager.io")
	if err != nil {
		if errors2.IsNotFound(err) {
			fmt.Println("Couldn't find the cert-manager, Installing the cert-manager...")
			if !confirmActions {
				utils.ConfirmAction()
			}

			res, err := http.Get(CertManagerManifests)
			if err != nil {
				return fmt.Errorf("failed to download %v", err)
			}
			defer res.Body.Close()

			// Check server response
			if res.StatusCode != http.StatusOK {
				return fmt.Errorf("bad status: %s", res.Status)
			}

			// Writer the body to file
			var out []byte
			out, err = ioutil.ReadAll(res.Body)
			if err != nil  {
				return err
			}

			k8sClient := k8s.NewClient()
			err = k8sClient.ApplyManifests(out)
			if err != nil {
				return fmt.Errorf("failed to apply cert-manager %v", err)
			}

			// verify cert-manager is up & running
			pods, err := k8sClient.KubeGetPods(CertManagerNamespace, map[string]string{"app.kubernetes.io/instance": "cert-manager"})
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
			fmt.Println("Successfully created cert-manager")
			return nil
		}
	}
	fmt.Println("cert-manager exists in the cluster, skipping the installation...")
	return err
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

	if otelColConfig == "" {
		otelColConfig = getDefaultOtelColConfig(c.ReleaseName, c.Namespace)
	}

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

func getDefaultOtelColConfig(release, namespace string) string {
	return fmt.Sprintf(`
    receivers:
      jaeger:
        protocols:
          grpc:
          thrift_http:

      otlp:
        protocols:
          grpc:
          http:

    exporters:
      logging:
      otlp:
        endpoint: "%s-promscale-connector.%s.svc.cluster.local:9202"
        insecure: true

    processors:
      batch:

    service:
      pipelines:
        traces:
          receivers: [jaeger, otlp]
          exporters: [logging, otlp]
          processors: [batch]
`, release, namespace)
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
