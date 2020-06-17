package cmd

import (
    "errors"
	"fmt"

	"github.com/spf13/cobra"
)

// grafanaGetInitialPasswordCmd represents the grafana get-initial-password command
var grafanaGetInitialPasswordCmd = &cobra.Command{
	Use:   "get-initial-password",
	Short: "Gets the initial admin password for Grafana",
	RunE:  grafanaGetInitialPassword,
}

func init() {
	grafanaCmd.AddCommand(grafanaGetInitialPasswordCmd)
}

func grafanaGetInitialPassword(cmd *cobra.Command, args []string) error {
    var err error

    if len(args) != 0 {
        return errors.New("\"ts-obs grafana get-password\" requires 0 arguments")
    }

    secret, err := kubeGetSecret("ts-obs-grafana")
    if err != nil {
        return err
    }

    pass := secret.Data["admin-password"]
    fmt.Printf("Password: %v\n", string(pass))

    return nil
}
