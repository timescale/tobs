package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const FORWARD_PORT_TSDB = 5432
const FORWARD_PORT_GRAFANA = 3000
const FORWARD_PORT_PROM = 9090
const LISTEN_PORT_TSDB = 5432
const LISTEN_PORT_GRAFANA = 8080
const LISTEN_PORT_PROM = 9090

// portForwardCmd represents the port-forward command
var portForwardCmd = &cobra.Command{
	Use:   "port-forward",
	Short: "Port-forwards TimescaleDB, Grafana, and Prometheus to localhost",
	Args:  cobra.ExactArgs(0),
	RunE:  portForward,
}

func init() {
	rootCmd.AddCommand(portForwardCmd)
	portForwardCmd.Flags().IntP("timescaledb", "t", LISTEN_PORT_TSDB, "Port to listen from for TimescaleDB")
	portForwardCmd.Flags().IntP("grafana", "g", LISTEN_PORT_GRAFANA, "Port to listen from for Grafana")
	portForwardCmd.Flags().IntP("prometheus", "p", LISTEN_PORT_PROM, "Port to listen from for Prometheus")
}

func portForward(cmd *cobra.Command, args []string) error {
	var err error

	var timescaledb int
	timescaledb, err = cmd.Flags().GetInt("timescaledb")
	if err != nil {
		return fmt.Errorf("could not port-forward: %w", err)
	}

	var grafana int
	grafana, err = cmd.Flags().GetInt("grafana")
	if err != nil {
		return fmt.Errorf("could not port-forward: %w", err)
	}

	var prometheus int
	prometheus, err = cmd.Flags().GetInt("prometheus")
	if err != nil {
		return fmt.Errorf("could not port-forward: %w", err)
	}

	// Port-forward TimescaleDB
	podName, err := KubeGetPodName(namespace, map[string]string{"release": name, "role": "master"})
	if err != nil {
		return fmt.Errorf("could not port-forward: %w", err)
	}

	_, err = KubePortForwardPod(namespace, podName, timescaledb, FORWARD_PORT_TSDB)
	if err != nil {
		return fmt.Errorf("could not port-forward: %w", err)
	}

	// Port-forward Grafana
	serviceName, err := KubeGetServiceName(namespace, map[string]string{"app.kubernetes.io/instance": name, "app.kubernetes.io/name": "grafana"})
	if err != nil {
		return fmt.Errorf("could not port-forward: %w", err)
	}

	_, err = KubePortForwardService(namespace, serviceName, grafana, FORWARD_PORT_GRAFANA)
	if err != nil {
		return fmt.Errorf("could not port-forward: %w", err)
	}

	// Port-forward Prometheus
	serviceName, err = KubeGetServiceName(namespace, map[string]string{"release": name, "app": "prometheus", "component": "server"})
	if err != nil {
		return fmt.Errorf("could not port-forward: %w", err)
	}

	_, err = KubePortForwardService(namespace, serviceName, prometheus, FORWARD_PORT_PROM)
	if err != nil {
		return fmt.Errorf("could not port-forward: %w", err)
	}

	select {}

	return nil
}
