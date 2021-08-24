package port_forward

import (
	"fmt"

	"github.com/spf13/cobra"
	root "github.com/timescale/tobs/cli/cmd"
	"github.com/timescale/tobs/cli/cmd/common"
	"github.com/timescale/tobs/cli/cmd/grafana"
	"github.com/timescale/tobs/cli/cmd/jaeger"
	"github.com/timescale/tobs/cli/cmd/prometheus"
	"github.com/timescale/tobs/cli/cmd/promlens"
	"github.com/timescale/tobs/cli/cmd/promscale"
	"github.com/timescale/tobs/cli/cmd/timescaledb"
	"github.com/timescale/tobs/cli/pkg/helm"
	"github.com/timescale/tobs/cli/pkg/k8s"
	"github.com/timescale/tobs/cli/pkg/utils"
)

// portForwardCmd represents the port-forward command
var portForwardCmd = &cobra.Command{
	Use:   "port-forward",
	Short: "Port-forwards TimescaleDB, Promscale, Promlens, Grafana, Prometheus, and Jaeger to localhost",
	Args:  cobra.ExactArgs(0),
	RunE:  portForward,
}

func init() {
	root.RootCmd.AddCommand(portForwardCmd)
	portForwardCmd.Flags().IntP("timescaledb", "t", common.LISTEN_PORT_TSDB, "Port to listen from for TimescaleDB")
	portForwardCmd.Flags().IntP("grafana", "g", common.LISTEN_PORT_GRAFANA, "Port to listen from for Grafana")
	portForwardCmd.Flags().IntP("prometheus", "p", common.LISTEN_PORT_PROM, "Port to listen from for Prometheus")
	portForwardCmd.Flags().IntP("promscale", "c", common.LISTEN_PORT_PROMSCALE, "Port to listen from for the Promscale")
	portForwardCmd.Flags().IntP("promlens", "l", common.LISTEN_PORT_PROMLENS, "Port to listen from for PromLens")
	portForwardCmd.Flags().IntP("jaeger", "j", common.LISTEN_PORT_JAEGER, "Port to listen from for Jaeger")
}

func portForward(cmd *cobra.Command, args []string) error {
	var err error

	timescaledbPort, err := cmd.Flags().GetInt("timescaledb")
	if err != nil {
		return fmt.Errorf("could not port-forward: %w", err)
	}

	grafanaPort, err := cmd.Flags().GetInt("grafana")
	if err != nil {
		return fmt.Errorf("could not port-forward: %w", err)
	}

	prometheusPort, err := cmd.Flags().GetInt("prometheus")
	if err != nil {
		return fmt.Errorf("could not port-forward: %w", err)
	}

	promscalePort, err := cmd.Flags().GetInt("promscale")
	if err != nil {
		return fmt.Errorf("could not port-forward: %w", err)
	}

	promlensPort, err := cmd.Flags().GetInt("promlens")
	if err != nil {
		return fmt.Errorf("could not port-forward: %w", err)
	}

	jaegerPort, err := cmd.Flags().GetInt("jaeger")
	if err != nil {
		return fmt.Errorf("could not port-forward: %w", err)
	}

	// Port-forward TimescaleDB
	// if db-uri exists skip the port-forwarding as it isn't the db within the cluster
	k8sClient := k8s.NewClient()
	uri, err := utils.GetTimescaleDBURI(k8sClient, root.Namespace, root.HelmReleaseName)
	if err != nil {
		return err
	}

	if uri == "" {
		if err := timescaledb.PortForwardTimescaleDB(timescaledbPort); err != nil {
			return err
		}
	}

	// Port-forward Grafana
	if err := grafana.PortForwardGrafana(grafanaPort); err != nil {
		return err
	}

	// Port-forward Prometheus
	if err := prometheus.PortForwardPrometheus(prometheusPort); err != nil {
		return err
	}

	// Port-forward Promlens
	if err := promlens.PortForwardPromlens(promlensPort); err != nil {
		return err
	}

	// Port-forward Promscale
	if err := promscale.PortForwardPromscale(promscalePort); err != nil {
		return err
	}

	// As Jaeger isn't part of default install, check whether
	// Jaeger is enabled
	helmClient := helm.NewClient(root.Namespace)
	e, err := helmClient.ExportValuesFieldFromRelease(root.HelmReleaseName, []string{"opentelemetryOperator", "enabled"})
	if err != nil {
		return err
	}
	enableOtel, ok := e.(bool)
	if !ok {
		return fmt.Errorf("timescaledb-single.enabled was not a bool")
	}
	// Port-forward Jaeger
	if enableOtel {
		if err := jaeger.PortForwardJaeger(jaegerPort); err != nil {
			return err
		}
	}
	select {}
}
