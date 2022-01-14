package jaeger

import (
	"fmt"
	"github.com/spf13/cobra"
	root "github.com/timescale/tobs/cli/cmd"
	"github.com/timescale/tobs/cli/cmd/common"
	"github.com/timescale/tobs/cli/pkg/k8s"
)

// jaegerPortForwardCmd represents the jaeger port-forward command
var jaegerPortForwardCmd = &cobra.Command{
	Use:   "port-forward",
	Short: "Port-forwards Jaeger server to localhost",
	Args:  cobra.ExactArgs(0),
	RunE:  jaegerPortForward,
}

func init() {
	jaegerCmd.AddCommand(jaegerPortForwardCmd)
	jaegerPortForwardCmd.Flags().IntP("port", "p", common.LISTEN_PORT_JAEGER, "Port to listen from")
}

func jaegerPortForward(cmd *cobra.Command, args []string) error {
	var err error

	var port int
	port, err = cmd.Flags().GetInt("port")
	if err != nil {
		return fmt.Errorf("could not port-forward Jaeger: %w", err)
	}

	if err := PortForwardJaeger(port); err != nil {
		return err
	}

	select {}
}

func PortForwardJaeger(listenPort int) error {
	k8sClient := k8s.NewClient()
	serviceName, err := k8sClient.KubeGetServiceName(root.Namespace, map[string]string{"release": root.HelmReleaseName, "app": "jaeger"})
	if err != nil {
		return fmt.Errorf("could not port-forward Jaeger: %w", err)
	}

	_, err = k8sClient.KubePortForwardService(root.Namespace, serviceName, listenPort, common.FORWARD_PORT_JAEGER)
	if err != nil {
		return fmt.Errorf("could not port-forward Jaeger: %w", err)
	}

	return nil
}
