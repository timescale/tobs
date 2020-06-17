package cmd

import (
    "errors"

	"github.com/spf13/cobra"
)

// portForwardCmd represents the port-forward command
var portForwardCmd = &cobra.Command{
	Use:   "port-forward",
	Short: "Port-forwards TimescaleDB, Grafana, and Prometheus to localhost",
	RunE:  portForward,
}

func init() {
	rootCmd.AddCommand(portForwardCmd)
    portForwardCmd.Flags().IntP("timescaledb", "t", 5432, "Port to forward TimescaleDB to")
    portForwardCmd.Flags().IntP("grafana", "g", 8080, "Port to forward Grafana to")
    portForwardCmd.Flags().IntP("prometheus", "p", 9090, "Port to forward Prometheus to")
}

func portForward(cmd *cobra.Command, args []string) error {
    var err error

    if len(args) != 0 {
        return errors.New("\"ts-obs port-forward\" requires 0 arguments")
    }

    var timescaledb int
    timescaledb, err = cmd.Flags().GetInt("timescaledb")
    if err != nil {
        return err
    }

    var grafana int
    grafana, err = cmd.Flags().GetInt("grafana")
    if err != nil {
        return err
    }

    var prometheus int
    prometheus, err = cmd.Flags().GetInt("prometheus")
    if err != nil {
        return err
    }

    // Port-forward TimescaleDB
    err = kubePortForwardPod("ts-obs-timescaledb-0", timescaledb, 5432)
    if err != nil {
        return err
    }

    // Port-forward Grafana
    err = kubePortForwardService("ts-obs-grafana", grafana, 3000)
    if err != nil {
        return err
    }

    // Port-forward Prometheus
    serviceName, err := kubeGetServiceName(map[string]string{"app" : "prometheus", "component" : "server"})
    if err != nil {
        return err
    }

    err = kubePortForwardService(serviceName, prometheus, 9090)
    if err != nil {
        return err
    }

    select {}

    return nil
}
