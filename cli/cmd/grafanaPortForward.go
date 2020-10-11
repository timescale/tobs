package cmd

import (
	"fmt"

	"github.com/timescale/tobs/cli/pkg/k8s"

	"github.com/spf13/cobra"
)

const LISTEN_PORT_GRAFANA = 8080
const FORWARD_PORT_GRAFANA = 3000

// grafanaPortForwardCmd represents the grafana port-forward command
var grafanaPortForwardCmd = &cobra.Command{
	Use:   "port-forward",
	Short: "Port-forwards Grafana server to localhost",
	Args:  cobra.ExactArgs(0),
	RunE:  grafanaPortForward,
}

func init() {
	grafanaCmd.AddCommand(grafanaPortForwardCmd)
	grafanaPortForwardCmd.Flags().IntP("port", "p", LISTEN_PORT_GRAFANA, "Port to listen from")
}

func grafanaPortForward(cmd *cobra.Command, args []string) error {
	var err error

	var port int
	port, err = cmd.Flags().GetInt("port")
	if err != nil {
		return fmt.Errorf("could not port-forward Grafana: %w", err)
	}

	serviceName, err := k8s.KubeGetServiceName(namespace, map[string]string{"app.kubernetes.io/instance": name, "app.kubernetes.io/name": "grafana"})
	if err != nil {
		return fmt.Errorf("could not port-forward Grafana: %w", err)
	}

	_, err = k8s.KubePortForwardService(namespace, serviceName, port, FORWARD_PORT_GRAFANA)
	if err != nil {
		return fmt.Errorf("could not port-forward Grafana: %w", err)
	}

	select {}
}
