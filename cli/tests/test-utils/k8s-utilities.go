package test_utils

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"log"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

/*
#########################################
Kubernetes utils for e2e tests.
#########################################
*/

var HOME = os.Getenv("HOME")

func kubeInit() (kubernetes.Interface, *rest.Config) {
	var err error

	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: HOME + "/.kube/config"},
		&clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		log.Fatal(err)
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	return client, config
}

// By default local storage provider doesn't let us to expand PVC's
// For e2e tests to run we are configuring storageClass to allow PVC expansion
func UpdateStorageClassAllowVolumeExpand() error {
	client, _ := kubeInit()
	storageClass, err := client.StorageV1().StorageClasses().Get(context.Background(), "standard", metav1.GetOptions{})
	if err != nil {
		return err
	}

	setTrue := true
	storageClass.AllowVolumeExpansion = &setTrue
	_, err = client.StorageV1().StorageClasses().Update(context.Background(), storageClass, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func GetAllPVCSizes() (map[string]string, error) {
	client, _ := kubeInit()
	pvcs, err := client.CoreV1().PersistentVolumeClaims("ns").List(context.Background(), metav1.ListOptions{})
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

func GetUpdatedPromscaleMemResource() (string, error) {
	client, _ := kubeInit()
	promscale, err := client.AppsV1().Deployments("ns").Get(context.Background(), "gg-promscale", metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	mem := promscale.Spec.Template.Spec.Containers[0].Resources.Requests["memory"]
	return mem.String(), nil
}

func CheckPodsRunning(namespace string) error {
	fmt.Println("Performing check on all are pods are running.")
	client, _ := kubeInit()
	pods, err := client.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
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

func CreateTimescaleDBNodePortService(namespace string) (string, error) {
	client, _ := kubeInit()
	nodes, err := client.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
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

	_, err = client.CoreV1().Services(namespace).Create(context.Background(), lb, metav1.CreateOptions{})
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

func GetTSDBBackUpSecret(releaseName, namespace string) (*v1.Secret, error) {
	client, _ := kubeInit()
	secret, err := client.CoreV1().Secrets(namespace).Get(context.Background(), releaseName+"-pgbackrest", metav1.GetOptions{})
	if err != nil {
		return &v1.Secret{}, err
	}
	return secret, nil
}
