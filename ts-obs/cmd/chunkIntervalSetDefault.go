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

// chunkIntervalSetDefaultCmd represents the chunk-interval set-default command
var chunkIntervalSetDefaultCmd = &cobra.Command{
	Use:   "set-default <duration>",
	Short: "Sets default chunk interval in minutes",
	RunE:  chunkIntervalSetDefault,
}

func init() {
	chunkIntervalCmd.AddCommand(chunkIntervalSetDefaultCmd)
}

func chunkIntervalSetDefault(cmd *cobra.Command, args []string) error {
    var err error

    if len(args) != 1 {
        return errors.New("\"ts-obs metrics set-default-chunk\" requires 1 argument")
    }

    if os.Getenv("PGPASSWORD_POSTGRES")  == "" {
        return errors.New("Password for postgres user must be set in environment variable PGPASSWORD_POSTGRES")
    }

    var chunk_interval time.Duration
    chunk_interval, err = time.ParseDuration(args[0])
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

    fmt.Printf("Setting default chunk interval to %v\n", chunk_interval)
    _, err = pool.Exec(context.Background(), "SELECT set_default_chunk_interval(INTERVAL '1 second' * " + strconv.FormatFloat(chunk_interval.Seconds(), 'f', -1, 64) + ")")
    if err != nil {
        return err
    }

    return nil
}
