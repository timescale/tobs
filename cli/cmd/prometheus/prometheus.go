package prometheus

import (
	"github.com/spf13/cobra"
	"github.com/timescale/tobs/cli/cmd"
	"github.com/timescale/tobs/cli/pkg/k8s"
)

// prometheusCmd represents the prometheus command
var prometheusCmd = &cobra.Command{
	Use:   "prometheus",
	Short: "Subcommand for Prometheus operations",
}

var (
	kubeClient      *k8s.Client
)

func init() {
	cmd.RootCmd.AddCommand(prometheusCmd)
	kubeClient, _ = k8s.NewClient()
}
