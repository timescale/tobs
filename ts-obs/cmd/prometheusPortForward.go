package cmd

import (
	"github.com/spf13/cobra"
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
	prometheusPortForwardCmd.Flags().IntP("port", "p", LISTEN_PORT_PROM, "Port to listen from")
}

func prometheusPortForward(cmd *cobra.Command, args []string) error {
	var err error

	var port int
	port, err = cmd.Flags().GetInt("port")
	if err != nil {
		return err
	}

	var name string
	name, err = cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	var namespace string
	namespace, err = cmd.Flags().GetString("namespace")
	if err != nil {
		return err
	}

	serviceName, err := KubeGetServiceName(namespace, map[string]string{"release": name, "app": "prometheus", "component": "server"})
	if err != nil {
		return err
	}

	err = KubePortForwardService(namespace, serviceName, port, FORWARD_PORT_PROM)
	if err != nil {
		return err
	}

	select {}

	return nil
}
