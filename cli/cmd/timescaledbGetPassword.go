package cmd

import (
	"errors"
	"fmt"

	"github.com/timescale/tobs/cli/pkg/k8s"

	"github.com/spf13/cobra"
)

// timescaledbGetPasswordCmd represents the timescaledb get-password command
var timescaledbGetPasswordCmd = &cobra.Command{
	Use:   "get-password",
	Short: "Gets the TimescaleDB password for a specific user",
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
		return fmt.Errorf("could not get TimescaleDB password: %w", err)
	}

	secret, err := k8s.KubeGetSecret(namespace, name+"-timescaledb-passwords")
	if err != nil {
		return fmt.Errorf("could not get TimescaleDB password: %w", err)
	}

	var pass string
	if bytepass, exists := secret.Data[user]; exists {
		pass = string(bytepass)
	} else {
		return fmt.Errorf("could not get TimescaleDB password: %w", errors.New("user not found"))
	}
	fmt.Println(pass)

	return nil
}
