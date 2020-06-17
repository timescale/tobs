package cmd

import (
    "context"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"

    corev1 "k8s.io/api/core/v1"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/kubernetes/scheme"
    rest "k8s.io/client-go/rest"
    "k8s.io/client-go/transport/spdy"
    "k8s.io/client-go/tools/clientcmd"
    "k8s.io/client-go/tools/portforward"
    "k8s.io/client-go/tools/remotecommand"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/labels"
)

func kubeInit() (kubernetes.Interface, *rest.Config) {
    var err error

    config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: os.Getenv("HOME") + "/.kube/config"},
		&clientcmd.ConfigOverrides{CurrentContext: "minikube"}).ClientConfig()
    if err != nil {
        log.Fatal(err)
    }

    client, err := kubernetes.NewForConfig(config)
    if err != nil {
        log.Fatal(err)
    }

    return client, config
}

func kubeGetPodName(labelmap map[string]string) (string, error) {
    var err error

    client, _ := kubeInit()

    labelSelector := metav1.LabelSelector{MatchLabels: labelmap}
    listOptions := metav1.ListOptions{
        LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
    }

    pods, err := client.CoreV1().Pods("default").List(context.Background(), listOptions)
    if err != nil {
        return "", err
    }

    return pods.Items[0].Name, nil
}

func kubeGetServiceName(labelmap map[string]string) (string, error) {
    var err error

    client, _ := kubeInit()

    labelSelector := metav1.LabelSelector{MatchLabels: labelmap}
    listOptions := metav1.ListOptions{
        LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
    }

    services, err := client.CoreV1().Services("default").List(context.Background(), listOptions)
    if err != nil {
        return "", err
    }

    return services.Items[0].Name, nil
}

func kubeGetSecret(secretName string) (*corev1.Secret, error) {
    var err error

    client, _ := kubeInit()

    secret, err := client.CoreV1().Secrets("default").Get(context.Background(), secretName, metav1.GetOptions{})
    if err != nil {
        return nil, err
    }

    return secret, nil
}

func kubeGetPVCNames(labelmap map[string]string) ([]string, error) {
    var err error

    client, _ := kubeInit()

    labelSelector := metav1.LabelSelector{MatchLabels: labelmap}
    listOptions := metav1.ListOptions{
        LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
    }

    pvcs, err := client.CoreV1().PersistentVolumeClaims("default").List(context.Background(), listOptions)
    if err != nil {
        return []string{}, err
    }

    var names []string 
    for _, pvc := range pvcs.Items {
        names = append(names, pvc.Name)
    }

    return names, nil
}

// ExecCmd exec command on specific pod and wait the command's output.
func kubeExecCmd(podName string, container string, command string, stdin io.Reader, tty bool) error {
    var err error

    client, config := kubeInit()

    cmd := []string{
        "/bin/sh",
        "-c",
        command,
    }
    req := client.CoreV1().RESTClient().Post().Resource("pods").Namespace("default").
        Name(podName).SubResource("exec")
    option := &corev1.PodExecOptions{
        Container: container,
        Command: cmd,
        Stdin:   true,
        Stdout:  true,
        Stderr:  true,
        TTY:     tty,
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

func kubePortForwardPod(podName string, local int, remote int) error {
    var err error

    client, config := kubeInit() 

    url := client.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace("default").
		Name(podName).
		SubResource("portforward").URL()

    transport, upgrader, err := spdy.RoundTripperFor(config)
	if err != nil {
		return err
	}

    dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", url)

    ports := []string{fmt.Sprintf("%d:%d", local, remote)}
    
    pf, err := portforward.New(dialer, ports, make(chan struct{}, 1), make(chan struct{}, 1), os.Stdout, os.Stderr)
    if err != nil {
        return err
    }

    errChan := make(chan error)
	go func() {
		errChan <- pf.ForwardPorts()
	}()

    select {
	case err = <- errChan:
		return err
	case <-pf.Ready:
		return nil
	}
}

func kubePortForwardService(serviceName string, local int, remote int) error {
    var err error

    client, _ := kubeInit()

    service, err :=  client.CoreV1().Services("default").Get(context.Background(), serviceName, metav1.GetOptions{})

    set := labels.Set(service.Spec.Selector)
    listOptions := metav1.ListOptions{LabelSelector: set.AsSelector().String()}
    pods, err :=  client.CoreV1().Pods("default").List(context.Background(), listOptions)
    
    podName := pods.Items[0].Name

    err = kubePortForwardPod(podName, local, remote)
    if err != nil {
        return err
    }

    return nil
}

func kubeCreatePod(pod *corev1.Pod) error {
    var err error

    client, _ := kubeInit()

    fmt.Println("Creating pod...")
    _, err = client.CoreV1().Pods(pod.Namespace).Create(context.Background(), pod, metav1.CreateOptions{})
    if err != nil {
        return err
    }

    return nil
}

func kubeDeletePod(podName string) error {
    var err error

    client, _ := kubeInit()

    fmt.Println("Deleting pod...")
    err = client.CoreV1().Pods("default").Delete(context.Background(), podName, metav1.DeleteOptions{})
    if err != nil {
        return err
    }

    return nil
}

func kubeWaitOnPod(podName string) error {
    client, _ := kubeInit()

    for {
        pod, err := client.CoreV1().Pods("default").Get(context.Background(), podName, metav1.GetOptions{})
        if err != nil {
            return err
        }
        if pod.Status.Phase != corev1.PodPending {
            break
        }
    }

    return nil
}

func kubeDeleteService(serviceName string) error {
    var err error

    client, _ := kubeInit()

    fmt.Println("Deleting service...")
    err = client.CoreV1().Services("default").Delete(context.Background(), serviceName, metav1.DeleteOptions{})
    if err != nil {
        return err
    }

    return nil
}

func kubeDeleteEndpoint(endpointName string) error {
    var err error

    client, _ := kubeInit()

    fmt.Println("Deleting endpoint...")
    err = client.CoreV1().Endpoints("default").Delete(context.Background(), endpointName, metav1.DeleteOptions{})
    if err != nil {
        return err
    }

    return nil
}

func kubeDeletePVC(PVCName string) error {
    var err error

    client, _ := kubeInit()

    fmt.Println("Deleting PVC...")
    err = client.CoreV1().PersistentVolumeClaims("default").Delete(context.Background(), PVCName, metav1.DeleteOptions{})
    if err != nil {
        return err
    }

    return nil
}
