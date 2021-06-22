package volume

import (
	"github.com/spf13/cobra"
	"github.com/timescale/tobs/cli/cmd"
	"github.com/timescale/tobs/cli/pkg/k8s"
)

// volumeCmd represents the volume command
var volumeCmd = &cobra.Command{
	Use:   "volume",
	Short: "Subcommand for Volume operations",
}

var (
	kubeClient      *k8s.Client
)

func init() {
	cmd.RootCmd.AddCommand(volumeCmd)
	kubeClient, _ = k8s.NewClient()
}
