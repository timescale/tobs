package grafana

import (
	"fmt"

	"github.com/spf13/cobra"
	root "github.com/timescale/tobs/cli/cmd"
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

	secret, err := kubeClient.KubeGetSecret(root.Namespace, root.HelmReleaseName+"-grafana")
	if err != nil {
		return fmt.Errorf("could not change Grafana password: %w", err)
	}

	oldpassword := secret.Data["admin-password"]

	secret.Data["admin-password"] = []byte(password)
	err = kubeClient.KubeUpdateSecret(root.Namespace, secret)
	if err != nil {
		return fmt.Errorf("could not change Grafana password: %w", err)
	}

	fmt.Println("Changing password...")
	grafanaPod, err := kubeClient.KubeGetPodName(root.Namespace, map[string]string{"app.kubernetes.io/instance": root.HelmReleaseName, "app.kubernetes.io/name": "grafana"})
	if err != nil {
		return fmt.Errorf("could not change Grafana password: %w", err)
	}

	err = kubeClient.KubeExecCmd(root.Namespace, grafanaPod, "grafana", "grafana-cli admin reset-admin-password "+password, nil, false)
	if err != nil {
		err1 := updateToOldPassword(oldpassword)
		if err1 != nil {
			// on failure just print the error, to indicate users the there is an inconsistency in pwd change.
			fmt.Println(err1)
		}
		return fmt.Errorf("could not change Grafana password: %s", err)
	}

	return nil
}

func updateToOldPassword(oldpassword []byte) error {
	secret, err := kubeClient.KubeGetSecret(root.Namespace, root.HelmReleaseName+"-grafana")
	if err != nil {
		return fmt.Errorf("could not change Grafana password: %w", err)
	}

	secret.Data["admin-password"] = oldpassword
	err = kubeClient.KubeUpdateSecret(root.Namespace, secret)
	if err != nil {
		return fmt.Errorf("failed to update secret to old password on change password failure %v", err)
	}
	return nil
}
