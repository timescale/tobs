package otel

import (
	"encoding/json"
	"fmt"

	"github.com/open-telemetry/opentelemetry-operator/api/v1alpha1"
	"github.com/timescale/tobs/cli/pkg/k8s"
	"github.com/timescale/tobs/cli/pkg/utils"
	errors2 "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	CertManagerManifests  = "https://github.com/jetstack/cert-manager/releases/download/v1.5.2/cert-manager.yaml"
	CertManagerNamespace  = "cert-manager"
	otelColKind           = "OpenTelemetryCollector"
	otelColApiVersion     = "opentelemetry.io/v1alpha1"
	otelOperatorNamespace = "opentelemetry-operator-system"
	otelColResourceName   = "opentelemetrycollectors"
)

func CreateCertManager(confirmActions bool) error {
	apiClient := k8s.NewAPIClient()
	_, err := apiClient.GetCRD("certificates.cert-manager.io")
	if err != nil {
		if errors2.IsNotFound(err) {
			fmt.Println("Couldn't find the cert-manager, Installing the cert-manager...")
			if !confirmActions {
				utils.ConfirmAction()
			}
			err = k8s.CreateK8sManifests([]string{CertManagerManifests})
			if err != nil {
				return err
			}
			// verify cert-manager is up & running
			k8sClient := k8s.NewClient()
			pods, err := k8sClient.KubeGetPods(CertManagerNamespace, map[string]string{"app.kubernetes.io/instance": "cert-manager"})
			if err != nil {
				return err
			}

			for _, pod := range pods {
				err = k8sClient.KubeWaitOnPod(CertManagerNamespace, pod.Name)
				if err != nil {
					return err
				}
			}
			fmt.Println("Successfully created cert-manager")
			return nil
		}
	}
	fmt.Println("cert-manager exists in the cluster, skipping the installation...")
	return err
}

func CreateDefaultCollector(release, namespace, otelColConfig string) error {
	// check the status of otel operator as CR creation needs webhooks
	// validation from operator
	k8sClient := k8s.NewClient()
	otelOperatorPod, err := k8sClient.KubeGetPodName(otelOperatorNamespace, map[string]string{"control-plane": "controller-manager"})
	if err != nil {
		return fmt.Errorf("failed to find otel operator: %v", err)
	}
	err = k8sClient.KubeWaitOnPod(otelOperatorNamespace, otelOperatorPod)
	if err != nil {
		return err
	}

	if otelColConfig == "" {
		otelColConfig = getDefaultOtelColConfig(release, namespace)
	}

	defaultOtelCol := &v1alpha1.OpenTelemetryCollector{
		ObjectMeta: metav1.ObjectMeta{
			Name: release + "-opentelemetry-collector",
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

	return k8sClient.CreateCustomResource(namespace, otelColApiVersion, otelColResourceName, body)
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
