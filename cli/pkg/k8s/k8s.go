package k8s

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	errors2 "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/transport/spdy"
)

var HOME = os.Getenv("HOME")

type Client struct {
	kubernetes.Clientset
}

func NewClient() (*Client, *rest.Config) {
	var err error

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()

	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		&clientcmd.ConfigOverrides{Timeout: "10m"}).ClientConfig()
	if err != nil {
		log.Fatal(err)
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	return &Client{*client}, config
}

func (c *Client) KubeGetPodName(namespace string, labelmap map[string]string) (string, error) {
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

func (c *Client) KubeGetServiceName(namespace string, labelmap map[string]string) (string, error) {
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

func (c *Client) KubeGetPVCNames(namespace string, labelmap map[string]string) ([]string, error) {
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

func (c *Client) KubeGetPods(namespace string, labelmap map[string]string) ([]corev1.Pod, error) {
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

func (c *Client) KubeGetSecret(namespace string, secretName string) (*corev1.Secret, error) {
	var err error
	secret, err := c.CoreV1().Secrets(namespace).Get(context.Background(), secretName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return secret, nil
}

func (c *Client) KubeGetAllSecrets(namespace string) (*corev1.SecretList, error) {
	var err error
	secrets, err := c.CoreV1().Secrets(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return secrets, nil
}

func (c *Client) KubeGetAllPods(namespace string, name string) ([]corev1.Pod, error) {
	var err error
	var allpods []corev1.Pod

	pods, err := c.KubeGetPods(namespace, map[string]string{"release": name})
	if err != nil {
		return nil, err
	}
	allpods = append(allpods, pods...)

	pods, err = c.KubeGetPods(namespace, map[string]string{"app.kubernetes.io/instance": name})
	if err != nil {
		return nil, err
	}
	allpods = append(allpods, pods...)

	pods, err = c.KubeGetPods(namespace, map[string]string{"app": name + "-promscale"})
	if err != nil {
		return nil, err
	}
	allpods = append(allpods, pods...)

	pods, err = c.KubeGetPods(namespace, map[string]string{"job-name": name + "-grafana-db"})
	if err != nil {
		return nil, err
	}
	allpods = append(allpods, pods...)

	return allpods, nil
}

// ExecCmd exec command on specific pod and wait the command's output.
func (c *Client) KubeExecCmd(namespace string, podName string, container string, command string, stdin io.Reader, tty bool) error {
	var err error
	_, config := NewClient()
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

func (c *Client) KubePortForwardPod(namespace string, podName string, local int, remote int) (*portforward.PortForwarder, error) {
	var err error
	fmt.Printf("Listening to pod %v from port %d\n", podName, local)
	url := c.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(namespace).
		Name(podName).
		SubResource("portforward").URL()

	_, config := NewClient()
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

func (c *Client) KubePortForwardService(namespace string, serviceName string, local int, remote int) (*portforward.PortForwarder, error) {
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

	podName := pods.Items[0].Name

	pf, err := c.KubePortForwardPod(namespace, podName, local, remote)
	if err != nil {
		return nil, err
	}

	time.Sleep(1 * time.Second)
	return pf, nil
}

func (c *Client) KubeCreatePod(pod *corev1.Pod) error {
	var err error
	fmt.Println("Creating pod...")
	_, err = c.CoreV1().Pods(pod.Namespace).Create(context.Background(), pod, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) KubeDeletePod(namespace string, podName string) error {
	var err error
	fmt.Printf("Deleting pod %v...\n", podName)
	gracePeriodSecs := int64(0)
	err = c.CoreV1().Pods(namespace).Delete(context.Background(), podName, metav1.DeleteOptions{GracePeriodSeconds: &gracePeriodSecs})
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) KubeWaitOnPod(namespace string, podName string) error {
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

func (c *Client) KubeDeleteService(namespace string, serviceName string) error {
	var err error
	fmt.Printf("Deleting service %v...\n", serviceName)
	err = c.CoreV1().Services(namespace).Delete(context.Background(), serviceName, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) KubeDeleteEndpoint(namespace string, endpointName string) error {
	var err error
	fmt.Printf("Deleting endpoint %v...\n", endpointName)
	err = c.CoreV1().Endpoints(namespace).Delete(context.Background(), endpointName, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) KubeDeletePVC(namespace string, PVCName string) error {
	var err error
	fmt.Printf("Deleting PVC %v...\n", PVCName)
	err = c.CoreV1().PersistentVolumeClaims(namespace).Delete(context.Background(), PVCName, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) KubeUpdateSecret(namespace string, secret *corev1.Secret) error {
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

func (c *Client) GetPVCSizes(namespace, pvcPrefix string, labels map[string]string) ([]*PVCData, error) {
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

func (c *Client) ExpandPVCsForAllPods(namespace, value, pvcPrefix string, labels map[string]string) (map[string]string, error) {
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

func (c *Client) ExpandPVC(namespace, pvcName, value string) error {
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

func (c *Client) DeletePods(namespace string, labels map[string]string, forceKill bool) error {
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

func (c *Client) CreateSecret(secret *corev1.Secret) error {
	_, err := c.CoreV1().Secrets(secret.Namespace).Create(context.Background(), secret, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create %s secret %v", secret.Name, err)
	}
	return nil
}

func (c *Client) DeleteSecret(secretName, namespace string) error {
	err := c.CoreV1().Secrets(namespace).Delete(context.Background(), secretName, metav1.DeleteOptions{})
	if err != nil {
		ok := errors2.IsNotFound(err)
		if !ok {
			return fmt.Errorf("failed to delete %s secret %v", secretName, err)
		}
	}
	return nil
}

func (c *Client) CreateNamespaceIfNotExists(namespace string) error {
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

func (c *Client) CheckSecretExists(secretName, namespace string) (bool, error) {
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

func (c *Client) DeleteJob(name, namespace string) error {
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

func (c *Client) DeleteDaemonset(name, namespace string) error {
	fmt.Printf("Deleting daemonset %v...\n", name)
	err := c.AppsV1().DaemonSets(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) CreateJob(job *batchv1.Job) error {
	_, err := c.BatchV1().Jobs(job.Namespace).Create(context.Background(), job, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return err
}

func (c *Client) GetJob(jobName, namespace string) (*batchv1.Job, error) {
	return c.BatchV1().Jobs(namespace).Get(context.Background(), jobName, metav1.GetOptions{})
}

func CreateCRDS(crds []string) error {
	for _, crd := range crds {
		out := exec.Command("kubectl", "apply", "-f", crd)
		output, err := out.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to create CRD %s %s: %w", crd, output, err)
		}
	}
	return nil
}

func (c *Client) GetDeployment(name, namespace string) (*appsv1.Deployment, error) {
	return c.AppsV1().Deployments(namespace).Get(context.Background(), name, metav1.GetOptions{})
}

func (c *Client) UpdateDeployment(deployment *appsv1.Deployment) error {
	_, err := c.AppsV1().Deployments(deployment.Namespace).Update(context.Background(), deployment, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return err
}

func (c *Client) UpdatePrometheusPV(pvcName, newPVCName, namespace, releaseName string) error {
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
			Labels:    map[string]string{"app": "prometheus", "prometheus": "tobs-kube-prometheus", "release": releaseName},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{"ReadWriteOnce"},
			Resources: corev1.ResourceRequirements{
				Requests: map[corev1.ResourceName]resource.Quantity{"storage": resource.MustParse("8Gi")},
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

func (c *Client) GetTimescaleDBURI(namespace, name string) (string, error) {
	secretName := name + "-timescaledb-uri"
	secrets, err := c.KubeGetAllSecrets(namespace)
	if err != nil {
		return "", err
	}

	for _, s := range secrets.Items {
		if s.Name == secretName {
			if bytepass, exists := s.Data["db-uri"]; exists {
				uriData := string(bytepass)
				return uriData, nil
			} else {
				// found the secret but failed to find the value with indexed key.
				return "", fmt.Errorf("could not get TimescaleDB URI with secret key index as db-uri from %s", secretName)
			}
		}
	}

	return "", nil
}

func (c *Client) GetDBPassword(secretKey, name, namespace string) ([]byte, error) {
	secret, err := c.KubeGetSecret(namespace, name+"-credentials")
	if err != nil {
		return nil, fmt.Errorf("could not get TimescaleDB password: %w", err)
	}

	if bytepass, exists := secret.Data[secretKey]; exists {
		return bytepass, nil
	}

	return nil, fmt.Errorf("user not found")
}