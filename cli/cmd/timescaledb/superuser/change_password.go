package superuser

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	root "github.com/timescale/tobs/cli/cmd"
	"github.com/timescale/tobs/cli/cmd/common"
	"github.com/timescale/tobs/cli/pkg/k8s"
	"github.com/timescale/tobs/cli/pkg/pgconn"
	"github.com/timescale/tobs/cli/pkg/utils"
)

// timescaledbChangePasswordCmd represents the timescaledb change-password command
var timescaledbChangePasswordCmd = &cobra.Command{
	Use:   "change-password",
	Short: "Changes the TimescaleDB super-user password",
	Args:  cobra.ExactArgs(1),
	RunE:  timescaledbChangePassword,
}

func init() {
	superuserCmd.AddCommand(timescaledbChangePasswordCmd)
}

func timescaledbChangePassword(cmd *cobra.Command, args []string) error {
	var err error

	password := args[0]

	fmt.Println("Changing password...")
	dbDetails, err := common.GetSuperuserDBDetails(root.Namespace, root.HelmReleaseName)
	if err != nil {
		return err
	}

	pool, err := dbDetails.OpenConnectionToDB()
	if err != nil {
		return fmt.Errorf("could not change TimescaleDB password: %w", err)
	}
	defer pool.Close()

	k8sClient := k8s.NewClient()
	oldpassword, err := utils.GetDBPassword(k8sClient, dbDetails.SecretKey, root.HelmReleaseName, root.Namespace)
	if err != nil {
		return fmt.Errorf("could not get existing TimescaleDB password: %w", err)
	}

	err = updateDBPwdSecrets(k8sClient, dbDetails.SecretKey, dbDetails.User, password)
	if err != nil {
		return err
	}

	_, err = pool.Exec(context.Background(), "ALTER USER "+dbDetails.User+" WITH PASSWORD '"+password+"'")
	if err != nil {
		err1 := updateDBPwdSecrets(k8sClient, dbDetails.SecretKey, dbDetails.User, string(oldpassword))
		if err1 != nil {
			fmt.Printf("failed to revert to the old password on password update failure %v", err1)
		}
		return fmt.Errorf("could not change TimescaleDB password: %w", err)
	}

	uri, err := utils.GetTimescaleDBURI(k8sClient, root.Namespace, root.HelmReleaseName)
	if err != nil {
		return err
	}

	if uri != "" {
		secret, err := k8sClient.KubeGetSecret(root.Namespace, root.HelmReleaseName+"-timescaledb-uri")
		if err != nil {
			return fmt.Errorf("could not get TimescaleDB password: %w", err)
		}

		newURI, err := pgconn.UpdatePasswordInDBURI(uri, password)
		if err != nil {
			return fmt.Errorf("failed to upodate password in db uri: %w", err)
		}

		secret.Data["db-uri"] = []byte(newURI)
		err = k8sClient.KubeUpdateSecret(root.Namespace, secret)
		if err != nil {
			return fmt.Errorf("could not change TimescaleDB password in external TimescaleDB uri secret: %w", err)
		}
	}

	return nil
}

func updateDBPwdSecrets(k8sClient k8s.Client, secretKey, user, password string) error {
	secret, err := k8sClient.KubeGetSecret(root.Namespace, root.HelmReleaseName+"-credentials")
	if err != nil {
		return fmt.Errorf("could not get TimescaleDB password: %w", err)
	}

	secret.Data[secretKey] = []byte(password)
	err = k8sClient.KubeUpdateSecret(root.Namespace, secret)
	if err != nil {
		return fmt.Errorf("could not change TimescaleDB password: %w", err)
	}

	secret, err = k8sClient.KubeGetSecret(root.Namespace, root.HelmReleaseName+"-grafana-db")
	if err != nil {
		return fmt.Errorf("could not get TimescaleDB password: %w", err)
	}

	dbUser, ok := secret.Data["GF_DATABASE_USER"]
	if ok && string(dbUser) == user {
		secret.Data["GF_DATABASE_PASSWORD"] = []byte(password)
		err = k8sClient.KubeUpdateSecret(root.Namespace, secret)
		if err != nil {
			return fmt.Errorf("could not change TimescaleDB password: %w", err)
		}
	}
	return nil
}
