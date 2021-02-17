package test_utils

import (
	"context"
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
