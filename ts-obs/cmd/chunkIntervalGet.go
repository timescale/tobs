package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
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
	chunkIntervalGetCmd.Flags().StringP("user", "U", "postgres", "database user name")
	chunkIntervalGetCmd.Flags().StringP("dbname", "d", "postgres", "database name to connect to")
}

func chunkIntervalGet(cmd *cobra.Command, args []string) error {
	var err error

	metric := args[0]

	var user string
	user, err = cmd.Flags().GetString("user")
	if err != nil {
		return fmt.Errorf("could not get chunk interval for %v: %w", metric, err)
	}

	var dbname string
	dbname, err = cmd.Flags().GetString("dbname")
	if err != nil {
		return fmt.Errorf("could not get chunk interval for %v: %w", metric, err)
	}

	var name string
	name, err = cmd.Flags().GetString("name")
	if err != nil {
		return fmt.Errorf("could not get chunk interval for %v: %w", metric, err)
	}

	var namespace string
	namespace, err = cmd.Flags().GetString("namespace")
	if err != nil {
		return fmt.Errorf("could not get chunk interval for %v: %w", metric, err)
	}

	pool, err := OpenConnectionToDB(namespace, name, user, dbname, FORWARD_PORT_TSDB)
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
