package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/spf13/cobra"
)

// retentionSetDefaultCmd represents the retention set-default command
var retentionSetDefaultCmd = &cobra.Command{
	Use:   "set-default <days>",
	Short: "Sets default data retention period in days",
	Args:  cobra.ExactArgs(1),
	RunE:  retentionSetDefault,
}

func init() {
	retentionCmd.AddCommand(retentionSetDefaultCmd)
	retentionSetDefaultCmd.Flags().StringP("user", "U", "postgres", "database user name")
	retentionSetDefaultCmd.Flags().StringP("dbname", "d", "postgres", "database name to connect to")
}

func retentionSetDefault(cmd *cobra.Command, args []string) error {
	var err error

	retention_period := args[0]

	if os.Getenv("PGPASSWORD") == "" {
		return fmt.Errorf("could not set default retention period: %w", errors.New("password for postgres user must be set in environment variable PGPASSWORD"))
	}

	var user string
	user, err = cmd.Flags().GetString("user")
	if err != nil {
		return fmt.Errorf("could not set default retention period: %w", err)
	}

	var dbname string
	dbname, err = cmd.Flags().GetString("dbname")
	if err != nil {
		return fmt.Errorf("could not set default retention period: %w", err)
	}

	var name string
	name, err = cmd.Flags().GetString("name")
	if err != nil {
		return fmt.Errorf("could not set default retention period: %w", err)
	}

	var namespace string
	namespace, err = cmd.Flags().GetString("namespace")
	if err != nil {
		return fmt.Errorf("could not set default retention period: %w", err)
	}

	podName, err := KubeGetPodName(namespace, map[string]string{"release": name, "role": "master"})
	if err != nil {
		return fmt.Errorf("could not set default retention period: %w", err)
	}

	err = KubePortForwardPod(namespace, podName, LISTEN_PORT_TSDB, FORWARD_PORT_TSDB)
	if err != nil {
		return fmt.Errorf("could not set default retention period: %w", err)
	}

	var pool *pgxpool.Pool
	pool, err = pgxpool.Connect(context.Background(), "postgres://"+user+":"+os.Getenv("PGPASSWORD")+"@localhost:"+strconv.Itoa(LISTEN_PORT_TSDB)+"/"+dbname)
	if err != nil {
		return fmt.Errorf("could not set default retention period: %w", err)
	}
	defer pool.Close()

	fmt.Printf("Setting default retention period to %v days\n", retention_period)
	_, err = pool.Exec(context.Background(), "SELECT prom_api.set_default_retention_period(INTERVAL '1 day' * $1)", retention_period)
	if err != nil {
		return fmt.Errorf("could not set default retention period: %w", err)
	}

	return nil
}
