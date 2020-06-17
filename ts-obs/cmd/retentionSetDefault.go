package cmd

import (
    "context"
    "errors"
	"fmt"
    "os"

	"github.com/spf13/cobra"
    "github.com/jackc/pgx/v4/pgxpool"
)

// retentionSetDefaultCmd represents the retention set-default command
var retentionSetDefaultCmd = &cobra.Command{
	Use:   "set-default <days>",
	Short: "Sets default data retention period in days",
	RunE:  retentionSetDefault,
}

func init() {
	retentionCmd.AddCommand(retentionSetDefaultCmd)
}

func retentionSetDefault(cmd *cobra.Command, args []string) error {
    var err error

    if len(args) != 1 {
        return errors.New("\"ts-obs metrics set-default-retention\" requires 1 argument")
    }

    if os.Getenv("PGPASSWORD_POSTGRES")  == "" {
        return errors.New("Password for postgres user must be set in environment variable PGPASSWORD_POSTGRES")
    }

    retention_period := args[0]

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

    fmt.Printf("Setting default retention period to %v days\n", retention_period)
    _, err = pool.Exec(context.Background(), "SELECT set_default_retention_period(INTERVAL '1 day' * " + retention_period + ")")
    if err != nil {
        return err
    }

    return nil
}
