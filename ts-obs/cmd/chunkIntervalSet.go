package cmd

import (
    "context"
    "errors"
	"fmt"
    "os"
    "strconv"
    "time"

	"github.com/spf13/cobra"
    "github.com/jackc/pgx/v4/pgxpool"
)

// chunkIntervalSetCmd represents the chunk-interval set command
var chunkIntervalSetCmd = &cobra.Command{
	Use:   "set <metric> <duration>",
	Short: "Sets chunk interval in minutes for a specific metric",
	RunE:  chunkIntervalSet,
}

func init() {
	chunkIntervalCmd.AddCommand(chunkIntervalSetCmd)
}

func chunkIntervalSet(cmd *cobra.Command, args []string) error {
    var err error

    if len(args) != 2 {
        return errors.New("\"ts-obs metrics set-metric-chunk\" requires 2 arguments")
    }

    if os.Getenv("PGPASSWORD_POSTGRES")  == "" {
        return errors.New("Password for postgres user must be set in environment variable PGPASSWORD_POSTGRES")
    }

    metric := args[0]
    var chunk_interval time.Duration
    chunk_interval, err = time.ParseDuration(args[1])
    if err != nil {
        return err
    }

    if chunk_interval.Microseconds() < 1 {
        return errors.New("Chunk interval must be at least 1 microsecond")
    }

    err = kubePortForwardPod("ts-obs-timescaledb-0", 5432, 5432)
    if err != nil {
        return err
    }

    var pool *pgxpool.Pool
    pool, err = pgxpool.Connect(context.Background(), "postgres://postgres:" + os.Getenv("PGPASSWORD_POSTGRES") + "@localhost:5432/postgres")
    if err != nil {
        return err
    }
    defer pool.Close()

    fmt.Printf("Setting chunk interval of %v to %v\n", metric, chunk_interval)
    _, err = pool.Exec(context.Background(), "SELECT set_metric_chunk_interval('" + metric + "', INTERVAL '1 second' * " + strconv.FormatFloat(chunk_interval.Seconds(), 'f', -1, 64) + ")")
    if err != nil {
        return err
    }

    return nil
}
