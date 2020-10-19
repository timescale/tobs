package cmd

import (
	"fmt"

	"github.com/timescale/tobs/cli/pkg/k8s"

	"github.com/spf13/cobra"
)

const LISTEN_PORT_TSDB = 5432
const FORWARD_PORT_TSDB = 5432

// timescaledbPortForwardCmd represents the timescaledb port-forward command
var timescaledbPortForwardCmd = &cobra.Command{
	Use:   "port-forward",
	Short: "Port-forwards TimescaleDB server to localhost",
	Args:  cobra.ExactArgs(0),
	RunE:  timescaledbPortForward,
}

func init() {
	timescaledbCmd.AddCommand(timescaledbPortForwardCmd)
	timescaledbPortForwardCmd.Flags().IntP("port", "p", LISTEN_PORT_TSDB, "Port to listen from")
}

func timescaledbPortForward(cmd *cobra.Command, args []string) error {
	var err error

	var port int
	port, err = cmd.Flags().GetInt("port")
	if err != nil {
		return fmt.Errorf("could not port-forward TimescaleDB: %w", err)
	}

	podName, err := k8s.KubeGetPodName(namespace, map[string]string{"release": name, "role": "master"})
	if err != nil {
		return fmt.Errorf("could not port-forward TimescaleDB: %w", err)
	}

	_, err = k8s.KubePortForwardPod(namespace, podName, port, FORWARD_PORT_TSDB)
	if err != nil {
		return fmt.Errorf("could not port-forward TimescaleDB: %w", err)
	}

	select {}
}
