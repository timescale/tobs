package grafana

import (
	"github.com/spf13/cobra"
	root "github.com/timescale/tobs/cli/cmd"
	"github.com/timescale/tobs/cli/pkg/k8s"
)

// grafanaCmd represents the grafana command
var grafanaCmd = &cobra.Command{
	Use:   "grafana",
	Short: "Subcommand for Grafana operations",
}

var (
	kubeClient      *k8s.Client
)

func init() {
	root.RootCmd.AddCommand(grafanaCmd)
	kubeClient, _ = k8s.NewClient()
}