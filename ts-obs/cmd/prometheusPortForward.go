package cmd

import (
    "errors"

	"github.com/spf13/cobra"
)

// prometheusPortForwardCmd represents the prometheus port-forward command
var prometheusPortForwardCmd = &cobra.Command{
	Use:   "port-forward",
    Short: "Port-forwards Prometheus server to localhost",
	RunE:  prometheusPortForward,
}

func init() {
	prometheusCmd.AddCommand(prometheusPortForwardCmd)
    prometheusPortForwardCmd.Flags().IntP("port", "p", 9090, "Port to forward to")
}

func prometheusPortForward(cmd *cobra.Command, args []string) error {
    var err error

    if len(args) != 0 {
        return errors.New("\"ts-obs prometheus port-forward\" requires 0 arguments")
    }

    var port int
    port, err = cmd.Flags().GetInt("port")
    if err != nil {
        return err
    }

    serviceName, err := kubeGetServiceName(map[string]string{"app" : "prometheus", "component" : "server"})
    if err != nil {
        return err
    }

    err = kubePortForwardService(serviceName, port, 9090)
    if err != nil {
        return err
    }

    select {}

    return nil
}
