package cmd

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/timescale/tobs/cli/pkg/k8s"

	"github.com/spf13/cobra"

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
	timescaledbConnectCmd.Flags().StringP("user", "U", "postgres", "database user name")
	timescaledbConnectCmd.Flags().BoolP("master", "m", false, "directly execute session on master node")
}

func timescaledbConnect(cmd *cobra.Command, args []string) error {
	var err error

	var user string
	user, err = cmd.Flags().GetString("user")
	if err != nil {
		return fmt.Errorf("could not connect to TimescaleDB: %w", err)
	}

	var master bool
	master, err = cmd.Flags().GetBool("master")
	if err != nil {
		return fmt.Errorf("could not connect to TimescaleDB: %w", err)
	}

	secret, err := k8s.KubeGetSecret(namespace, name+"-timescaledb-passwords")
	if err != nil {
		return fmt.Errorf("could not get TimescaleDB password: %w", err)
	}

	var pass string
	if bytepass, exists := secret.Data[user]; exists {
		pass = string(bytepass)
	} else {
		return fmt.Errorf("could not get TimescaleDB password: %w", errors.New("user not found"))
	}

	if master {
		masterpod, err := k8s.KubeGetPodName(namespace, map[string]string{"release": name, "role": "master"})
		if err != nil {
			return fmt.Errorf("could not connect to TimescaleDB: %w", err)
		}

		err = k8s.KubeExecCmd(namespace, masterpod, "", "psql -U postgres", os.Stdin, true)
		if err != nil {
			return fmt.Errorf("could not connect to TimescaleDB: %w", err)
		}
	} else {
		pod := getPodObject(name, namespace, user, pass)

		err = k8s.KubeCreatePod(pod)
		if err != nil {
			return fmt.Errorf("could not connect to TimescaleDB: %w", err)
		}

		time.Sleep(time.Second)

		err = k8s.KubeWaitOnPod(namespace, "psql")
		if err != nil {
			_ = k8s.KubeDeletePod(namespace, "psql")
			return fmt.Errorf("could not connect to TimescaleDB: %w", err)
		}
		err = k8s.KubeExecCmd(namespace, "psql", "", "psql -U "+user+" -h "+name+"."+namespace+".svc.cluster.local postgres", os.Stdin, true)
		if err != nil {
			_ = k8s.KubeDeletePod(namespace, "psql")
			return fmt.Errorf("could not connect to TimescaleDB: %w", err)
		}

		err = k8s.KubeDeletePod(namespace, "psql")
		if err != nil {
			return fmt.Errorf("could not connect to TimescaleDB: %w", err)
		}
		time.Sleep(3 * time.Second)
	}

	return nil
}

func getPodObject(name, namespace, user, pass string) *corev1.Pod {
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
					Args: []string{
						"-U",
						user,
						"-h",
						name + "." + namespace + ".svc.cluster.local",
						"postgres",
					},
				},
			},
		},
	}
}
