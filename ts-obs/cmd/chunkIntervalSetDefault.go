package cmd

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

// chunkIntervalSetDefaultCmd represents the chunk-interval set-default command
var chunkIntervalSetDefaultCmd = &cobra.Command{
	Use:   "set-default <duration>",
	Short: "Sets default chunk interval in minutes",
	Args:  cobra.ExactArgs(1),
	RunE:  chunkIntervalSetDefault,
}

func init() {
	chunkIntervalCmd.AddCommand(chunkIntervalSetDefaultCmd)
	chunkIntervalSetDefaultCmd.Flags().StringP("user", "U", "postgres", "database user name")
	chunkIntervalSetDefaultCmd.Flags().StringP("dbname", "d", "postgres", "database name to connect to")
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

	var user string
	user, err = cmd.Flags().GetString("user")
	if err != nil {
		return fmt.Errorf("could not set default chunk interval: %w", err)
	}

	var dbname string
	dbname, err = cmd.Flags().GetString("dbname")
	if err != nil {
		return fmt.Errorf("could not set default chunk interval: %w", err)
	}

	pool, err := OpenConnectionToDB(namespace, name, user, dbname, FORWARD_PORT_TSDB)
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
