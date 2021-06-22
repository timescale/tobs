package grafana

import (
	"fmt"

	"github.com/spf13/cobra"
	root "github.com/timescale/tobs/cli/cmd"
	"github.com/timescale/tobs/cli/cmd/common"
)

// grafanaPortForwardCmd represents the grafana port-forward command
var grafanaPortForwardCmd = &cobra.Command{
	Use:   "port-forward",
	Short: "Port-forwards Grafana server to localhost",
	Args:  cobra.ExactArgs(0),
	RunE:  grafanaPortForward,
}

func init() {
	grafanaCmd.AddCommand(grafanaPortForwardCmd)
	grafanaPortForwardCmd.Flags().IntP("port", "p", common.LISTEN_PORT_GRAFANA, "Port to listen from")
}

func grafanaPortForward(cmd *cobra.Command, args []string) error {
	var err error

	var port int
	port, err = cmd.Flags().GetInt("port")
	if err != nil {
		return fmt.Errorf("could not port-forward Grafana: %w", err)
	}

	serviceName, err := kubeClient.KubeGetServiceName(root.Namespace, map[string]string{"app.kubernetes.io/instance": root.HelmReleaseName, "app.kubernetes.io/name": "grafana"})
	if err != nil {
		return fmt.Errorf("could not port-forward Grafana: %w", err)
	}

	_, err = kubeClient.KubePortForwardService(root.Namespace, serviceName, port, common.FORWARD_PORT_GRAFANA)
	if err != nil {
		return fmt.Errorf("could not port-forward Grafana: %w", err)
	}

	select {}
}
