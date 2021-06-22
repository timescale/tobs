package helm

import (
	"github.com/spf13/cobra"
	root "github.com/timescale/tobs/cli/cmd"
	"github.com/timescale/tobs/cli/pkg/helm"
	"github.com/timescale/tobs/cli/pkg/k8s"
)

// helmCmd represents the helm command
var helmCmd = &cobra.Command{
	Use:   "helm",
	Short: "Subcommand for Helm operations",
}

var (
	kubeClient      *k8s.Client
	helmClient      *helm.ClientInfo
)

func init() {
	root.RootCmd.AddCommand(helmCmd)
	kubeClient, _ = k8s.NewClient()
}
