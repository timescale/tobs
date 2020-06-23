package cmd

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// timescaledbConnectCmd represents the timescaledb connect command
var timescaledbConnectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Connects to the TimescaleDB/PostgreSQL database",
	Args:  cobra.ExactArgs(0),
	RunE:  timescaledbConnect,
}

func init() {
	timescaledbCmd.AddCommand(timescaledbConnectCmd)
	timescaledbConnectCmd.Flags().StringP("user", "U", "postgres", "database user name")
	timescaledbConnectCmd.Flags().StringP("password", "p", "PGPASSWORD", "environment variable where password is stored")
	timescaledbConnectCmd.Flags().BoolP("master", "m", false, "directly execute session on master node")
}

func timescaledbConnect(cmd *cobra.Command, args []string) error {
	var err error

	var user string
	user, err = cmd.Flags().GetString("user")
	if err != nil {
		return fmt.Errorf("could not connect to TimescaleDB: %w", err)
	}

	var password string
	password, err = cmd.Flags().GetString("password")
	if err != nil {
		return fmt.Errorf("could not connect to TimescaleDB: %w", err)
	}

	var master bool
	master, err = cmd.Flags().GetBool("master")
	if err != nil {
		return fmt.Errorf("could not connect to TimescaleDB: %w", err)
	}

	var name string
	name, err = cmd.Flags().GetString("name")
	if err != nil {
		return fmt.Errorf("could not connect to TimescaleDB: %w", err)
	}

	var namespace string
	namespace, err = cmd.Flags().GetString("namespace")
	if err != nil {
		return fmt.Errorf("could not connect to TimescaleDB: %w", err)
	}

	if (password != "") == master {
		return fmt.Errorf("could not connect to TimescaleDB: %w", errors.New("must connect through one of user/password or master"))
	}

	if master {
		masterpod, err := KubeGetPodName(namespace, map[string]string{"release": name, "role": "master"})
		if err != nil {
			return fmt.Errorf("could not connect to TimescaleDB: %w", err)
		}

		err = KubeExecCmd(namespace, masterpod, "", "psql -U postgres", os.Stdin, true)
		if err != nil {
			return fmt.Errorf("could not connect to TimescaleDB: %w", err)
		}
	} else {
		pod := getPodObject(name, user, password)

		err = KubeCreatePod(pod)
		if err != nil {
			return fmt.Errorf("could not connect to TimescaleDB: %w", err)
		}

		err = KubeWaitOnPod(namespace, "psql")
		if err != nil {
			return fmt.Errorf("could not connect to TimescaleDB: %w", err)
		}
		err = KubeExecCmd(namespace, "psql", "", "psql -U "+user+" -h "+name+".default.svc.cluster.local postgres", os.Stdin, true)
		if err != nil {
			return fmt.Errorf("could not connect to TimescaleDB: %w", err)
		}

		err = KubeDeletePod(namespace, "psql")
		if err != nil {
			return fmt.Errorf("could not connect to TimescaleDB: %w", err)
		}
		time.Sleep(3 * time.Second)
	}

	return nil
}

func getPodObject(name, user, password string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "psql",
			Namespace: "default",
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
							Value: os.Getenv(password),
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
						name + ".default.svc.cluster.local",
						"postgres",
					},
				},
			},
		},
	}
}
