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
		return fmt.Errorf("could not get Grafana password: %w", err)
	}

	var namespace string
	namespace, err = cmd.Flags().GetString("namespace")
	if err != nil {
		return fmt.Errorf("could not get Grafana password: %w", err)
	}

	secret, err := KubeGetSecret(namespace, name+"-grafana")
	if err != nil {
		return fmt.Errorf("could not get Grafana password: %w", err)
	}

	pass := secret.Data["admin-password"]
	fmt.Println(string(pass))

	return nil
}
