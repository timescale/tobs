package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/spf13/cobra"
)

// chunkIntervalSetCmd represents the chunk-interval set command
var chunkIntervalSetCmd = &cobra.Command{
	Use:   "set <metric> <duration>",
	Short: "Sets chunk interval in minutes for a specific metric",
	Args:  cobra.ExactArgs(2),
	RunE:  chunkIntervalSet,
}

func init() {
	chunkIntervalCmd.AddCommand(chunkIntervalSetCmd)
	chunkIntervalSetCmd.Flags().StringP("user", "U", "postgres", "database user name")
	chunkIntervalSetCmd.Flags().StringP("dbname", "d", "postgres", "database name to connect to")
}

func chunkIntervalSet(cmd *cobra.Command, args []string) error {
	var err error

	metric := args[0]
	var chunk_interval time.Duration
	chunk_interval, err = time.ParseDuration(args[1])
	if err != nil {
		return fmt.Errorf("could not set chunk interval for %v: %w", metric, err)
	}

	if os.Getenv("PGPASSWORD") == "" {
		return fmt.Errorf("could not set chunk interval for %v: %w", metric, errors.New("password for user must be set in environment variable PGPASSWORD"))
	}

	var user string
	user, err = cmd.Flags().GetString("user")
	if err != nil {
		return fmt.Errorf("could not set chunk interval for %v: %w", metric, err)
	}

	var dbname string
	dbname, err = cmd.Flags().GetString("dbname")
	if err != nil {
		return fmt.Errorf("could not set chunk interval for %v: %w", metric, err)
	}

	var name string
	name, err = cmd.Flags().GetString("name")
	if err != nil {
		return fmt.Errorf("could not set chunk interval for %v: %w", metric, err)
	}

	var namespace string
	namespace, err = cmd.Flags().GetString("namespace")
	if err != nil {
		return fmt.Errorf("could not set chunk interval for %v: %w", metric, err)
	}

	if chunk_interval.Minutes() < 1.0 {
		return fmt.Errorf("could not set chunk interval for %v: %w", metric, errors.New("Chunk interval must be at least 1 minute"))
	}

	podName, err := KubeGetPodName(namespace, map[string]string{"release": name, "role": "master"})
	if err != nil {
		return fmt.Errorf("could not set chunk interval for %v: %w", metric, err)
	}

	err = KubePortForwardPod(namespace, podName, LISTEN_PORT_TSDB, FORWARD_PORT_TSDB)
	if err != nil {
		return fmt.Errorf("could not set chunk interval for %v: %w", metric, err)
	}

	var pool *pgxpool.Pool
	pool, err = pgxpool.Connect(context.Background(), "postgres://"+user+":"+os.Getenv("PGPASSWORD")+"@localhost:"+strconv.Itoa(LISTEN_PORT_TSDB)+"/"+dbname)
	if err != nil {
		return fmt.Errorf("could not set chunk interval for %v: %w", metric, err)
	}
	defer pool.Close()

	fmt.Printf("Setting chunk interval of %v to %v\n", metric, chunk_interval)
	_, err = pool.Exec(context.Background(), "SELECT prom_api.set_metric_chunk_interval($1, INTERVAL '1 second' * $2)", metric, strconv.FormatFloat(chunk_interval.Seconds(), 'f', -1, 64))
	if err != nil {
		return fmt.Errorf("could not set chunk interval for %v: %w", metric, err)
	}

	return nil
}
