package cmd

import (
	"fmt"

	"github.com/timescale/tobs/cli/pkg/k8s"

	"github.com/spf13/cobra"
)

const LISTEN_PORT_PROM = 9090
const FORWARD_PORT_PROM = 9090

// prometheusPortForwardCmd represents the prometheus port-forward command
var prometheusPortForwardCmd = &cobra.Command{
	Use:   "port-forward",
	Short: "Port-forwards Prometheus server to localhost",
	Args:  cobra.ExactArgs(0),
	RunE:  prometheusPortForward,
}

func init() {
	prometheusCmd.AddCommand(prometheusPortForwardCmd)
	prometheusPortForwardCmd.Flags().IntP("port", "p", LISTEN_PORT_PROM, "Port to listen from")
}

func prometheusPortForward(cmd *cobra.Command, args []string) error {
	var err error

	var port int
	port, err = cmd.Flags().GetInt("port")
	if err != nil {
		return fmt.Errorf("could not port-forward Prometheus: %w", err)
	}

	serviceName, err := k8s.KubeGetServiceName(namespace, map[string]string{"release": name, "app": "prometheus", "component": "server"})
	if err != nil {
		return fmt.Errorf("could not port-forward Prometheus: %w", err)
	}

	_, err = k8s.KubePortForwardService(namespace, serviceName, port, FORWARD_PORT_PROM)
	if err != nil {
		return fmt.Errorf("could not port-forward Prometheus: %w", err)
	}

	select {}
}
