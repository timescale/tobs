package timescaledb

import (
	"fmt"
	"github.com/spf13/cobra"
	root "github.com/timescale/tobs/cli/cmd"
	"github.com/timescale/tobs/cli/pkg/k8s"
)

var user string

// timescaledbCmd represents the timescaledb command
var timescaledbCmd = &cobra.Command{
	Use:   "timescaledb",
	Short: "Subcommand for TimescaleDB operations",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error

		err = root.RootCmd.PersistentPreRunE(cmd, args)
		if err != nil {
			return fmt.Errorf("could not read global flag: %w", err)
		}

		user, err = cmd.Flags().GetString("user")
		if err != nil {
			return fmt.Errorf("could not read flag: %w", err)
		}

		return nil
	},
}

var (
	kubeClient      *k8s.Client
)

func init() {
	root.RootCmd.AddCommand(timescaledbCmd)
	timescaledbCmd.PersistentFlags().StringP("user", "U", "", "database user name, if not provided super-user will be used from deployment")
    kubeClient, _ = k8s.NewClient()
}
