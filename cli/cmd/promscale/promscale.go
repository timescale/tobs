package promscale

import (
	"github.com/spf13/cobra"
	"github.com/timescale/tobs/cli/cmd"
	"github.com/timescale/tobs/cli/pkg/k8s"
)

// promscaleCmd represents the promscale command
var promscaleCmd = &cobra.Command{
	Use:   "promscale",
	Short: "Subcommand for Promscale operations",
}

var (
	kubeClient      *k8s.Client
)

func init() {
	cmd.RootCmd.AddCommand(promscaleCmd)
	kubeClient, _ = k8s.NewClient()
}