package cmd

import (
    "errors"

	"github.com/spf13/cobra"
)

// grafanaChangePasswordCmd represents the grafana change-password command
var grafanaChangePasswordCmd = &cobra.Command{
	Use:   "change-password <password>",
	Short: "Changes the admin password for Grafana",
	RunE:  grafanaChangePassword,
}

func init() {
	grafanaCmd.AddCommand(grafanaChangePasswordCmd)
}

func grafanaChangePassword(cmd *cobra.Command, args []string) error {
    var err error
    
    if len(args) != 1 {
        return errors.New("\"ts-obs grafana change-password\" requires 1 argument")
    }

    password := args[0]

    grafanaPod, err := kubeGetPodName(map[string]string{"app.kubernetes.io/name" : "grafana"})
    if err != nil {
        return err
    }

    err = kubeExecCmd(grafanaPod, "grafana", "grafana-cli admin reset-admin-password " + password, nil, false)

    return nil
}
