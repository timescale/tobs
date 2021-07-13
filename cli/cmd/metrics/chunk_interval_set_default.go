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

// chunkIntervalSetDefaultCmd represents the chunk-interval set-default command
var chunkIntervalSetDefaultCmd = &cobra.Command{
	Use:   "set-default <duration>",
	Short: "Sets default chunk interval",
	Args:  cobra.ExactArgs(1),
	RunE:  chunkIntervalSetDefault,
}

func init() {
	chunkIntervalCmd.AddCommand(chunkIntervalSetDefaultCmd)
}

func chunkIntervalSetDefault(cmd *cobra.Command, args []string) error {
	var err error

	var chunk_interval time.Duration
	chunk_interval, err = time.ParseDuration(args[0])
	if err != nil {
		return fmt.Errorf("could not set default chunk interval: %w", err)
	}

	if chunk_interval.Minutes() < 1.0 {
		return fmt.Errorf("could not set default chunk interval: %w", errors.New("Chunk interval must be at least 1 minute"))
	}

	d, err := common.FormDBDetails(user, dbname, root.Namespace, root.HelmReleaseName)
	if err != nil {
		return err
	}

	pool, err := d.OpenConnectionToDB()
	if err != nil {
		return fmt.Errorf("could not set default chunk interval: %w", err)
	}
	defer pool.Close()

	fmt.Printf("Setting default chunk interval to %v\n", chunk_interval)
	_, err = pool.Exec(context.Background(), "SELECT prom_api.set_default_chunk_interval($1::INTERVAL)", chunk_interval)
	if err != nil {
		return fmt.Errorf("could not set default chunk interval: %w", err)
	}

	return nil
}
