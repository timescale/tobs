package timescaledb

import (
	"fmt"

	"github.com/spf13/cobra"
	root "github.com/timescale/tobs/cli/cmd"
	"github.com/timescale/tobs/cli/cmd/common"
)

// timescaledbPortForwardCmd represents the timescaledb port-forward command
var timescaledbPortForwardCmd = &cobra.Command{
	Use:   "port-forward",
	Short: "Port-forwards TimescaleDB server to localhost",
	Args:  cobra.ExactArgs(0),
	RunE:  timescaledbPortForward,
}

func init() {
	timescaledbCmd.AddCommand(timescaledbPortForwardCmd)
	timescaledbPortForwardCmd.Flags().IntP("port", "p", common.LISTEN_PORT_TSDB, "Port to listen from")
}

func timescaledbPortForward(cmd *cobra.Command, args []string) error {
	var err error

	var port int
	port, err = cmd.Flags().GetInt("port")
	if err != nil {
		return fmt.Errorf("could not port-forward TimescaleDB: %w", err)
	}

	podName, err := kubeClient.KubeGetPodName(root.Namespace, map[string]string{"release": root.HelmReleaseName, "role": "master"})
	if err != nil {
		return fmt.Errorf("could not port-forward TimescaleDB: %w", err)
	}

	_, err = kubeClient.KubePortForwardPod(root.Namespace, podName, port, common.FORWARD_PORT_TSDB)
	if err != nil {
		return fmt.Errorf("could not port-forward TimescaleDB: %w", err)
	}

	select {}
}
