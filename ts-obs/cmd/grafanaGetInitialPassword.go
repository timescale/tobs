package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// grafanaGetInitialPasswordCmd represents the grafana get-initial-password command
var grafanaGetInitialPasswordCmd = &cobra.Command{
	Use:   "get-initial-password",
	Short: "Gets the initial admin password for Grafana",
	Args:  cobra.ExactArgs(0),
	RunE:  grafanaGetInitialPassword,
}

func init() {
	grafanaCmd.AddCommand(grafanaGetInitialPasswordCmd)
}

func grafanaGetInitialPassword(cmd *cobra.Command, args []string) error {
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
