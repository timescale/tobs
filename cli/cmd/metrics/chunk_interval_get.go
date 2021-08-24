package metrics

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	root "github.com/timescale/tobs/cli/cmd"
	"github.com/timescale/tobs/cli/cmd/common"
)

// chunkIntervalGetCmd represents the chunk-interval get command
var chunkIntervalGetCmd = &cobra.Command{
	Use:   "get <metric>",
	Short: "Gets chunk interval for a specific metric",
	Args:  cobra.ExactArgs(1),
	RunE:  chunkIntervalGet,
}

func init() {
	chunkIntervalCmd.AddCommand(chunkIntervalGetCmd)
}

func chunkIntervalGet(cmd *cobra.Command, args []string) error {
	var err error

	metric := args[0]

	d, err := common.GetSuperuserDBDetails(root.Namespace, root.HelmReleaseName)
	if err != nil {
		return err
	}

	pool, err := d.OpenConnectionToDB()
	if err != nil {
		return fmt.Errorf("could not get chunk interval for %v: %w", metric, err)
	}
	defer pool.Close()

	fmt.Printf("Getting chunk interval of %v\n", metric)
	var microsecs int64
	err = pool.QueryRow(context.Background(),
		`SELECT d.interval_length
	 FROM _timescaledb_catalog.hypertable h
	 INNER JOIN LATERAL
	 (SELECT dim.interval_length FROM _timescaledb_catalog.dimension dim WHERE dim.hypertable_id = h.id ORDER BY dim.id LIMIT 1) d
	    ON (true)
	 WHERE table_name = $1`,
		metric).Scan(&microsecs)
	if err != nil {
		return fmt.Errorf("could not get chunk interval for %v: %w", metric, err)
	}

	interval := time.Duration(microsecs) * time.Microsecond
	fmt.Println(interval)

	return nil
}
