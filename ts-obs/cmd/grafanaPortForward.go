package cmd

import (
    "errors"

	"github.com/spf13/cobra"
)

// grafanaPortForwardCmd represents the grafana port-forward command
var grafanaPortForwardCmd = &cobra.Command{
	Use:   "port-forward",
    Short: "Port-forwards Grafana server to localhost",
	RunE:  grafanaPortForward,
}

func init() {
	grafanaCmd.AddCommand(grafanaPortForwardCmd)
    grafanaPortForwardCmd.Flags().IntP("port", "p", 8080, "Port to forward to")
}

func grafanaPortForward(cmd *cobra.Command, args []string) error {
    var err error

    if len(args) != 0 {
        return errors.New("\"ts-obs grafana port-forward\" requires 0 arguments")
    }

    var port int
    port, err = cmd.Flags().GetInt("port")
    if err != nil {
        return err
    }

    err = kubePortForwardService("ts-obs-grafana", port, 3000)
    if err != nil {
        return err
    }
   
    select {}

    return nil
}
