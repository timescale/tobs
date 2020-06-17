package cmd

import (
    "errors"

	"github.com/spf13/cobra"
)

// timescaledbPortForwardCmd represents the timescaledb port-forward command
var timescaledbPortForwardCmd = &cobra.Command{
	Use:   "port-forward",
    Short: "Port-forwards TimescaleDB server to localhost",
	RunE:  timescaledbPortForward,
}

func init() {
	timescaledbCmd.AddCommand(timescaledbPortForwardCmd)
    timescaledbPortForwardCmd.Flags().IntP("port", "p", 5432, "Port to forward to")
}

func timescaledbPortForward(cmd *cobra.Command, args []string) error {
    var err error

    if len(args) != 0 {
        return errors.New("\"ts-obs port-forward\" requires 0 arguments")
    }

    var port int
    port, err = cmd.Flags().GetInt("port")
    if err != nil {
        return err
    }

    err = kubePortForwardPod("ts-obs-timescaledb-0", port, 5432)
    if err != nil {
        return err
    }

    select {}

    return nil
}
