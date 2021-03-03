package cmd

import (
	"context"
	"fmt"

	"github.com/timescale/tobs/cli/pkg/utils"

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
	timescaledbChangePasswordCmd.Flags().StringP("user", "U", "PATRONI_SUPERUSER_PASSWORD", "user whose password to change")
	timescaledbChangePasswordCmd.Flags().StringP("dbname", "d", "postgres", "database name to connect to")
}

func timescaledbChangePassword(cmd *cobra.Command, args []string) error {
	var err error

	password := args[0]

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

	oldpassword, err := getOldPassword()
	if err != nil {
		return fmt.Errorf("could not get existing TimescaleDB password: %w", err)
	}

	err = updateDBPwdSecrets(user, password)
	if err != nil {
		return err
	}

	var dbUser string
	if user == "PATRONI_SUPERUSER_PASSWORD" {
		dbUser = "postgres"
	}

	_, err = pool.Exec(context.Background(), "ALTER USER "+dbUser+" WITH PASSWORD '"+password+"'")
	if err != nil {
		err1 := updateDBPwdSecrets(user, string(oldpassword))
		if err1 != nil {
			fmt.Printf("failed to revert to the old password on password update failure %v", err1)
		}
		return fmt.Errorf("could not change TimescaleDB password: %v", err)
	}

	uri, err := utils.GetTimescaleDBURI(namespace, name)
	if err != nil {
		return err
	}

	if uri != "" {
		secret, err := k8s.KubeGetSecret(namespace, name+"-timescaledb-uri")
		if err != nil {
			return fmt.Errorf("could not get TimescaleDB password: %w", err)
		}

		newURI, err := pgconn.UpdatePasswordInDBURI(uri, password)
		if err != nil {
			return fmt.Errorf("failed to upodate password in db uri: %w", err)
		}

		secret.Data["db-uri"] = []byte(newURI)
		err = k8s.KubeUpdateSecret(namespace, secret)
		if err != nil {
			return fmt.Errorf("could not change TimescaleDB password in external TimescaleDB uri secret: %w", err)
		}
	}

	return nil
}

func getOldPassword() ([]byte, error){
	secret, err := k8s.KubeGetSecret(namespace, name+"-credentials")
	if err != nil {
		return nil, fmt.Errorf("could not get TimescaleDB password: %w", err)
	}
	oldpassword := secret.Data[user]
	return oldpassword, nil
}

func updateDBPwdSecrets(user, password string) error {
	secret, err := k8s.KubeGetSecret(namespace, name+"-credentials")
	if err != nil {
		return fmt.Errorf("could not get TimescaleDB password: %w", err)
	}

	secret.Data[user] = []byte(password)
	err = k8s.KubeUpdateSecret(namespace, secret)
	if err != nil {
		return fmt.Errorf("could not change TimescaleDB password: %w", err)
	}

	secret, err = k8s.KubeGetSecret(namespace, name+"-grafana-db")
	if err != nil {
		return fmt.Errorf("could not get TimescaleDB password: %w", err)
	}

	dbUser, ok := secret.Data["GF_DATABASE_USER"]
	if  ok && string(dbUser) == user {
		secret.Data["GF_DATABASE_PASSWORD"] = []byte(password)
		err = k8s.KubeUpdateSecret(namespace, secret)
		if err != nil {
			return fmt.Errorf("could not change TimescaleDB password: %w", err)
		}
	}
	return nil
}