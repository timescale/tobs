package prometheus

import (
	"fmt"

	"github.com/spf13/cobra"
	root "github.com/timescale/tobs/cli/cmd"
	"github.com/timescale/tobs/cli/cmd/common"
	"github.com/timescale/tobs/cli/pkg/k8s"
)

// prometheusPortForwardCmd represents the prometheus port-forward command
var prometheusPortForwardCmd = &cobra.Command{
	Use:   "port-forward",
	Short: "Port-forwards Prometheus server to localhost",
	Args:  cobra.ExactArgs(0),
	RunE:  prometheusPortForward,
}

func init() {
	prometheusCmd.AddCommand(prometheusPortForwardCmd)
	prometheusPortForwardCmd.Flags().IntP("port", "p", common.LISTEN_PORT_PROM, "Port to listen from")
}

func prometheusPortForward(cmd *cobra.Command, args []string) error {
	var err error

	var port int
	port, err = cmd.Flags().GetInt("port")
	if err != nil {
		return fmt.Errorf("could not port-forward Prometheus: %w", err)
	}

	serviceName, err := k8s.KubeGetServiceName(root.Namespace, map[string]string{"release": root.HelmReleaseName, "app": "kube-prometheus-stack-prometheus"})
	if err != nil {
		return fmt.Errorf("could not port-forward Prometheus: %w", err)
	}

	_, err = k8s.KubePortForwardService(root.Namespace, serviceName, port, common.FORWARD_PORT_PROM)
	if err != nil {
		return fmt.Errorf("could not port-forward Prometheus: %w", err)
	}

	select {}
}
