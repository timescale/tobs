package cmd

import (
	"fmt"
	"github.com/timescale/tobs/cli/pkg/k8s"

	"github.com/spf13/cobra"
)

// grafanaChangePasswordCmd represents the grafana change-password command
var grafanaChangePasswordCmd = &cobra.Command{
	Use:   "change-password <password>",
	Short: "Changes the admin password for Grafana",
	Args:  cobra.ExactArgs(1),
	RunE:  grafanaChangePassword,
}

func init() {
	grafanaCmd.AddCommand(grafanaChangePasswordCmd)
}

func grafanaChangePassword(cmd *cobra.Command, args []string) error {
	var err error

	password := args[0]

	secret, err := k8s.KubeGetSecret(namespace, name+"-grafana")
	if err != nil {
		return fmt.Errorf("could not change Grafana password: %w", err)
	}

	oldpassword := secret.Data["admin-password"]

	secret.Data["admin-password"] = []byte(password)
	err = k8s.KubeUpdateSecret(namespace, secret)
	if err != nil {
		return fmt.Errorf("could not change Grafana password: %w", err)
	}

	fmt.Println("Changing password...")
	grafanaPod, err := k8s.KubeGetPodName(namespace, map[string]string{"app.kubernetes.io/instance": name, "app.kubernetes.io/name": "grafana"})
	if err != nil {
		return fmt.Errorf("could not change Grafana password: %w", err)
	}

	err = k8s.KubeExecCmd(namespace, grafanaPod, "grafana", "grafana-cli admin reset-admin-password "+password, nil, false)
	if err != nil {
		secret.Data["admin-password"] = oldpassword
		_ = k8s.KubeUpdateSecret(namespace, secret)
		return fmt.Errorf("could not change Grafana password: %w", err)
	}

	return nil
}
