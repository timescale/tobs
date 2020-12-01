package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/timescale/tobs/cli/pkg/k8s"
)

const LISTEN_PORT_PROMSCALE = 9201
const FORWARD_PORT_PROMSCALE = 9201

// promscalePortForwardCmd represents the PromScale port-forward command
var promscalePortForwardCmd = &cobra.Command{
	Use:   "port-forward",
	Short: "Port-forwards Promscale to localhost",
	Args:  cobra.ExactArgs(0),
	RunE:  promscalePortForward,
}

func init() {
	promscaleCmd.AddCommand(promscalePortForwardCmd)
	promscalePortForwardCmd.Flags().IntP("port", "p", LISTEN_PORT_PROMSCALE, "Port to listen for the promscale")
}

func portForwardPromscale(listenPort int) error {
	serviceNamePromscale, err := k8s.KubeGetServiceName(namespace, map[string]string{"release": name, "app": name + "-promscale"})
	if err != nil {
		return fmt.Errorf("could not port-forward Promscale: %w", err)
	}

	_, err = k8s.KubePortForwardService(namespace, serviceNamePromscale, listenPort, FORWARD_PORT_PROMSCALE)
	if err != nil {
		return fmt.Errorf("could not port-forward Promscale: %w", err)
	}

	return nil
}

func promscalePortForward(cmd *cobra.Command, args []string) error {
	var err error

	var port int
	port, err = cmd.Flags().GetInt("port")
	if err != nil {
		return fmt.Errorf("could not port-forward Promscale: %w", err)
	}

	if err := portForwardPromscale(port); err != nil {
		return err
	}

	select {}
}
