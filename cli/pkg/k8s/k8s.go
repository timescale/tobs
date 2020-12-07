package k8s

import (
	"context"
	"errors"
	"fmt"
	"io"
	"k8s.io/apimachinery/pkg/api/resource"
	"log"
	"net/http"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	rest "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/transport/spdy"
)

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

func KubeGetPodName(namespace string, labelmap map[string]string) (string, error) {
	var err error

	client, _ := kubeInit()

	labelSelector := metav1.LabelSelector{MatchLabels: labelmap}
	listOptions := metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
	}

	pods, err := client.CoreV1().Pods(namespace).List(context.Background(), listOptions)
	if err != nil {
		return "", err
	}

	return pods.Items[0].Name, nil
}

func KubeGetServiceName(namespace string, labelmap map[string]string) (string, error) {
	var err error

	client, _ := kubeInit()

	labelSelector := metav1.LabelSelector{MatchLabels: labelmap}
	listOptions := metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
	}

	services, err := client.CoreV1().Services(namespace).List(context.Background(), listOptions)
	if err != nil {
		return "", err
	}

	if len(services.Items) < 1 {
		return "", fmt.Errorf("No such service found")
	}

	if len(services.Items) > 1 {
		return "", fmt.Errorf("Too many services found")
	}

	return services.Items[0].Name, nil
}

func KubeGetPVCNames(namespace string, labelmap map[string]string) ([]string, error) {
	var err error

	client, _ := kubeInit()

	labelSelector := metav1.LabelSelector{MatchLabels: labelmap}
	listOptions := metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
	}

	pvcs, err := client.CoreV1().PersistentVolumeClaims(namespace).List(context.Background(), listOptions)
	if err != nil {
		return nil, err
	}

	var names []string
	for _, pvc := range pvcs.Items {
		names = append(names, pvc.Name)
	}

	return names, nil
}

func KubeGetPods(namespace string, labelmap map[string]string) ([]corev1.Pod, error) {
	var err error

	client, _ := kubeInit()

	labelSelector := metav1.LabelSelector{MatchLabels: labelmap}
	listOptions := metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
	}

	pods, err := client.CoreV1().Pods(namespace).List(context.Background(), listOptions)
	if err != nil {
		return nil, err
	}

	return pods.Items, nil
}

func KubeGetSecret(namespace string, secretName string) (*corev1.Secret, error) {
	var err error

	client, _ := kubeInit()

	secret, err := client.CoreV1().Secrets(namespace).Get(context.Background(), secretName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return secret, nil
}

func KubeGetAllPods(namespace string, name string) ([]corev1.Pod, error) {
	var err error
	var allpods []corev1.Pod

	pods, err := KubeGetPods(namespace, map[string]string{"release": name})
	if err != nil {
		return nil, err
	}
	allpods = append(allpods, pods...)

	pods, err = KubeGetPods(namespace, map[string]string{"app.kubernetes.io/instance": name})
	if err != nil {
		return nil, err
	}
	allpods = append(allpods, pods...)

	pods, err = KubeGetPods(namespace, map[string]string{"app": name + "-promscale"})
	if err != nil {
		return nil, err
	}
	allpods = append(allpods, pods...)

	pods, err = KubeGetPods(namespace, map[string]string{"job-name": name + "-grafana-db"})
	if err != nil {
		return nil, err
	}
	allpods = append(allpods, pods...)

	return allpods, nil
}

// ExecCmd exec command on specific pod and wait the command's output.
func KubeExecCmd(namespace string, podName string, container string, command string, stdin io.Reader, tty bool) error {
	var err error

	client, config := kubeInit()

	shcmd := []string{
		"/bin/sh",
		"-c",
		command,
	}
	req := client.CoreV1().RESTClient().Post().Resource("pods").Namespace(namespace).
		Name(podName).SubResource("exec")
	option := &corev1.PodExecOptions{
		Container: container,
		Command:   shcmd,
		Stdin:     true,
		Stdout:    true,
		Stderr:    true,
		TTY:       tty,
	}
	if stdin == nil {
		option.Stdin = false
	}
	req.VersionedParams(
		option,
		scheme.ParameterCodec,
	)

	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return err
	}

	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  stdin,
		Stdout: os.Stdout,
		Stderr: os.Stdout,
	})
	if err != nil {
		return err
	}

	return nil
}

func KubePortForwardPod(namespace string, podName string, local int, remote int) (*portforward.PortForwarder, error) {
	var err error

	client, config := kubeInit()

	fmt.Printf("Listening to pod %v from port %d\n", podName, local)
	url := client.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(namespace).
		Name(podName).
		SubResource("portforward").URL()

	transport, upgrader, err := spdy.RoundTripperFor(config)
	if err != nil {
		return nil, err
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", url)

	ports := []string{fmt.Sprintf("%d:%d", local, remote)}

	pf, err := portforward.New(dialer, ports, make(chan struct{}, 1), make(chan struct{}, 1), os.Stdout, os.Stderr)
	if err != nil {
		return nil, err
	}

	errChan := make(chan error)
	go func() {
		errChan <- pf.ForwardPorts()
	}()

	select {
	case err = <-errChan:
		return nil, err
	case <-pf.Ready:
		return pf, nil
	}
}

func KubePortForwardService(namespace string, serviceName string, local int, remote int) (*portforward.PortForwarder, error) {
	var err error

	client, _ := kubeInit()

	service, err := client.CoreV1().Services(namespace).Get(context.Background(), serviceName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	set := labels.Set(service.Spec.Selector)
	listOptions := metav1.ListOptions{LabelSelector: set.AsSelector().String()}
	pods, err := client.CoreV1().Pods(namespace).List(context.Background(), listOptions)
	if err != nil {
		return nil, err
	}

	podName := pods.Items[0].Name

	pf, err := KubePortForwardPod(namespace, podName, local, remote)
	if err != nil {
		return nil, err
	}

	time.Sleep(1 * time.Second)
	return pf, nil
}

func KubeCreatePod(pod *corev1.Pod) error {
	var err error

	client, _ := kubeInit()

	fmt.Println("Creating pod...")
	_, err = client.CoreV1().Pods(pod.Namespace).Create(context.Background(), pod, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func KubeDeletePod(namespace string, podName string) error {
	var err error

	client, _ := kubeInit()

	fmt.Printf("Deleting pod %v...\n", podName)
	err = client.CoreV1().Pods(namespace).Delete(context.Background(), podName, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}

func KubeWaitOnPod(namespace string, podName string) error {
	client, _ := kubeInit()

	fmt.Printf("Waiting on pod %v...\n", podName)
	for i := 0; i < 6000; i++ {
		pod, err := client.CoreV1().Pods(namespace).Get(context.Background(), podName, metav1.GetOptions{})
		if err != nil {
			return err
		}

		podStatus := pod.Status.Phase
		if podStatus != corev1.PodPending && podStatus != corev1.PodFailed && podStatus != corev1.PodUnknown {
			fmt.Printf("Pod %v has started\n", podName)
			break
		} else if i == 5999 {
			fmt.Println("WARNING: pod did not come up in 10 minutes")
		}
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

func KubeDeleteService(namespace string, serviceName string) error {
	var err error

	client, _ := kubeInit()

	fmt.Printf("Deleting service %v...\n", serviceName)
	err = client.CoreV1().Services(namespace).Delete(context.Background(), serviceName, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}

func KubeDeleteEndpoint(namespace string, endpointName string) error {
	var err error

	client, _ := kubeInit()

	fmt.Printf("Deleting endpoint %v...\n", endpointName)
	err = client.CoreV1().Endpoints(namespace).Delete(context.Background(), endpointName, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}

func KubeDeletePVC(namespace string, PVCName string) error {
	var err error

	client, _ := kubeInit()

	fmt.Printf("Deleting PVC %v...\n", PVCName)
	err = client.CoreV1().PersistentVolumeClaims(namespace).Delete(context.Background(), PVCName, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}

func KubeUpdateSecret(namespace string, secret *corev1.Secret) error {
	var err error

	client, _ := kubeInit()

	fmt.Println("Updating secret...")
	_, err = client.CoreV1().Secrets(namespace).Update(context.Background(), secret, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func buildPVCNames(pvcPrefix string, pods []corev1.Pod) (pvcNames []string) {
	for _, pod := range pods {
		pvcNames = append(pvcNames, pvcPrefix+"-"+pod.Name)
	}
	return pvcNames
}

type PVCData struct {
	Name string
	SpecSize string
	StatusSize string
}

func GetPVCSizes(namespace, pvcPrefix string, labels map[string]string) ([]*PVCData, error) {
	var pvcs []string
	var pvcData []*PVCData
	if labels != nil {
		pods, err := KubeGetPods(namespace, labels)
		if err != nil {
			return nil, fmt.Errorf("failed to get the pods using labels %w", err)
		}
		if len(pods) == 0 {
			return nil, errors.New("failed to get the pod's for timescaledb")
		}
		pvcs = buildPVCNames(pvcPrefix, pods)
	} else {
		pvcs = append(pvcs, pvcPrefix)
	}

	client, _ := kubeInit()
	for _, pvcName := range pvcs {
		podPVC, err := client.CoreV1().PersistentVolumeClaims(namespace).Get(context.Background(), pvcName, metav1.GetOptions{})
		if err != nil {
			fmt.Println(fmt.Errorf("failed to get the pvc for %s %w", pvcName, err))
		}

		specSize := podPVC.Spec.Resources.Requests["storage"]
		statusSize := podPVC.Status.Capacity["storage"]
		pvc := &PVCData{
			Name:       pvcName,
			SpecSize:   specSize.String(),
			StatusSize: statusSize.String(),
		}
		if podPVC.Name != "" {
			pvcData = append(pvcData, pvc)
		}
	}

	return pvcData, nil
}

func ExpandTimescaleDBPVC(namespace, value, pvcPrefix string, labels map[string]string) (map[string]string, error) {
	pvcResults := make(map[string]string)
	pods, err := KubeGetPods(namespace, labels)
	if err != nil {
		return pvcResults, fmt.Errorf("failed to get the pods using labels %w", err)
	}

	pvcs := buildPVCNames(pvcPrefix, pods)
	for _, pvc := range pvcs {
		err := ExpandPVC(namespace, pvc, value)
		if err != nil {
			fmt.Println(fmt.Errorf("%w",err))
		} else {
			pvcResults[pvc] = value
		}
	}
	return pvcResults, nil
}

func ExpandPVC(namespace, pvcName, value string) error {
	client, _ := kubeInit()
	podPVC, err := client.CoreV1().PersistentVolumeClaims(namespace).Get(context.Background(), pvcName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get the pvc for %s %w", pvcName, err)
	}

	newSize, err := resource.ParseQuantity(value)
	if err != nil {
		return fmt.Errorf("failed to parse the volume size %w", err)
	}

	existingSize := podPVC.Spec.Resources.Requests["storage"]
	if yes := newSize.Cmp(existingSize); yes != 1 {
		return fmt.Errorf("provided volume size for pvc: %s is less than or equal to the existing size: %s", pvcName, existingSize.String())
	}

	podPVC.Spec.Resources.Requests["storage"] = newSize
	_, err = client.CoreV1().PersistentVolumeClaims(namespace).Update(context.Background(), podPVC, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update persistent volume claim %w", err)
	}

	return nil
}


/*
#########################################
Kubernetes utils for e2e tests.
#########################################
 */

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

func GetAllPVCSizes() (map[string]string, error){
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
