package superuser

import (
	"fmt"

	"github.com/spf13/cobra"
	root "github.com/timescale/tobs/cli/cmd"
	"github.com/timescale/tobs/cli/cmd/common"
	"github.com/timescale/tobs/cli/cmd/timescaledb"
	"github.com/timescale/tobs/cli/pkg/k8s"
)

// timescaledbConnectCmd represents the timescaledb superuser connect command
var timescaledbConnectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Connects to the TimescaleDB database using super-user",
	Args:  cobra.ExactArgs(0),
	RunE:  Connect,
}

func init() {
	superuserCmd.AddCommand(timescaledbConnectCmd)
	timescaledbConnectCmd.Flags().BoolP("master", "m", false, "directly execute session on master node")
}

func Connect(cmd *cobra.Command, args []string) error {
	master, err := cmd.Flags().GetBool("master")
	if err != nil {
		return fmt.Errorf("could not connect to TimescaleDB: %w", err)
	}

	dbDetails, err := common.GetSuperuserDBDetails(root.Namespace, root.HelmReleaseName)
	if err != nil {
		return fmt.Errorf("could not get DB secret key from helm release: %w", err)
	}

	k8sClient := k8s.NewClient()
	return timescaledb.PsqlConnect(k8sClient, dbDetails, master)
}
