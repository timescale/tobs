package cmd

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
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

	var name string
	name, err = cmd.Flags().GetString("name")
	if err != nil {
		return fmt.Errorf("could not set default chunk interval: %w", err)
	}

	var namespace string
	namespace, err = cmd.Flags().GetString("namespace")
	if err != nil {
		return fmt.Errorf("could not set default chunk interval: %w", err)
	}

	secret, err := KubeGetSecret(namespace, name+"-timescaledb-passwords")
	if err != nil {
		return fmt.Errorf("could not get TimescaleDB password: %w", err)
	}

	pass := string(secret.Data[user])

	podName, err := KubeGetPodName(namespace, map[string]string{"release": name, "role": "master"})
	if err != nil {
		return fmt.Errorf("could not set default chunk interval: %w", err)
	}

	err = KubePortForwardPod(namespace, podName, LISTEN_PORT_TSDB, FORWARD_PORT_TSDB)
	if err != nil {
		return fmt.Errorf("could not set default chunk interval: %w", err)
	}

	var pool *pgxpool.Pool
	pool, err = pgxpool.Connect(context.Background(), "postgres://"+user+":"+pass+"@localhost:"+strconv.Itoa(LISTEN_PORT_TSDB)+"/"+dbname)
	if err != nil {
		return fmt.Errorf("could not set default chunk interval: %w", err)
	}
	defer pool.Close()

	fmt.Printf("Setting default chunk interval to %v\n", chunk_interval)
	_, err = pool.Exec(context.Background(), "SELECT prom_api.set_default_chunk_interval(INTERVAL '1 second' * $1)", strconv.FormatFloat(chunk_interval.Seconds(), 'f', -1, 64))
	if err != nil {
		return fmt.Errorf("could not set default chunk interval: %w", err)
	}

	return nil
}
