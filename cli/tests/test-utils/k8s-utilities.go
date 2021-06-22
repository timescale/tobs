package test_utils

import (
	"context"
	"fmt"
	"github.com/timescale/tobs/cli/pkg/k8s"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

/*
#########################################
Kubernetes utils for e2e tests.
#########################################
*/

type TestClient struct {
	K8s *k8s.Client
}

// By default local storage provider doesn't let us to expand PVC's
// For e2e tests to run we are configuring storageClass to allow PVC expansion
func (c *TestClient) UpdateStorageClassAllowVolumeExpand() error {
	storageClass, err := c.K8s.StorageV1().StorageClasses().Get(context.Background(), "standard", metav1.GetOptions{})
	if err != nil {
		return err
	}

	setTrue := true
	storageClass.AllowVolumeExpansion = &setTrue
	_, err = c.K8s.StorageV1().StorageClasses().Update(context.Background(), storageClass, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (c *TestClient) GetAllPVCSizes() (map[string]string, error) {
	pvcs, err := c.K8s.CoreV1().PersistentVolumeClaims("ns").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	results := make(map[string]string)
	for _, pvc := range pvcs.Items {
		size := pvc.Spec.Resources.Requests["storage"]
		results[pvc.Name] = size.String()
	}
	return results, nil
}

func (c *TestClient) GetUpdatedPromscaleMemResource(releaseName, namespace string) (string, error) {
	promscale, err := c.K8s.AppsV1().Deployments(namespace).Get(context.Background(), releaseName+"-promscale", metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	mem := promscale.Spec.Template.Spec.Containers[0].Resources.Requests["memory"]
	return mem.String(), nil
}

func (c *TestClient) CheckPodsRunning(namespace string) error {
	fmt.Println("Performing check on all are pods are running.")
	pods, err := c.K8s.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list pods to check status %v", err)
	}

	for _, pod := range pods.Items {
		for _, c := range pod.Status.ContainerStatuses {
			if !c.Ready {
				// If container is not ready some times they be in succeeded state which is fine
				if pod.Status.Phase != "Succeeded" && pod.Status.Phase != "Running" {
					return fmt.Errorf("failed to verify all the pods are running, "+
						"%s pod is not running, current status %s, container %s is not ready", pod.Name, pod.Status.Phase, c.Name)
				}
			}
		}
	}

	fmt.Println("All pods are in running state.")
	return nil
}

func (c *TestClient) CreateTimescaleDBNodePortService(namespace string) (string, error) {
	nodes, err := c.K8s.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to list nodes %v", err)
	}

	lb := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tsdb-lb-svc",
			Namespace: namespace,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Port:     5432,
					NodePort: 30007,
				},
			},
			Selector: map[string]string{"app": "external-db-tests-timescaledb", "release": "external-db-tests", "role": "master"},
			Type:     "NodePort",
		},
	}

	_, err = c.K8s.CoreV1().Services(namespace).Create(context.Background(), lb, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to create load balancer service for timescaledb %v", err)
	}

	ip := ""
	if len(nodes.Items) > 0 {
		for _, t := range nodes.Items[0].Status.Addresses {
			if t.Type == "InternalIP" {
				ip = t.Address
				break
			}
		}
	}

	return ip + ":30007", nil
}

func (c *TestClient) GetTSDBBackUpSecret(releaseName, namespace string) (*v1.Secret, error) {
	return c.K8s.CoreV1().Secrets(namespace).Get(context.Background(), releaseName+"-pgbackrest", metav1.GetOptions{})
}

func (c *TestClient) DeleteNamespace(namespace string) error {
	err := c.K8s.CoreV1().Namespaces().Delete(context.Background(), namespace, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (c *TestClient) DeletePod(pod, namespace string) error {
	gracePeriod := int64(0)
	return c.K8s.CoreV1().Pods(namespace).Delete(context.Background(), pod, metav1.DeleteOptions{GracePeriodSeconds: &gracePeriod})
}

func (c *TestClient) DeleteWebhooks() error {
	err := c.K8s.AdmissionregistrationV1().ValidatingWebhookConfigurations().Delete(context.Background(), "tobs-kube-prometheus-admission", metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete validatingwebhook %v", err)
	}

	err = c.K8s.AdmissionregistrationV1().MutatingWebhookConfigurations().Delete(context.Background(), "tobs-kube-prometheus-admission", metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete mutatingwebhook %v", err)
	}

	return nil
}