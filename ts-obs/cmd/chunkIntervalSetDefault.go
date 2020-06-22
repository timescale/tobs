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
	chunkIntervalSetDefaultCmd.Flags().StringP("user", "U", "postgres", "database user name")
	chunkIntervalSetDefaultCmd.Flags().StringP("dbname", "d", "postgres", "database name to connect to")
}

func chunkIntervalSetDefault(cmd *cobra.Command, args []string) error {
	var err error

	if os.Getenv("PGPASSWORD") == "" {
		return errors.New("Password for user must be set in environment variable PGPASSWORD")
	}

	var chunk_interval time.Duration
	chunk_interval, err = time.ParseDuration(args[0])
	if err != nil {
		return err
	}

	if chunk_interval.Minutes() < 1.0 {
		return errors.New("Chunk interval must be at least 1 minute")
	}

	var user string
	user, err = cmd.Flags().GetString("user")
	if err != nil {
		return err
	}

	var dbname string
	dbname, err = cmd.Flags().GetString("dbname")
	if err != nil {
		return err
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
	pool, err = pgxpool.Connect(context.Background(), "postgres://"+user+":"+os.Getenv("PGPASSWORD")+"@localhost:"+strconv.Itoa(LISTEN_PORT_TSDB)+"/"+dbname)
	if err != nil {
		return err
	}
	defer pool.Close()

	fmt.Printf("Setting default chunk interval to %v\n", chunk_interval)
	_, err = pool.Exec(context.Background(), "SELECT prom_api.set_default_chunk_interval(INTERVAL '1 second' * $1)", strconv.FormatFloat(chunk_interval.Seconds(), 'f', -1, 64))
	if err != nil {
		return err
	}

	return nil
}
