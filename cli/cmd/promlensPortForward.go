package cmd

import (
	"fmt"

	"github.com/timescale/tobs/cli/pkg/k8s"

	"github.com/spf13/cobra"
)

const LISTEN_PORT_PROMLENS = 8081
const FORWARD_PORT_PROMLENS = 8080
const LISTEN_PORT_CONNECTOR = 9201
const FORWARD_PORT_CONNECTOR = 9201

// promlensPortForwardCmd represents the PromLens port-forward command
var promlensPortForwardCmd = &cobra.Command{
	Use:   "port-forward",
	Short: "Port-forwards PromLens UI and connector to localhost",
	Args:  cobra.ExactArgs(0),
	RunE:  promlensPortForward,
}

func init() {
	promlensCmd.AddCommand(promlensPortForwardCmd)
	promlensPortForwardCmd.Flags().IntP("port", "p", LISTEN_PORT_PROMLENS, "Port to listen from for promlens")
	promlensPortForwardCmd.Flags().IntP("port-connector", "c", LISTEN_PORT_CONNECTOR, "Port to listen for the connector")
}

func portForwardPromlens(listenPort int) error {
	serviceNamePromlens, err := k8s.KubeGetServiceName(namespace, map[string]string{"release": name, "component": "promlens"})
	if err != nil {
		return fmt.Errorf("could not port-forward PromLens: %w", err)
	}

	_, err = k8s.KubePortForwardService(namespace, serviceNamePromlens, listenPort, FORWARD_PORT_PROMLENS)
	if err != nil {
		return fmt.Errorf("could not port-forward PromLens: %w", err)
	}

	return nil
}

func portForwardConnector(listenPort int) error {
	serviceNameConnector, err := k8s.KubeGetServiceName(namespace, map[string]string{"release": name, "app": name + "-promscale"})
	if err != nil {
		return fmt.Errorf("could not port-forward PromLens: %w", err)
	}

	_, err = k8s.KubePortForwardService(namespace, serviceNameConnector, listenPort, FORWARD_PORT_CONNECTOR)
	if err != nil {
		return fmt.Errorf("could not port-forward PromLens: %w", err)
	}

	return nil
}

func promlensPortForward(cmd *cobra.Command, args []string) error {
	var err error

	var port int
	port, err = cmd.Flags().GetInt("port")
	if err != nil {
		return fmt.Errorf("could not port-forward PromLens: %w", err)
	}

	var portConnector int
	portConnector, err = cmd.Flags().GetInt("port-connector")
	if err != nil {
		return fmt.Errorf("could not port-forward PromLens: %w", err)
	}

	if err := portForwardPromlens(port); err != nil {
		return err
	}

	if err := portForwardConnector(portConnector); err != nil {
		return err
	}

	select {}

}
