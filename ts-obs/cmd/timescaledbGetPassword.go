package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// timescaledbGetPasswordCmd represents the timescaledb get-password command
var timescaledbGetPasswordCmd = &cobra.Command{
	Use:   "get-password",
	Short: "Gets the TimescaleDB/PostgreSQL password for a specific user",
	Args:  cobra.ExactArgs(0),
	RunE:  timescaledbGetPassword,
}

func init() {
	timescaledbCmd.AddCommand(timescaledbGetPasswordCmd)
	timescaledbGetPasswordCmd.Flags().StringP("user", "U", "postgres", "user whose password to get")
}

func timescaledbGetPassword(cmd *cobra.Command, args []string) error {
	var err error

	var user string
	user, err = cmd.Flags().GetString("user")
	if err != nil {
		return err
	}

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

	secret, err := KubeGetSecret(namespace, name+"-timescaledb-passwords")
	if err != nil {
		return err
	}

	pass := secret.Data[user]
	fmt.Printf("Password: %v\n", string(pass))

	return nil
}
