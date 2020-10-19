package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/timescale/tobs/cli/pkg/k8s"
	"github.com/timescale/tobs/cli/pkg/pgconn"
)

// timescaledbChangePasswordCmd represents the timescaledb change-password command
var timescaledbChangePasswordCmd = &cobra.Command{
	Use:   "change-password",
	Short: "Changes the TimescaleDB password for a specific user",
	Args:  cobra.ExactArgs(1),
	RunE:  timescaledbChangePassword,
}

func init() {
	timescaledbCmd.AddCommand(timescaledbChangePasswordCmd)
	timescaledbChangePasswordCmd.Flags().StringP("user", "U", "postgres", "user whose password to change")
	timescaledbChangePasswordCmd.Flags().StringP("dbname", "d", "postgres", "database name to connect to")
}

func timescaledbChangePassword(cmd *cobra.Command, args []string) error {
	var err error

	password := args[0]

	var user string
	user, err = cmd.Flags().GetString("user")
	if err != nil {
		return fmt.Errorf("could not change TimescaleDB password: %w", err)
	}

	var dbname string
	dbname, err = cmd.Flags().GetString("dbname")
	if err != nil {
		return fmt.Errorf("could not change TimescaleDB password: %w", err)
	}

	fmt.Println("Changing password...")
	pool, err := pgconn.OpenConnectionToDB(namespace, name, user, dbname, FORWARD_PORT_TSDB)
	if err != nil {
		return fmt.Errorf("could not change TimescaleDB password: %w", err)
	}
	defer pool.Close()

	secret, err := k8s.KubeGetSecret(namespace, name+"-timescaledb-passwords")
	if err != nil {
		return fmt.Errorf("could not get TimescaleDB password: %w", err)
	}

	oldpassword := secret.Data[user]

	secret.Data[user] = []byte(password)
	err = k8s.KubeUpdateSecret(namespace, secret)
	if err != nil {
		return fmt.Errorf("could not change TimescaleDB password: %w", err)
	}
	_, err = pool.Exec(context.Background(), "ALTER USER "+user+" WITH PASSWORD '"+password+"'")
	if err != nil {
		secret.Data[user] = oldpassword
		_ = k8s.KubeUpdateSecret(namespace, secret)
		return fmt.Errorf("could not change TimescaleDB password: %w", err)
	}

	return nil
}
