package promscale

import (
	"fmt"

	"github.com/spf13/cobra"
	root "github.com/timescale/tobs/cli/cmd"
	"github.com/timescale/tobs/cli/cmd/common"
)

// promscalePortForwardCmd represents the PromScale port-forward command
var promscalePortForwardCmd = &cobra.Command{
	Use:   "port-forward",
	Short: "Port-forwards Promscale to localhost",
	Args:  cobra.ExactArgs(0),
	RunE:  promscalePortForward,
}

func init() {
	promscaleCmd.AddCommand(promscalePortForwardCmd)
	promscalePortForwardCmd.Flags().IntP("port", "p", common.LISTEN_PORT_PROMSCALE, "Port to listen for the promscale")
}

func PortForwardPromscale(listenPort int) error {
	serviceNamePromscale, err := kubeClient.KubeGetServiceName(root.Namespace, map[string]string{"release": root.HelmReleaseName, "app": root.HelmReleaseName + "-promscale"})
	if err != nil {
		return fmt.Errorf("could not port-forward Promscale: %w", err)
	}

	_, err = kubeClient.KubePortForwardService(root.Namespace, serviceNamePromscale, listenPort, common.FORWARD_PORT_PROMSCALE)
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

	if err := PortForwardPromscale(port); err != nil {
		return err
	}

	select {}
}
