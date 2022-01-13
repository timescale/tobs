package superuser

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	root "github.com/timescale/tobs/cli/cmd"
	"github.com/timescale/tobs/cli/cmd/common"
	"github.com/timescale/tobs/cli/pkg/k8s"
	"github.com/timescale/tobs/cli/pkg/pgconn"
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
	err = updateDBPwdSecrets(k8sClient, dbDetails.SecretKey, dbDetails.User, password)
	if err != nil {
		return err
	}

	_, err = pool.Exec(context.Background(), "ALTER USER "+dbDetails.User+" WITH PASSWORD '"+password+"'")
	if err != nil {
		err1 := updateDBPwdSecrets(k8sClient, dbDetails.SecretKey, dbDetails.User, dbDetails.Password)
		if err1 != nil {
			fmt.Printf("failed to revert to the old password on password update failure %v", err1)
		}
		return fmt.Errorf("could not change TimescaleDB password: %w", err)
	}

	return nil
}

func updateDBPwdSecrets(k8sClient k8s.Client, secretKey, user, password string) error {
	// Update TimescaleDB credentials secret
	tsdb, err := common.IsTimescaleDBEnabled(root.HelmReleaseName, root.Namespace)
	if err != nil {
		return err
	}

	if tsdb {
		secretName := root.HelmReleaseName+"-credentials"
		tsdbSecret, err := k8sClient.KubeGetSecret(root.Namespace, secretName)
		if err != nil {
			return fmt.Errorf("could not get secret with name %s: %w", secretName, err)
		}

		tsdbSecret.Data[secretKey] = []byte(password)
		err = k8sClient.KubeUpdateSecret(root.Namespace, tsdbSecret)
		if err != nil {
			return fmt.Errorf("could not update secret with name %s: %w", secretName, err)
		}
	}

	// Update Grafana DB creds
	grafanaSecretName := root.HelmReleaseName+"-grafana-db"
	grafanaSecret, err := k8sClient.KubeGetSecret(root.Namespace, grafanaSecretName)
	if err != nil {
		return fmt.Errorf("could not get secret with name %s: %w", grafanaSecretName, err)
	}

	dbUser, ok := grafanaSecret.Data["GF_DATABASE_USER"]
	if ok && string(dbUser) == user {
		grafanaSecret.Data["GF_DATABASE_PASSWORD"] = []byte(password)
		err = k8sClient.KubeUpdateSecret(root.Namespace, grafanaSecret)
		if err != nil {
			return fmt.Errorf("could not update secret with name %s: %w", grafanaSecretName, err)
		}
	}

	// Update Promscale connection details
	promscaleSecretName, err :=  common.GetPromscaleSecretName(root.HelmReleaseName, root.Namespace)
	if err != nil {
		return err
	}
	promscaleSecret, err := k8sClient.KubeGetSecret(root.Namespace, promscaleSecretName)
	if err != nil {
		return fmt.Errorf("could not get secret with name %s: %w", promscaleSecretName, err)
	}

	if bytepass, exists := promscaleSecret.Data["PROMSCALE_DB_URI"]; exists && string(bytepass) != "" {
		dbDetails, err := pgconn.ParseDBURI(string(bytepass))
		if err != nil {
			return fmt.Errorf("failed to parse db-uri %v", err)
		}
		dbDetails.ConnConfig.Password = password
		updatedDBURI, err  := pgconn.UpdatePasswordInDBURI(string(bytepass), password)
		if err != nil {
			return fmt.Errorf("failed to update password in db-uri %v", err)
		}
		promscaleSecret.Data["PROMSCALE_DB_URI"] = []byte(updatedDBURI)
	} else {
		promscaleSecret.Data["PROMSCALE_DB_PASSWORD"] = []byte(password)
	}

	err = k8sClient.KubeUpdateSecret(root.Namespace, promscaleSecret)
	if err != nil {
		return fmt.Errorf("could not update secret with name %s: %w", promscaleSecretName, err)
	}

	return nil
}
