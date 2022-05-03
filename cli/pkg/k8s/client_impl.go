package k8s

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	cmclient "github.com/cert-manager/cert-manager/pkg/client/clientset/versioned"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiext "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	errors2 "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/discovery/cached/disk"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/transport/spdy"
)

var HOME = os.Getenv("HOME")

type clientImpl struct {
	*kubernetes.Clientset
	Config *rest.Config
}

type apiClient struct {
	*apiext.Clientset
}

func NewClient() Client {
	var err error

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()

	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		&clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		log.Fatal(err)
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	return &clientImpl{client, config}
}

func NewAPIClient() apiClient {
	var err error
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		&clientcmd.ConfigOverrides{},
	).ClientConfig()
	if err != nil {
		log.Fatal(err)
	}

	client, err := apiext.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	return apiClient{client}
}

func newCMClient() *cmclient.Clientset {
	var err error
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		&clientcmd.ConfigOverrides{},
	).ClientConfig()
	if err != nil {
		log.Fatal(err)
	}

	cmClient, err := cmclient.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	return cmClient
}

func (c *clientImpl) KubeGetPodName(namespace string, labelmap map[string]string) (string, error) {
	var err error
	var podName string

	labelSelector := metav1.LabelSelector{MatchLabels: labelmap}
	listOptions := metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
	}

	pods, err := c.CoreV1().Pods(namespace).List(context.Background(), listOptions)
	if err != nil {
		return podName, err
	}

	if len(pods.Items) > 0 {
		podName = pods.Items[0].Name
	}

	return podName, nil
}

func (c *clientImpl) KubeGetServiceName(namespace string, labelmap map[string]string) (string, error) {
	var err error

	labelSelector := metav1.LabelSelector{MatchLabels: labelmap}
	listOptions := metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
	}

	services, err := c.CoreV1().Services(namespace).List(context.Background(), listOptions)
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

func (c *clientImpl) KubeGetPVCNames(namespace string, labelmap map[string]string) ([]string, error) {
	var err error

	labelSelector := metav1.LabelSelector{MatchLabels: labelmap}
	listOptions := metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
	}

	pvcs, err := c.CoreV1().PersistentVolumeClaims(namespace).List(context.Background(), listOptions)
	if err != nil {
		return nil, err
	}

	var names []string
	for _, pvc := range pvcs.Items {
		names = append(names, pvc.Name)
	}

	return names, nil
}

func (c *clientImpl) KubeGetPods(namespace string, labelmap map[string]string) ([]corev1.Pod, error) {
	var err error

	labelSelector := metav1.LabelSelector{MatchLabels: labelmap}
	listOptions := metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
	}

	pods, err := c.CoreV1().Pods(namespace).List(context.Background(), listOptions)
	if err != nil {
		return nil, err
	}

	return pods.Items, nil
}

func (c *clientImpl) KubeGetSecret(namespace string, secretName string) (*corev1.Secret, error) {
	var err error

	secret, err := c.CoreV1().Secrets(namespace).Get(context.Background(), secretName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return secret, nil
}

func (c *clientImpl) KubeGetAllSecrets(namespace string) (*corev1.SecretList, error) {
	var err error

	secrets, err := c.CoreV1().Secrets(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return secrets, nil
}

func (c *clientImpl) KubeGetAllPods(namespace string, releaseName string) ([]corev1.Pod, error) {
	var err error
	var allpods []corev1.Pod

	pods, err := c.KubeGetPods(namespace, map[string]string{"release": releaseName})
	if err != nil {
		return nil, err
	}
	allpods = append(allpods, pods...)

	pods, err = c.KubeGetPods(namespace, map[string]string{"app.kubernetes.io/instance": releaseName})
	if err != nil {
		return nil, err
	}
	allpods = append(allpods, pods...)

	pods, err = c.KubeGetPods(namespace, map[string]string{"app": releaseName + "-promscale"})
	if err != nil {
		return nil, err
	}
	allpods = append(allpods, pods...)

	pods, err = c.KubeGetPods(namespace, map[string]string{"job-name": releaseName + "-grafana-db"})
	if err != nil {
		return nil, err
	}
	allpods = append(allpods, pods...)

	return allpods, nil
}

// ExecCmd exec command on specific pod and wait the command's output.
func (c *clientImpl) KubeExecCmd(namespace string, podName string, container string, command string, stdin io.Reader, tty bool) error {
	var err error

	shcmd := []string{
		"/bin/sh",
		"-c",
		command,
	}
	req := c.CoreV1().RESTClient().Post().Resource("pods").Namespace(namespace).
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

	exec, err := remotecommand.NewSPDYExecutor(c.Config, "POST", req.URL())
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

func (c *clientImpl) KubePortForwardPod(namespace string, podName string, local int, remote int) (*portforward.PortForwarder, error) {
	var err error

	fmt.Printf("Listening to pod %v from port %d\n", podName, local)
	url := c.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(namespace).
		Name(podName).
		SubResource("portforward").URL()

	transport, upgrader, err := spdy.RoundTripperFor(c.Config)
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

func (c *clientImpl) KubePortForwardService(namespace string, serviceName string, local int, remote int) (*portforward.PortForwarder, error) {
	var err error

	service, err := c.CoreV1().Services(namespace).Get(context.Background(), serviceName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	set := labels.Set(service.Spec.Selector)
	listOptions := metav1.ListOptions{LabelSelector: set.AsSelector().String()}
	pods, err := c.CoreV1().Pods(namespace).List(context.Background(), listOptions)
	if err != nil {
		return nil, err
	}

	if len(pods.Items) == 0 {
		return nil, fmt.Errorf("couldn't find the pods for service: %s", serviceName)
	}
	podName := pods.Items[0].Name

	pf, err := c.KubePortForwardPod(namespace, podName, local, remote)
	if err != nil {
		return nil, err
	}

	time.Sleep(1 * time.Second)
	return pf, nil
}

func (c *clientImpl) KubeCreatePod(pod *corev1.Pod) error {
	var err error

	fmt.Println("Creating pod...")
	_, err = c.CoreV1().Pods(pod.Namespace).Create(context.Background(), pod, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (c *clientImpl) KubeDeletePod(namespace string, podName string) error {
	var err error

	fmt.Printf("Deleting pod %v...\n", podName)
	gracePeriodSecs := int64(0)
	err = c.CoreV1().Pods(namespace).Delete(context.Background(), podName, metav1.DeleteOptions{GracePeriodSeconds: &gracePeriodSecs})
	if err != nil {
		return err
	}

	return nil
}

func (c *clientImpl) KubeWaitOnPod(namespace string, podName string) error {
	fmt.Printf("Waiting on pod %v...\n", podName)
	for i := 0; i < 6000; i++ {
		pod, err := c.CoreV1().Pods(namespace).Get(context.Background(), podName, metav1.GetOptions{})
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

func (c *clientImpl) KubeDeleteService(namespace string, serviceName string) error {
	var err error
	fmt.Printf("Deleting service %v...\n", serviceName)
	err = c.CoreV1().Services(namespace).Delete(context.Background(), serviceName, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (c *clientImpl) KubeDeleteEndpoint(namespace string, endpointName string) error {
	var err error
	fmt.Printf("Deleting endpoint %v...\n", endpointName)
	err = c.CoreV1().Endpoints(namespace).Delete(context.Background(), endpointName, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (c *clientImpl) KubeDeletePVC(namespace string, PVCName string) error {
	var err error

	fmt.Printf("Deleting PVC %v...\n", PVCName)
	err = c.CoreV1().PersistentVolumeClaims(namespace).Delete(context.Background(), PVCName, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (c *clientImpl) KubeUpdateSecret(namespace string, secret *corev1.Secret) error {
	var err error
	fmt.Println("Updating secret...")
	_, err = c.CoreV1().Secrets(namespace).Update(context.Background(), secret, metav1.UpdateOptions{})
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
	Name       string
	SpecSize   string
	StatusSize string
}

func (c *clientImpl) GetPVCSizes(namespace, pvcPrefix string, labels map[string]string) ([]*PVCData, error) {
	var pvcs []string
	var pvcData []*PVCData
	if labels != nil {
		pods, err := c.KubeGetPods(namespace, labels)
		if err != nil {
			return nil, fmt.Errorf("failed to get the pods using labels %w", err)
		}
		if len(pods) == 0 {
			return nil, fmt.Errorf("failed to get the pod's for pvc %s with labelSet: %v", pvcPrefix, labels)
		}
		pvcs = buildPVCNames(pvcPrefix, pods)
	} else {
		pvcs = append(pvcs, pvcPrefix)
	}

	for _, pvcName := range pvcs {
		podPVC, err := c.CoreV1().PersistentVolumeClaims(namespace).Get(context.Background(), pvcName, metav1.GetOptions{})
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

func (c *clientImpl) ExpandPVCsForAllPods(namespace, value, pvcPrefix string, labels map[string]string) (map[string]string, error) {
	pvcResults := make(map[string]string)
	pods, err := c.KubeGetPods(namespace, labels)
	if err != nil {
		return pvcResults, fmt.Errorf("failed to get the pods using labels %w", err)
	}

	pvcs := buildPVCNames(pvcPrefix, pods)
	for _, pvc := range pvcs {
		err := c.ExpandPVC(namespace, pvc, value)
		if err != nil {
			fmt.Println(fmt.Errorf("%w", err))
		} else {
			pvcResults[pvc] = value
		}
	}
	return pvcResults, nil
}

func (c *clientImpl) ExpandPVC(namespace, pvcName, value string) error {
	podPVC, err := c.CoreV1().PersistentVolumeClaims(namespace).Get(context.Background(), pvcName, metav1.GetOptions{})
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
	_, err = c.CoreV1().PersistentVolumeClaims(namespace).Update(context.Background(), podPVC, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update persistent volume claim %w", err)
	}

	return nil
}

func (c *clientImpl) DeletePods(namespace string, labels map[string]string, forceKill bool) error {
	pods, err := c.KubeGetPods(namespace, labels)
	if err != nil {
		return fmt.Errorf("failed to get the pods using labels %w", err)
	}

	var deleteOptions metav1.DeleteOptions
	if forceKill {
		gracePeriodSecs := int64(0)
		deleteOptions = metav1.DeleteOptions{GracePeriodSeconds: &gracePeriodSecs}
	}

	for _, pod := range pods {
		err = c.CoreV1().Pods(namespace).Delete(context.Background(), pod.Name, deleteOptions)
		if err != nil {
			return fmt.Errorf("failed to delete the pod: %s %v\n", pod.Name, err)
		}
	}
	return nil
}

func (c *clientImpl) CreateSecret(secret *corev1.Secret) error {
	_, err := c.CoreV1().Secrets(secret.Namespace).Create(context.Background(), secret, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create %s secret %v", secret.Name, err)
	}
	return nil
}

func (c *clientImpl) DeleteSecret(secretName, namespace string) error {
	err := c.CoreV1().Secrets(namespace).Delete(context.Background(), secretName, metav1.DeleteOptions{})
	if err != nil {
		ok := errors2.IsNotFound(err)
		if !ok {
			return fmt.Errorf("failed to delete %s secret %v", secretName, err)
		}
	}
	return nil
}

func (c *clientImpl) CreateNamespaceIfNotExists(namespace string) error {
	namespaces, err := c.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list the namespaces to verify namespace existence %v", err)
	}

	for _, n := range namespaces.Items {
		if n.Name == namespace {
			return nil
		}
	}

	n := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err = c.CoreV1().Namespaces().Create(context.Background(), n, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create the namespace %v", err)
	}

	return nil
}

func (c *clientImpl) UpdateNamespaceLabels(name string, labels map[string]string) error {
	namespace, err := c.CoreV1().Namespaces().Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get the namespace %v", err)
	}

	for k, v := range labels {
		namespace.Labels[k] = v
	}

	_, err = c.CoreV1().Namespaces().Update(context.TODO(), namespace, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update the namespace annotations %v", err)
	}

	return nil
}

func (c *clientImpl) GetNamespaceLabels(name string) (map[string]string, error) {
	namespace, err := c.CoreV1().Namespaces().Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get the namespace %v", err)
	}

	return namespace.Labels, nil
}

func (c *clientImpl) CheckSecretExists(secretName, namespace string) (bool, error) {
	secExists, err := c.CoreV1().Secrets(namespace).Get(context.Background(), secretName, metav1.GetOptions{})
	if err != nil {
		ok := errors2.IsNotFound(err)
		if !ok {
			return false, fmt.Errorf("failed to get %s secret present or not %v", secretName, err)
		}
	}

	if secExists.Name != "" {
		return true, nil
	}

	return false, nil
}

func (c *clientImpl) DeleteJob(name, namespace string) error {
	// deleting the job leaves the completed pods
	// this might cause pvc to be in terminating state forever
	// so delete the job with propagation as background
	propagation := metav1.DeletePropagationBackground
	err := c.BatchV1().Jobs(namespace).Delete(context.Background(), name, metav1.DeleteOptions{PropagationPolicy: &propagation})
	if err != nil {
		return err
	}
	return nil
}

func (c *clientImpl) DeleteDaemonset(name, namespace string) error {
	fmt.Printf("Deleting daemonset %v...\n", name)
	err := c.AppsV1().DaemonSets(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (c *clientImpl) CreateJob(job *batchv1.Job) error {
	_, err := c.BatchV1().Jobs(job.Namespace).Create(context.Background(), job, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return err
}

func (c *clientImpl) GetJob(jobName, namespace string) (*batchv1.Job, error) {
	return c.BatchV1().Jobs(namespace).Get(context.Background(), jobName, metav1.GetOptions{})
}

func (c *clientImpl) GetDeployment(name, namespace string) (*appsv1.Deployment, error) {
	return c.AppsV1().Deployments(namespace).Get(context.Background(), name, metav1.GetOptions{})
}

func (c *clientImpl) UpdateDeployment(deployment *appsv1.Deployment) error {
	_, err := c.AppsV1().Deployments(deployment.Namespace).Update(context.Background(), deployment, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return err
}

func (c *clientImpl) DeleteDeployment(labelmap map[string]string, namespace string) error {
	labelSelector := metav1.LabelSelector{MatchLabels: labelmap}
	listOptions := metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
	}

	dList, err := c.AppsV1().Deployments(namespace).List(context.Background(), listOptions)
	if err != nil {
		return err
	}

	for _, ind := range dList.Items {
		dPolicy := metav1.DeletionPropagation("Foreground")
		err = c.AppsV1().Deployments(namespace).Delete(context.Background(), ind.Name, metav1.DeleteOptions{PropagationPolicy: &dPolicy})
		if err != nil {
			return err
		}
	}
	return nil
}

// This func helps to map the existing PV from a older PVC to new PVC
func (c *clientImpl) UpdatePVToNewPVC(pvcName, newPVCName, namespace string, pvcLabels map[string]string) error {
	pvc, err := c.CoreV1().PersistentVolumeClaims(namespace).Get(context.Background(), pvcName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get persistent volume claim %s: %v", pvcName, err)
	}

	pvName := pvc.Spec.VolumeName
	pv, err := c.CoreV1().PersistentVolumes().Get(context.Background(), pvName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get the persistent volume %s: %v", pvName, err)
	}

	pv.Spec.ClaimRef = nil
	pv, err = c.CoreV1().PersistentVolumes().Update(context.Background(), pv, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update persistent volume %s: %v", pv.Name, err)
	}

	newPVC := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      newPVCName,
			Namespace: namespace,
			Labels:    pvcLabels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: pvc.Spec.AccessModes,
			Resources: corev1.ResourceRequirements{
				Requests: pvc.Spec.Resources.Requests,
			},
			VolumeName: pv.Name,
			VolumeMode: pvc.Spec.VolumeMode,
		},
	}

	_, err = c.CoreV1().PersistentVolumeClaims(namespace).Create(context.Background(), newPVC, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to get persistent volume claim %s: %v", pvcName, err)
	}

	return nil
}

func (c *apiClient) GetCRD(name string) (*v1.CustomResourceDefinition, error) {
	return c.ApiextensionsV1().CustomResourceDefinitions().Get(context.Background(), name, metav1.GetOptions{})
}

func (c *apiClient) DeleteCRD(name string) error {
	return c.ApiextensionsV1().CustomResourceDefinitions().Delete(context.TODO(), name, metav1.DeleteOptions{})
}

func (c *clientImpl) CreateCustomResource(namespace, apiVersion, resourceName string, body []byte) error {
	_, err := c.RESTClient().Post().
		AbsPath("/apis/" + apiVersion).Namespace(namespace).Resource(resourceName).
		Body(body).
		DoRaw(context.TODO())
	return err
}

func (c *clientImpl) DeleteCustomResource(namespace, apiVersion, resourceName, crName string) error {
	_, err := c.RESTClient().Delete().
		AbsPath("/apis/" + apiVersion).Namespace(namespace).Resource(resourceName).Name(crName).
		DoRaw(context.TODO())
	return err
}

type ResourceDetails struct {
	Name         string
	Namespace    string
	APIVersion   string
	ResourceType string
}

func (c *clientImpl) ListCertManagerDeprecatedCRs() ([]ResourceDetails, error) {
	var resources []ResourceDetails
	client := newCMClient()

	namespaces, err := c.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return resources, err
	}

	for _, namespace := range namespaces.Items {
		// v1beta1 Certificates
		certs1, err := client.CertmanagerV1beta1().Certificates(namespace.Name).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return resources, fmt.Errorf("failed to list custom reosurce Certificates wih apiVersion: v1beta1 %v", err)
		}
		for _, c := range certs1.Items {
			resources = append(resources, ResourceDetails{
				Name:         c.Name,
				Namespace:    c.Namespace,
				APIVersion:   c.APIVersion,
				ResourceType: c.Kind,
			})
		}

		// v1beta1 CertificateRequests
		certsReq1, err := client.CertmanagerV1beta1().CertificateRequests(namespace.Name).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return resources, fmt.Errorf("failed to list custom reosurce CertificateRequests wih apiVersion: v1beta1 %v", err)
		}
		for _, c := range certsReq1.Items {
			resources = append(resources, ResourceDetails{
				Name:         c.Name,
				Namespace:    c.Namespace,
				APIVersion:   c.APIVersion,
				ResourceType: c.Kind,
			})
		}

		// v1beta1 ClusterIssuers
		clusterIssuers1, err := client.CertmanagerV1beta1().ClusterIssuers().List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return resources, fmt.Errorf("failed to list custom reosurce ClusterIssuers wih apiVersion: v1beta1 %v", err)
		}
		for _, c := range clusterIssuers1.Items {
			resources = append(resources, ResourceDetails{
				Name:         c.Name,
				APIVersion:   c.APIVersion,
				ResourceType: c.Kind,
			})
		}

		// v1beta1 Issuers
		issuers1, err := client.CertmanagerV1beta1().Issuers(namespace.Name).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return resources, fmt.Errorf("failed to list custom reosurce Issuers wih apiVersion: v1beta1 %v", err)
		}
		for _, c := range issuers1.Items {
			resources = append(resources, ResourceDetails{
				Name:         c.Name,
				APIVersion:   c.APIVersion,
				ResourceType: c.Kind,
			})
		}

		// v1alpha2 Certificates
		certs2, err := client.CertmanagerV1alpha2().Certificates(namespace.Name).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return resources, fmt.Errorf("failed to list custom reosurce Certificates with apiVersion: v1alpha2 %v", err)
		}
		for _, c := range certs2.Items {
			resources = append(resources, ResourceDetails{
				Name:         c.Name,
				Namespace:    c.Namespace,
				APIVersion:   c.APIVersion,
				ResourceType: c.Kind,
			})
		}

		// v1alpha2 CertificateRequests
		certsReq2, err := client.CertmanagerV1alpha2().CertificateRequests(namespace.Name).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return resources, fmt.Errorf("failed to list custom reosurce CertificateRequests with apiVersion: v1alpha2 %v", err)
		}
		for _, c := range certsReq2.Items {
			resources = append(resources, ResourceDetails{
				Name:         c.Name,
				Namespace:    c.Namespace,
				APIVersion:   c.APIVersion,
				ResourceType: c.Kind,
			})
		}

		// v1alpha2 ClusterIssuers
		clusterIssuer2, err := client.CertmanagerV1alpha2().ClusterIssuers().List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return resources, fmt.Errorf("failed to list custom reosurce ClusterIssuers with apiVersion: v1alpha2 %v", err)
		}
		for _, c := range clusterIssuer2.Items {
			resources = append(resources, ResourceDetails{
				Name:         c.Name,
				Namespace:    c.Namespace,
				APIVersion:   c.APIVersion,
				ResourceType: c.Kind,
			})
		}

		// v1alpha2 Issuers
		issuers2, err := client.CertmanagerV1alpha2().Issuers(namespace.Name).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return resources, fmt.Errorf("failed to list custom reosurce Issuers with apiVersion: v1alpha2 %v", err)
		}
		for _, c := range issuers2.Items {
			resources = append(resources, ResourceDetails{
				Name:         c.Name,
				Namespace:    c.Namespace,
				APIVersion:   c.APIVersion,
				ResourceType: c.Kind,
			})
		}

		// v1alpha3 Certificates
		certs3, err := client.CertmanagerV1alpha3().Certificates(namespace.Name).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return resources, fmt.Errorf("failed to list custom reosurce Certificates with apiVersion: v1alpha3 %v", err)
		}
		for _, c := range certs3.Items {
			resources = append(resources, ResourceDetails{
				Name:         c.Name,
				Namespace:    c.Namespace,
				APIVersion:   c.APIVersion,
				ResourceType: c.Kind,
			})
		}

		// v1alpha3 CertificateRequests
		certReq3, err := client.CertmanagerV1alpha3().CertificateRequests(namespace.Name).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return resources, fmt.Errorf("failed to list custom reosurce CertificateRequests with apiVersion: v1alpha3 %v", err)
		}
		for _, c := range certReq3.Items {
			resources = append(resources, ResourceDetails{
				Name:         c.Name,
				Namespace:    c.Namespace,
				APIVersion:   c.APIVersion,
				ResourceType: c.Kind,
			})
		}

		// v1alpha3 ClusterIssuers
		clusterIssuers3, err := client.CertmanagerV1alpha3().ClusterIssuers().List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return resources, fmt.Errorf("failed to list custom reosurce ClusterIssuers with apiVersion: v1alpha3 %v", err)
		}
		for _, c := range clusterIssuers3.Items {
			resources = append(resources, ResourceDetails{
				Name:         c.Name,
				Namespace:    c.Namespace,
				APIVersion:   c.APIVersion,
				ResourceType: c.Kind,
			})
		}

		// v1alpha3 Issuers
		issuers3, err := client.CertmanagerV1alpha3().Issuers(namespace.Name).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return resources, fmt.Errorf("failed to list custom reosurce Issuers with apiVersion: v1alpha3 %v", err)
		}
		for _, c := range issuers3.Items {
			resources = append(resources, ResourceDetails{
				Name:         c.Name,
				Namespace:    c.Namespace,
				APIVersion:   c.APIVersion,
				ResourceType: c.Kind,
			})
		}
	}

	return resources, nil
}

// Apply manifests helps to apply the k8s resources
// to the cluster this is equivalent to
// kubectl apply -f
func (c *clientImpl) ApplyManifests(manifests map[string]string) error {
	for name, manifestURL := range manifests {
		res, err := http.Get(manifestURL)
		if err != nil {
			return fmt.Errorf("failed to download %s: %v", name, err)
		}
		// Check server response
		if res.StatusCode != http.StatusOK {
			return fmt.Errorf("bad status: %s", res.Status)
		}

		// scan the response
		out, err := ioutil.ReadAll(res.Body)
		_ = res.Body.Close()
		if err != nil {
			return err
		}

		if err = c.applyYaml(out); err != nil {
			return err
		}
	}

	return nil
}

func (c *clientImpl) applyYaml(data []byte) error {
	chanMes, chanErr := readYaml(data)
	for {
		select {
		case dataBytes, ok := <-chanMes:
			{
				if !ok {
					return nil
				}

				// Get obj and dr
				obj, dr, err := c.buildDynamicResourceClient(dataBytes)
				if err != nil {
					continue
				}

				// Create or Update
				forceApply := true
				_, err = dr.Patch(context.TODO(), obj.GetName(), types.ApplyPatchType, dataBytes, metav1.PatchOptions{
					FieldManager: "tobs", Force: &forceApply,
				})
				if err != nil {
					return fmt.Errorf("failed to apply manifest with error %v", err)
				}
			}
		case err, ok := <-chanErr:
			if !ok {
				return nil
			}
			if err != nil {
				return fmt.Errorf("failed to read yaml %v", err)
			}
		}
	}
}

func readYaml(data []byte) (<-chan []byte, <-chan error) {
	var (
		chanErr        = make(chan error)
		chanBytes      = make(chan []byte)
		multidocReader = utilyaml.NewYAMLReader(bufio.NewReader(bytes.NewReader(data)))
	)

	go func() {
		defer close(chanErr)
		defer close(chanBytes)

		for {
			buf, err := multidocReader.Read()
			if err != nil {
				if err == io.EOF {
					return
				}
				chanErr <- fmt.Errorf("failed to read yaml data %v", err)
				return
			}
			chanBytes <- buf
		}
	}()
	return chanBytes, chanErr
}

func (c *clientImpl) buildDynamicResourceClient(data []byte) (obj *unstructured.Unstructured, dr dynamic.ResourceInterface, err error) {
	// Decode YAML manifest into unstructured.Unstructured
	obj = &unstructured.Unstructured{}
	decUnstructured := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	_, gvk, err := decUnstructured.Decode(data, nil, obj)
	if err != nil {
		return obj, dr, fmt.Errorf("decode yaml failed %v", err)
	}

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()

	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		&clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		log.Fatal(err)
	}

	// Some code to define this take from
	// https://github.com/kubernetes/cli-runtime/blob/master/pkg/genericclioptions/config_flags.go#L215
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	cacheDir := homeDir + "/.cache/tobs"
	httpCacheDir := filepath.Join(cacheDir, "http")
	discoveryCacheDir := computeDiscoverCacheDir(filepath.Join(cacheDir, "discovery"), config.Host)

	// DiscoveryClient queries API server about the resources
	cdc, err := disk.NewCachedDiscoveryClientForConfig(config, discoveryCacheDir, httpCacheDir, 10*time.Minute)
	if err != nil {
		log.Fatal(err)
	}

	mapper := restmapper.NewDeferredDiscoveryRESTMapper(cdc)

	// Find GVR
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return obj, dr, fmt.Errorf("mapping kind with version failed %v", err)
	}

	// Prepare dynamic client
	dynamicClient, err := dynamic.NewForConfig(c.Config)
	if err != nil {
		return obj, dr, fmt.Errorf("failed to create dynamic client %v", err)
	}

	// Obtain REST interface for the GVR
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		// namespaced resources should specify the namespace
		dr = dynamicClient.Resource(mapping.Resource).Namespace(obj.GetNamespace())
	} else {
		// for cluster-wide resources
		dr = dynamicClient.Resource(mapping.Resource)
	}
	return obj, dr, nil
}

// computeDiscoverCacheDir takes the parentDir and the host and comes up with a "usually non-colliding" name.
func computeDiscoverCacheDir(parentDir, host string) string {
	// strip the optional scheme from host if its there:
	schemelessHost := strings.Replace(strings.Replace(host, "https://", "", 1), "http://", "", 1)
	// now do a simple collapse of non-AZ09 characters.  Collisions are possible but unlikely.
	// Even if we do collide the problem is short lived
	var overlyCautiousIllegalFileCharacters = regexp.MustCompile(`[^(\w/\.)]`)
	safeHost := overlyCautiousIllegalFileCharacters.ReplaceAllString(schemelessHost, "_")
	return filepath.Join(parentDir, safeHost)
}
