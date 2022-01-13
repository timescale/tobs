package k8s

import (
	"io"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/portforward"
)

type Client interface {
	// pod specific actions
	KubeCreatePod(pod *corev1.Pod) error
	KubeDeletePod(namespace string, podName string) error
	KubeWaitOnPod(namespace string, podName string) error
	KubeGetPodName(namespace string, labelmap map[string]string) (string, error)
	KubeGetPods(namespace string, labelmap map[string]string) ([]corev1.Pod, error)
	KubeGetAllPods(namespace string, name string) ([]corev1.Pod, error)
	DeletePods(namespace string, labels map[string]string, forceKill bool) error
	KubePortForwardPod(namespace string, podName string, local int, remote int) (*portforward.PortForwarder, error)

	// deployment specific actions
	GetDeployment(name, namespace string) (*appsv1.Deployment, error)
	UpdateDeployment(deployment *appsv1.Deployment) error
	DeleteDeployment(labels map[string]string, namespace string) error

	// service specific actions
	KubeGetServiceName(namespace string, labelmap map[string]string) (string, error)
	KubeDeleteService(namespace string, serviceName string) error
	KubeDeleteEndpoint(namespace string, endpointName string) error
	KubePortForwardService(namespace string, serviceName string, local int, remote int) (*portforward.PortForwarder, error)

	// job specific actions
	CreateJob(job *batchv1.Job) error
	GetJob(jobName, namespace string) (*batchv1.Job, error)
	DeleteJob(name, namespace string) error

	// daemonset specific actions
	DeleteDaemonset(name, namespace string) error

	// pvc specific actions
	KubeGetPVCNames(namespace string, labelmap map[string]string) ([]string, error)
	KubeDeletePVC(namespace string, PVCName string) error
	GetPVCSizes(namespace, pvcPrefix string, labels map[string]string) ([]*PVCData, error)
	ExpandPVCsForAllPods(namespace, value, pvcPrefix string, labels map[string]string) (map[string]string, error)
	ExpandPVC(namespace, pvcName, value string) error

	// pv specific actions
	UpdatePVToNewPVC(pvcName, newPVCName, namespace string, pvcLabels map[string]string) error

	// secrets specific actions
	KubeGetSecret(namespace string, secretName string) (*corev1.Secret, error)
	KubeGetAllSecrets(namespace string) (*corev1.SecretList, error)
	KubeUpdateSecret(namespace string, secret *corev1.Secret) error
	CreateSecret(secret *corev1.Secret) error
	DeleteSecret(secretName, namespace string) error
	CheckSecretExists(secretName, namespace string) (bool, error)

	// exec into the container
	KubeExecCmd(namespace string, podName string, container string, command string, stdin io.Reader, tty bool) error

	// namespace specific actions
	CreateNamespaceIfNotExists(namespace string) error
	UpdateNamespaceLabels(name string, labels map[string]string) error
	GetNamespaceLabels(name string) (map[string]string, error)

	// CR operations
	CreateCustomResource(namespace, apiVersion, resourceName string, body []byte) error
	DeleteCustomResource(namespace, apiVersion, resourceName, crName string) error

	// List cert-manager Certificate resources
	ListCertManagerDeprecatedCRs() ([]ResourceDetails, error)

	// Apply manifests
	ApplyManifests(map[string]string) error
}