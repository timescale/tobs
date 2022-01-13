package superuser

import (
	"fmt"
	"github.com/spf13/cobra"
	root "github.com/timescale/tobs/cli/cmd"
	"github.com/timescale/tobs/cli/cmd/common"
)

// timescaledbGetPasswordCmd represents the timescaledb get-password command
var timescaledbGetPasswordCmd = &cobra.Command{
	Use:   "get-password",
	Short: "Gets the TimescaleDB password for a specific user",
	Args:  cobra.ExactArgs(0),
	RunE:  timescaledbGetPassword,
}

func init() {
	superuserCmd.AddCommand(timescaledbGetPasswordCmd)
}

func timescaledbGetPassword(cmd *cobra.Command, args []string) error {
	d, err := common.GetSuperuserDBDetails(root.Namespace, root.HelmReleaseName)
	if err != nil {
		return err
	}

	fmt.Println(d.Password)
	return nil
}
