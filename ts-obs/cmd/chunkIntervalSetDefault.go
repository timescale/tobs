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

// chunkIntervalSetDefaultCmd represents the chunk-interval set-default command
var chunkIntervalSetDefaultCmd = &cobra.Command{
	Use:   "set-default <duration>",
	Short: "Sets default chunk interval in minutes",
	Args:  cobra.ExactArgs(1),
	RunE:  chunkIntervalSetDefault,
}

func init() {
	chunkIntervalCmd.AddCommand(chunkIntervalSetDefaultCmd)
}

func chunkIntervalSetDefault(cmd *cobra.Command, args []string) error {
	var err error

	if os.Getenv("PGPASSWORD_POSTGRES") == "" {
		return errors.New("Password for postgres user must be set in environment variable PGPASSWORD_POSTGRES")
	}

	var chunk_interval time.Duration
	chunk_interval, err = time.ParseDuration(args[0])
	if err != nil {
		return err
	}

	if chunk_interval.Minutes() < 1.0 {
		return errors.New("Chunk interval must be at least 1 minute")
	}

	var name string
	name, err = cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	var namespace string
	namespace, err = cmd.Flags().GetString("namespace")
	if err != nil {
		return err
	}

	podName, err := KubeGetPodName(namespace, map[string]string{"release": name, "role": "master"})
	if err != nil {
		return err
	}

	err = KubePortForwardPod(namespace, podName, LISTEN_PORT_TSDB, FORWARD_PORT_TSDB)
	if err != nil {
		return err
	}

	var pool *pgxpool.Pool
	pool, err = pgxpool.Connect(context.Background(), "postgres://postgres:"+os.Getenv("PGPASSWORD_POSTGRES")+"@localhost:"+strconv.Itoa(LISTEN_PORT_TSDB)+"/postgres")
	if err != nil {
		return err
	}
	defer pool.Close()

	fmt.Printf("Setting default chunk interval to %v\n", chunk_interval)
	_, err = pool.Exec(context.Background(), "SELECT prom_api.set_default_chunk_interval(INTERVAL '1 second' * "+strconv.FormatFloat(chunk_interval.Seconds(), 'f', -1, 64)+")")
	if err != nil {
		return err
	}

	return nil
}
