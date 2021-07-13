package metrics

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	root "github.com/timescale/tobs/cli/cmd"
	"github.com/timescale/tobs/cli/cmd/common"
)

// retentionGetCmd represents the metrics retention get command
var retentionGetCmd = &cobra.Command{
	Use:   "get <metric>",
	Short: "Gets data retention period in days for a specific metric",
	Args:  cobra.ExactArgs(1),
	RunE:  retentionGet,
}

func init() {
	retentionCmd.AddCommand(retentionGetCmd)
}

func retentionGet(cmd *cobra.Command, args []string) error {
	var err error

	metric := args[0]

	d, err := common.FormDBDetails(user, dbname, root.Namespace, root.HelmReleaseName)
	if err != nil {
		return err
	}

	pool, err := d.OpenConnectionToDB()
	if err != nil {
		return fmt.Errorf("could not get retention period for %v: %w", metric, err)
	}
	defer pool.Close()

	fmt.Printf("Getting retention period for %v\n", metric)
	var days int
	err = pool.QueryRow(context.Background(), "SELECT EXTRACT(day FROM _prom_catalog.get_metric_retention_period($1))", metric).Scan(&days)
	if err != nil {
		return fmt.Errorf("could not get retention period for %v: %w", metric, err)
	}

	fmt.Println(days, "days")

	return nil
}
