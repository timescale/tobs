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

	"github.com/timescale/tobs/cli/cmd"
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

func kubeInit() (kubernetes.Interface, *rest.Config) {
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

	return client, config
}

func KubeGetPodName(namespace string, labelmap map[string]string) (string, error) {
	var err error
	var podName string
	client, _ := kubeInit()

	labelSelector := metav1.LabelSelector{MatchLabels: labelmap}
	listOptions := metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
	}

	pods, err := client.CoreV1().Pods(namespace).List(context.Background(), listOptions)
	if err != nil {
		return podName, err
	}

	if len(pods.Items) > 0 {
		podName = pods.Items[0].Name
	}

	return podName, nil
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

func KubeGetAllSecrets(namespace string) (*corev1.SecretList, error) {
	var err error

	client, _ := kubeInit()

	secrets, err := client.CoreV1().Secrets(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return secrets, nil
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
	gracePeriodSecs := int64(0)
	err = client.CoreV1().Pods(namespace).Delete(context.Background(), podName, metav1.DeleteOptions{GracePeriodSeconds: &gracePeriodSecs})
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
	Name       string
	SpecSize   string
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
			return nil, fmt.Errorf("failed to get the pod's for pvc %s with labelSet: %v", pvcPrefix, labels)
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

func ExpandPVCsForAllPods(namespace, value, pvcPrefix string, labels map[string]string) (map[string]string, error) {
	pvcResults := make(map[string]string)
	pods, err := KubeGetPods(namespace, labels)
	if err != nil {
		return pvcResults, fmt.Errorf("failed to get the pods using labels %w", err)
	}

	pvcs := buildPVCNames(pvcPrefix, pods)
	for _, pvc := range pvcs {
		err := ExpandPVC(namespace, pvc, value)
		if err != nil {
			fmt.Println(fmt.Errorf("%w", err))
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

func DeletePods(namespace string, labels map[string]string, forceKill bool) error {
	pods, err := KubeGetPods(namespace, labels)
	if err != nil {
		return fmt.Errorf("failed to get the pods using labels %w", err)
	}

	var deleteOptions metav1.DeleteOptions
	if forceKill {
		gracePeriodSecs := int64(0)
		deleteOptions = metav1.DeleteOptions{GracePeriodSeconds: &gracePeriodSecs}
	}

	client, _ := kubeInit()
	for _, pod := range pods {
		err = client.CoreV1().Pods(namespace).Delete(context.Background(), pod.Name, deleteOptions)
		if err != nil {
			return fmt.Errorf("failed to delete the pod: %s %v\n", pod.Name, err)
		}
	}
	return nil
}

func CreateSecret(secret *corev1.Secret) error {
	client, _ := kubeInit()
	_, err := client.CoreV1().Secrets(secret.Namespace).Create(context.Background(), secret, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create %s secret %v", secret.Name, err)
	}
	return nil
}

func DeleteSecret(secretName, namespace string) error {
	client, _ := kubeInit()
	err := client.CoreV1().Secrets(namespace).Delete(context.Background(), secretName, metav1.DeleteOptions{})
	if err != nil {
		ok := errors2.IsNotFound(err)
		if !ok {
			return fmt.Errorf("failed to delete %s secret %v", secretName, err)
		}
	}
	return nil
}

func CreateNamespaceIfNotExists(namespace string) error {
	client, _ := kubeInit()
	namespaces, err := client.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
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
	_, err = client.CoreV1().Namespaces().Create(context.Background(), n, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create the namespace %v", err)
	}

	return nil
}

func CheckSecretExists(secretName, namespace string) (bool, error) {
	client, _ := kubeInit()
	secExists, err := client.CoreV1().Secrets(namespace).Get(context.Background(), secretName, metav1.GetOptions{})
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

func DeleteJob(name, namespace string) error {
	client, _ := kubeInit()
	// deleting the job leaves the completed pods
	// this might cause pvc to be in terminating state forever
	// so delete the job with propagation as background
	propagation := metav1.DeletePropagationBackground
	err := client.BatchV1().Jobs(namespace).Delete(context.Background(), name, metav1.DeleteOptions{PropagationPolicy: &propagation})
	if err != nil {
		return err
	}
	return nil
}

func DeleteDaemonset(name, namespace string) error {
	client, _ := kubeInit()
	fmt.Printf("Deleting daemonset %v...\n", name)
	err := client.AppsV1().DaemonSets(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

func CreateJob(job *batchv1.Job) error {
	client, _ := kubeInit()
	_, err := client.BatchV1().Jobs(job.Namespace).Create(context.Background(), job, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return err
}

func GetJob(jobName, namespace string) (*batchv1.Job, error) {
	client, _ := kubeInit()
	return client.BatchV1().Jobs(namespace).Get(context.Background(), jobName, metav1.GetOptions{})
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

func GetDeployment(name, namespace string) (*appsv1.Deployment, error) {
	client, _ := kubeInit()
	return client.AppsV1().Deployments(namespace).Get(context.Background(), name, metav1.GetOptions{})
}

func UpdateDeployment(deployment *appsv1.Deployment) error {
	client, _ := kubeInit()
	_, err := client.AppsV1().Deployments(deployment.Namespace).Update(context.Background(), deployment, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return err
}

func UpdatePrometheusPV(pvcName, newPVCName, namespace string) error {
	client, _ := kubeInit()
	pvc, err := client.CoreV1().PersistentVolumeClaims(namespace).Get(context.Background(), pvcName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get persistent volume claim %s: %v", pvcName, err)
	}

	pvName := pvc.Spec.VolumeName
	pv, err := client.CoreV1().PersistentVolumes().Get(context.Background(), pvName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get the persistent volume %s: %v", pvName, err)
	}

	pv.Spec.ClaimRef = nil
	pv, err = client.CoreV1().PersistentVolumes().Update(context.Background(), pv, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update persistent volume %s: %v", pv.Name, err)
	}

	newPVC := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      newPVCName,
			Namespace: namespace,
			Labels:    map[string]string{"app": "prometheus", "prometheus": "tobs-kube-prometheus", "release": cmd.HelmReleaseName},
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

	_, err = client.CoreV1().PersistentVolumeClaims(namespace).Create(context.Background(), newPVC, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to get persistent volume claim %s: %v", pvcName, err)
	}

	return nil
}
