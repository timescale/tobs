package cmd

import (
    "context"
    "errors"
	"fmt"
    "os"

	"github.com/spf13/cobra"
    "github.com/jackc/pgx/v4/pgxpool"
)

// chunkIntervalResetCmd represents the chunk-interval reset command
var chunkIntervalResetCmd = &cobra.Command{
	Use:   "reset <metric>",
	Short: "Resets the chunk interval for a specific metric back to the default",
	RunE:  chunkIntervalReset,
}

func init() {
	chunkIntervalCmd.AddCommand(chunkIntervalResetCmd)
}

func chunkIntervalReset(cmd *cobra.Command, args []string) error {
    var err error

    if len(args) != 1 {
        return errors.New("\"ts-obs metrics reset-metric-chunk-interval\" requires 1 arguments")
    }

    if os.Getenv("PGPASSWORD_POSTGRES")  == "" {
        return errors.New("Password for postgres user must be set in environment variable PGPASSWORD_POSTGRES")
    }

    metric := args[0]

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

    fmt.Printf("Resetting chunk interval for %v back to default\n", metric)
    _, err = pool.Exec(context.Background(), "SELECT reset_metric_chunk_interval('" + metric + "')")
    if err != nil {
        return err
    }

    return nil
}
