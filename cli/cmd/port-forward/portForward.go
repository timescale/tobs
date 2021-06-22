package port_forward

import (
	"fmt"
	"github.com/timescale/tobs/cli/pkg/k8s"

	"github.com/spf13/cobra"
	root "github.com/timescale/tobs/cli/cmd"
	"github.com/timescale/tobs/cli/cmd/common"
	"github.com/timescale/tobs/cli/cmd/promlens"
	"github.com/timescale/tobs/cli/cmd/promscale"
)

// portForwardCmd represents the port-forward command
var portForwardCmd = &cobra.Command{
	Use:   "port-forward",
	Short: "Port-forwards TimescaleDB, Promscale, Promlens, Grafana, and Prometheus to localhost",
	Args:  cobra.ExactArgs(0),
	RunE:  portForward,
}

var (
	kubeClient      *k8s.Client
)

func init() {
	root.RootCmd.AddCommand(portForwardCmd)
	portForwardCmd.Flags().IntP("timescaledb", "t", common.LISTEN_PORT_TSDB, "Port to listen from for TimescaleDB")
	portForwardCmd.Flags().IntP("grafana", "g", common.LISTEN_PORT_GRAFANA, "Port to listen from for Grafana")
	portForwardCmd.Flags().IntP("prometheus", "p", common.LISTEN_PORT_PROM, "Port to listen from for Prometheus")
	portForwardCmd.Flags().IntP("promscale", "c", common.LISTEN_PORT_PROMSCALE, "Port to listen from for the Promscale")
	portForwardCmd.Flags().IntP("promlens", "l", common.LISTEN_PORT_PROMLENS, "Port to listen from for PromLens")
	kubeClient, _ = k8s.NewClient()
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

	var promscalePort int
	promscalePort, err = cmd.Flags().GetInt("promscale")
	if err != nil {
		return fmt.Errorf("could not port-forward: %w", err)
	}

	var promlensPort int
	promlensPort, err = cmd.Flags().GetInt("promlens")
	if err != nil {
		return fmt.Errorf("could not port-forward: %w", err)
	}

	// Port-forward TimescaleDB
	// if db-uri exists skip the port-forwarding as it isn't the db within the cluster
	uri, err := kubeClient.GetTimescaleDBURI(root.Namespace, root.HelmReleaseName)
	if err != nil {
		return err
	}
	if uri == "" {
		podName, err := kubeClient.KubeGetPodName(root.Namespace, map[string]string{"release": root.HelmReleaseName, "role": "master"})
		if err != nil {
			return fmt.Errorf("could not port-forward: %w", err)
		}

		_, err = kubeClient.KubePortForwardPod(root.Namespace, podName, timescaledb, common.FORWARD_PORT_TSDB)
		if err != nil {
			return fmt.Errorf("could not port-forward: %w", err)
		}
	}

	// Port-forward Grafana
	serviceName, err := kubeClient.KubeGetServiceName(root.Namespace, map[string]string{"app.kubernetes.io/instance": root.HelmReleaseName, "app.kubernetes.io/name": "grafana"})
	if err != nil {
		return fmt.Errorf("could not port-forward: %w", err)
	}

	_, err = kubeClient.KubePortForwardService(root.Namespace, serviceName, grafana, common.FORWARD_PORT_GRAFANA)
	if err != nil {
		return fmt.Errorf("could not port-forward: %w", err)
	}

	// Port-forward Prometheus
	serviceName, err = kubeClient.KubeGetServiceName(root.Namespace, map[string]string{"release": root.HelmReleaseName, "app": "kube-prometheus-stack-prometheus"})
	if err != nil {
		return fmt.Errorf("could not port-forward: %w", err)
	}

	_, err = kubeClient.KubePortForwardService(root.Namespace, serviceName, prometheus, common.FORWARD_PORT_PROM)
	if err != nil {
		return fmt.Errorf("could not port-forward: %w", err)
	}

	if err := promlens.PortForwardPromlens(promlensPort); err != nil {
		return err
	}

	if err := promscale.PortForwardPromscale(promscalePort); err != nil {
		return err
	}
	select {}
}
