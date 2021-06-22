package timescaledb

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	root "github.com/timescale/tobs/cli/cmd"
	"github.com/timescale/tobs/cli/cmd/common"
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
	timescaledbChangePasswordCmd.Flags().StringP("dbname", "d", "postgres", "database name to connect to")
}

func timescaledbChangePassword(cmd *cobra.Command, args []string) error {
	var err error

	password := args[0]

	dbname, err := cmd.Flags().GetString("dbname")
	if err != nil {
		return fmt.Errorf("could not change TimescaleDB password: %w", err)
	}

	fmt.Println("Changing password...")

	d, err := common.FormDBDetails(user, dbname)
	if err != nil {
		return err
	}

	pool, err := d.OpenConnectionToDB()
	if err != nil {
		return fmt.Errorf("could not change TimescaleDB password: %w", err)
	}
	defer pool.Close()

	oldpassword, err := kubeClient.GetDBPassword(d.SecretKey, root.HelmReleaseName, root.Namespace)
	if err != nil {
		return fmt.Errorf("could not get existing TimescaleDB password: %w", err)
	}

	err = updateDBPwdSecrets(d.SecretKey, d.User, password)
	if err != nil {
		return err
	}

	_, err = pool.Exec(context.Background(), "ALTER USER "+d.User+" WITH PASSWORD '"+password+"'")
	if err != nil {
		err1 := updateDBPwdSecrets(d.SecretKey, d.User, string(oldpassword))
		if err1 != nil {
			fmt.Printf("failed to revert to the old password on password update failure %v", err1)
		}
		return fmt.Errorf("could not change TimescaleDB password: %w", err)
	}

	uri, err := kubeClient.GetTimescaleDBURI(root.Namespace, root.HelmReleaseName)
	if err != nil {
		return err
	}

	if uri != "" {
		secret, err := kubeClient.KubeGetSecret(root.Namespace, root.HelmReleaseName+"-timescaledb-uri")
		if err != nil {
			return fmt.Errorf("could not get TimescaleDB password: %w", err)
		}

		newURI, err := pgconn.UpdatePasswordInDBURI(uri, password)
		if err != nil {
			return fmt.Errorf("failed to upodate password in db uri: %w", err)
		}

		secret.Data["db-uri"] = []byte(newURI)
		err = kubeClient.KubeUpdateSecret(root.Namespace, secret)
		if err != nil {
			return fmt.Errorf("could not change TimescaleDB password in external TimescaleDB uri secret: %w", err)
		}
	}

	return nil
}

func updateDBPwdSecrets(secretKey, user, password string) error {
	secret, err := kubeClient.KubeGetSecret(root.Namespace, root.HelmReleaseName+"-credentials")
	if err != nil {
		return fmt.Errorf("could not get TimescaleDB password: %w", err)
	}

	secret.Data[secretKey] = []byte(password)
	err = kubeClient.KubeUpdateSecret(root.Namespace, secret)
	if err != nil {
		return fmt.Errorf("could not change TimescaleDB password: %w", err)
	}

	secret, err = kubeClient.KubeGetSecret(root.Namespace, root.HelmReleaseName+"-grafana-db")
	if err != nil {
		return fmt.Errorf("could not get TimescaleDB password: %w", err)
	}

	dbUser, ok := secret.Data["GF_DATABASE_USER"]
	if ok && string(dbUser) == user {
		secret.Data["GF_DATABASE_PASSWORD"] = []byte(password)
		err = kubeClient.KubeUpdateSecret(root.Namespace, secret)
		if err != nil {
			return fmt.Errorf("could not change TimescaleDB password: %w", err)
		}
	}
	return nil
}
