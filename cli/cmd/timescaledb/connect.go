package timescaledb

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/cobra"
	root "github.com/timescale/tobs/cli/cmd"
	"github.com/timescale/tobs/cli/cmd/common"
	"github.com/timescale/tobs/cli/pkg/helm"
	"github.com/timescale/tobs/cli/pkg/k8s"
	"github.com/timescale/tobs/cli/pkg/pgconn"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// timescaledbConnectCmd represents the timescaledb connect command
var timescaledbConnectCmd = &cobra.Command{
	Use:   "connect <user>",
	Short: "Connects to the TimescaleDB database with provided user",
	Args:  cobra.ExactArgs(1),
	RunE:  timescaledbConnect,
}

func init() {
	Cmd.AddCommand(timescaledbConnectCmd)
	timescaledbConnectCmd.Flags().BoolP("master", "m", false, "directly execute session on master node")
	timescaledbConnectCmd.Flags().StringP("dbname", "d", "", "database name to connect to, defaults to dbname from the helm release")
}

func timescaledbConnect(cmd *cobra.Command, args []string) error {
	master, err := cmd.Flags().GetBool("master")
	if err != nil {
		return fmt.Errorf("could not connect to TimescaleDB: %w", err)
	}

	dbname, err := cmd.Flags().GetString("dbname")
	if err != nil {
		return fmt.Errorf("could not connect to TimescaleDB: %w", err)
	}

	if dbname == "" {
		// if dbname is empty get the default db name from helm release
		dbname, err = getDBNameFromValues()
		if err != nil {
			return fmt.Errorf("failed to get db name from helm values %v", err)
		}
	}

	if len(args) == 0 {
		return fmt.Errorf("provide db-user to connect or to use super-user use timescaledb superuser cmd")
	}

	dbDetails := &pgconn.DBDetails{
		Namespace:   root.Namespace,
		ReleaseName: root.HelmReleaseName,
		DBName:      dbname,
		User:        args[0],
	}
	k8sClient := k8s.NewClient()
	return PsqlConnect(k8sClient, dbDetails, master)
}

func getDBNameFromValues() (string, error) {
	helmClient := helm.NewClient(root.Namespace)
	dbName, err := helmClient.ExportValuesFieldFromRelease(root.HelmReleaseName, []string{"promscale", "connection", "dbName"})
	return fmt.Sprint(dbName), err
}

func PsqlConnect(k8sClient k8s.Client, dbDetails *pgconn.DBDetails, master bool) error {
	var host, psqlCMD string
	if master {
		masterpod, err := k8sClient.KubeGetPodName(root.Namespace, map[string]string{"release": root.HelmReleaseName, "role": "master"})
		if err != nil {
			return fmt.Errorf("could not connect to TimescaleDB: %w", err)
		}

		err = k8sClient.KubeExecCmd(root.Namespace, masterpod, "", "psql -U "+dbDetails.User, os.Stdin, true)
		if err != nil {
			return fmt.Errorf("could not connect to TimescaleDB: %w", err)
		}
		return nil
	}

	uri, err := common.GetTimescaleDBURI(k8sClient, root.Namespace, root.HelmReleaseName)
	if err != nil {
		return err
	}
	if uri == "" {
		host = root.HelmReleaseName + "." + root.Namespace + ".svc"
		psqlCMD = "psql -U " + dbDetails.User + " -h " + host + " " + dbDetails.DBName
	} else {
		psqlCMD = "psql " + uri
	}

	pod := formPsqlPodObject(dbDetails.DBName, root.Namespace, dbDetails.User, dbDetails.Password, host, uri)

	err = k8sClient.KubeCreatePod(pod)
	if err != nil {
		return fmt.Errorf("could not connect to TimescaleDB: %w", err)
	}

	time.Sleep(time.Second)

	defer func() {
		err = k8sClient.KubeDeletePod(root.Namespace, "psql")
		if err != nil {
			log.Fatalf("failed to delete psql pod %v", err)
		}
	}()

	err = k8sClient.KubeWaitOnPod(root.Namespace, "psql")
	if err != nil {
		return fmt.Errorf("failed to wait for psql pod: %w", err)
	}

	err = k8sClient.KubeExecCmd(root.Namespace, "psql", "", psqlCMD, os.Stdin, true)
	if err != nil {
		return fmt.Errorf("could not connect to TimescaleDB with psql pod: %w", err)
	}

	return nil
}

func formPsqlPodObject(name, namespace, user, pass, host, uri string) *corev1.Pod {
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
