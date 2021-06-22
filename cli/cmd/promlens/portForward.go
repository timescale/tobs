package promlens

import (
	"fmt"

	"github.com/spf13/cobra"
	root "github.com/timescale/tobs/cli/cmd"
	"github.com/timescale/tobs/cli/cmd/common"
)

// promlensPortForwardCmd represents the PromLens port-forward command
var promlensPortForwardCmd = &cobra.Command{
	Use:   "port-forward",
	Short: "Port-forwards PromLens UI to localhost",
	Args:  cobra.ExactArgs(0),
	RunE:  promlensPortForward,
}

func init() {
	promlensCmd.AddCommand(promlensPortForwardCmd)
	promlensPortForwardCmd.Flags().IntP("port", "p", common.LISTEN_PORT_PROMLENS, "Port to listen from for promlens")
}

func PortForwardPromlens(listenPort int) error {
	serviceNamePromlens, err := kubeClient.KubeGetServiceName(root.Namespace, map[string]string{"release": root.HelmReleaseName, "component": "promlens"})
	if err != nil {
		return fmt.Errorf("could not port-forward PromLens: %w", err)
	}

	_, err = kubeClient.KubePortForwardService(root.Namespace, serviceNamePromlens, listenPort, common.FORWARD_PORT_PROMLENS)
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

	if err := PortForwardPromlens(port); err != nil {
		return err
	}

	select {}

}
