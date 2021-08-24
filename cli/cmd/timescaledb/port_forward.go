package timescaledb

import (
	"fmt"

	"github.com/spf13/cobra"
	root "github.com/timescale/tobs/cli/cmd"
	"github.com/timescale/tobs/cli/cmd/common"
	"github.com/timescale/tobs/cli/pkg/k8s"
)

// timescaledbPortForwardCmd represents the timescaledb port-forward command
var timescaledbPortForwardCmd = &cobra.Command{
	Use:   "port-forward",
	Short: "Port-forwards TimescaleDB server to localhost",
	Args:  cobra.ExactArgs(0),
	RunE:  timescaledbPortForward,
}

func init() {
	Cmd.AddCommand(timescaledbPortForwardCmd)
	timescaledbPortForwardCmd.Flags().IntP("port", "p", common.LISTEN_PORT_TSDB, "Port to listen from")
}

func timescaledbPortForward(cmd *cobra.Command, args []string) error {
	var err error

	var port int
	port, err = cmd.Flags().GetInt("port")
	if err != nil {
		return fmt.Errorf("could not port-forward TimescaleDB: %w", err)
	}

	if err := PortForwardTimescaleDB(port); err != nil {
		return err
	}

	select {}
}

func PortForwardTimescaleDB(listenPort int) error {
	k8sClient := k8s.NewClient()
	podName, err := k8sClient.KubeGetPodName(root.Namespace, map[string]string{"release": root.HelmReleaseName, "role": "master"})
	if err != nil {
		return fmt.Errorf("could not port-forward TimescaleDB: %w", err)
	}

	_, err = k8sClient.KubePortForwardPod(root.Namespace, podName, listenPort, common.FORWARD_PORT_TSDB)
	if err != nil {
		return fmt.Errorf("could not port-forward TimescaleDB: %w", err)
	}

	return nil
}