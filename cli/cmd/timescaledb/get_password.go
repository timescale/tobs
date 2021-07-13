package timescaledb

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	root "github.com/timescale/tobs/cli/cmd"
	"github.com/timescale/tobs/cli/cmd/common"
	"github.com/timescale/tobs/cli/pkg/k8s"
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
}

func timescaledbGetPassword(cmd *cobra.Command, args []string) error {
	k8sClient := k8s.NewClient()
	secret, err := k8sClient.KubeGetSecret(root.Namespace, root.HelmReleaseName+"-credentials")
	if err != nil {
		return fmt.Errorf("could not get TimescaleDB password: %w", err)
	}

	var pass string

	d, err := common.FormDBDetails(user, "", root.Namespace, root.HelmReleaseName)
	if err != nil {
		return err
	}

	if bytepass, exists := secret.Data[d.SecretKey]; exists {
		pass = string(bytepass)
	} else {
		return fmt.Errorf("could not get TimescaleDB password: %w", errors.New("user not found"))
	}
	fmt.Println(pass)

	return nil
}
