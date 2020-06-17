package cmd

import (
    "context"
    "errors"
	"fmt"
    "os"

	"github.com/spf13/cobra"
    "github.com/jackc/pgx/v4/pgxpool"
)

// retentionSetCmd represents the metrics retention set command
var retentionSetCmd = &cobra.Command{
	Use:   "set <metric> <days>",
	Short: "Sets data retention period in days for a specific metric",
	RunE:  retentionSet,
}

func init() {
	retentionCmd.AddCommand(retentionSetCmd)
}

func retentionSet(cmd *cobra.Command, args []string) error {
    var err error

    if len(args) != 2 {
        return errors.New("\"ts-obs metrics set-metric-retention\" requires 2 arguments")
    }

    if os.Getenv("PGPASSWORD_POSTGRES")  == "" {
        return errors.New("Password for postgres user must be set in environment variable PGPASSWORD_POSTGRES")
    }

    metric := args[0]
    retention_period := args[1]

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

    fmt.Printf("Setting retention period for %v to %v days\n", metric, retention_period)
    _, err = pool.Exec(context.Background(), "SELECT set_metric_retention_period('" + metric + "', INTERVAL '1 day' * " + retention_period + ")")
    if err != nil {
        return err
    }

    return nil
}
