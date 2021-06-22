package promlens

import (
	"github.com/spf13/cobra"
	"github.com/timescale/tobs/cli/cmd"
	"github.com/timescale/tobs/cli/pkg/k8s"
)

// promlensCmd represents the promlens command
var promlensCmd = &cobra.Command{
	Use:   "promlens",
	Short: "Subcommand for Promlens operations",
}

var (
	kubeClient      *k8s.Client
)

func init() {
	cmd.RootCmd.AddCommand(promlensCmd)
	kubeClient, _ = k8s.NewClient()
}
