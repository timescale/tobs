package metrics

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	root "github.com/timescale/tobs/cli/cmd"
	"github.com/timescale/tobs/cli/cmd/common"
)

// chunkIntervalSetCmd represents the chunk-interval set command
var chunkIntervalSetCmd = &cobra.Command{
	Use:   "set <metric> <duration>",
	Short: "Sets chunk interval for a specific metric",
	Args:  cobra.ExactArgs(2),
	RunE:  chunkIntervalSet,
}

func init() {
	chunkIntervalCmd.AddCommand(chunkIntervalSetCmd)
}

func chunkIntervalSet(cmd *cobra.Command, args []string) error {
	var err error

	metric := args[0]
	var chunk_interval time.Duration
	chunk_interval, err = time.ParseDuration(args[1])
	if err != nil {
		return fmt.Errorf("could not set chunk interval for %v: %w", metric, err)
	}

	if chunk_interval.Minutes() < 1.0 {
		return fmt.Errorf("could not set chunk interval for %v: %w", metric, errors.New("Chunk interval must be at least 1 minute"))
	}

	d, err := common.FormDBDetails(user, dbname, root.Namespace, root.HelmReleaseName)
	if err != nil {
		return err
	}

	pool, err := d.OpenConnectionToDB()
	if err != nil {
		return fmt.Errorf("could not set chunk interval for %v: %w", metric, err)
	}
	defer pool.Close()

	fmt.Printf("Setting chunk interval of %v to %v\n", metric, chunk_interval)
	_, err = pool.Exec(context.Background(), "SELECT prom_api.set_metric_chunk_interval($1, $2::INTERVAL)", metric, chunk_interval)
	if err != nil {
		return fmt.Errorf("could not set chunk interval for %v: %w", metric, err)
	}

	return nil
}
