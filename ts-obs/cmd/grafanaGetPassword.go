package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// grafanaGetPasswordCmd represents the grafana get-password command
var grafanaGetPasswordCmd = &cobra.Command{
	Use:   "get-password",
	Short: "Gets the admin password for Grafana",
	Args:  cobra.ExactArgs(0),
	RunE:  grafanaGetPassword,
}

func init() {
	grafanaCmd.AddCommand(grafanaGetPasswordCmd)
}

func grafanaGetPassword(cmd *cobra.Command, args []string) error {
	var err error

	var name string
	name, err = cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	var namespace string
	namespace, err = cmd.Flags().GetString("namespace")
	if err != nil {
		return err
	}

	secret, err := KubeGetSecret(namespace, name+"-grafana")
	if err != nil {
		return err
	}

	pass := secret.Data["admin-password"]
	fmt.Printf("Password: %v\n", string(pass))

	return nil
}
