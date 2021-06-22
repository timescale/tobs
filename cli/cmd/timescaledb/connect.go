package timescaledb

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	root "github.com/timescale/tobs/cli/cmd"
	"github.com/timescale/tobs/cli/cmd/common"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// timescaledbConnectCmd represents the timescaledb connect command
var timescaledbConnectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Connects to the TimescaleDB database",
	Args:  cobra.ExactArgs(0),
	RunE:  timescaledbConnect,
}

func init() {
	timescaledbCmd.AddCommand(timescaledbConnectCmd)
	timescaledbConnectCmd.Flags().BoolP("master", "m", false, "directly execute session on master node")
	timescaledbConnectCmd.Flags().StringP("dbname", "d", "postgres", "database name to connect to")
}

func timescaledbConnect(cmd *cobra.Command, args []string) error {
	var err error

	var host, psqlCMD string

	var master bool
	master, err = cmd.Flags().GetBool("master")
	if err != nil {
		return fmt.Errorf("could not connect to TimescaleDB: %w", err)
	}

	dbname, err := cmd.Flags().GetString("dbname")
	if err != nil {
		return fmt.Errorf("could not change TimescaleDB password: %w", err)
	}

	secret, err := kubeClient.KubeGetSecret(root.Namespace, root.HelmReleaseName+"-credentials")
	if err != nil {
		return fmt.Errorf("could not get TimescaleDB password: %w", err)
	}

	secretKey, user, err := common.GetDBSecretKeyAndDBUser(root.HelmReleaseName, user)
	if err != nil {
		return fmt.Errorf("could not get DB secret key from helm release: %w", err)
	}

	var pass string
	if bytepass, exists := secret.Data[secretKey]; exists {
		pass = string(bytepass)
	} else {
		return fmt.Errorf("could not get TimescaleDB password: %w", errors.New("user not found"))
	}

	uri, err := kubeClient.GetTimescaleDBURI(root.Namespace, root.HelmReleaseName)
	if err != nil {
		return err
	}

	if master {
		masterpod, err := kubeClient.KubeGetPodName(root.Namespace, map[string]string{"release": root.HelmReleaseName, "role": "master"})
		if err != nil {
			return fmt.Errorf("could not connect to TimescaleDB: %w", err)
		}

		err = kubeClient.KubeExecCmd(root.Namespace, masterpod, "", "psql -U "+user, os.Stdin, true)
		if err != nil {
			return fmt.Errorf("could not connect to TimescaleDB: %w", err)
		}
	} else {
		if uri == "" {
			host = root.HelmReleaseName + "." + root.Namespace + ".svc.cluster.local"
			psqlCMD = "psql -U " + user + " -h " + host + " " + dbname
		} else {
			psqlCMD = "psql " + uri
		}

		pod := getPodObject(dbname, root.Namespace, user, pass, host, uri)

		err = kubeClient.KubeCreatePod(pod)
		if err != nil {
			return fmt.Errorf("could not connect to TimescaleDB: %w", err)
		}

		time.Sleep(time.Second)

		err = kubeClient.KubeWaitOnPod(root.Namespace, "psql")
		if err != nil {
			_ = kubeClient.KubeDeletePod(root.Namespace, "psql")
			return fmt.Errorf("could not connect to TimescaleDB: %w", err)
		}

		err = kubeClient.KubeExecCmd(root.Namespace, "psql", "", psqlCMD, os.Stdin, true)
		if err != nil {
			_ = kubeClient.KubeDeletePod(root.Namespace, "psql")
			return fmt.Errorf("could not connect to TimescaleDB: %w", err)
		}

		err = kubeClient.KubeDeletePod(root.Namespace, "psql")
		if err != nil {
			return fmt.Errorf("could not connect to TimescaleDB: %w", err)
		}
	}

	return nil
}

func getPodObject(name, namespace, user, pass, host, uri string) *corev1.Pod {
	var args []string
	if uri == "" {
		args = []string{"-U", user, "-h", host, name}
	} else {
		args = []string{uri}
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "psql",
			Namespace: namespace,
			Labels: map[string]string{
				"app": "psql",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:            "postgres",
					Image:           "postgres",
					ImagePullPolicy: corev1.PullIfNotPresent,
					Env: []corev1.EnvVar{
						{
							Name:  "PGPASSWORD",
							Value: pass,
						},
					},
					Stdin: true,
					TTY:   true,
					Command: []string{
						"psql",
					},
					Args: args,
				},
			},
		},
	}
}
