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

// retentionSetCmd represents the metrics retention set command
var retentionSetCmd = &cobra.Command{
	Use:   "set <metric> <days>",
	Short: "Sets data retention period in days for a specific metric",
	Args:  cobra.ExactArgs(2),
	RunE:  retentionSet,
}

func init() {
	retentionCmd.AddCommand(retentionSetCmd)
}

func retentionSet(cmd *cobra.Command, args []string) error {
	var err error

	if os.Getenv("PGPASSWORD_POSTGRES") == "" {
		return errors.New("Password for postgres user must be set in environment variable PGPASSWORD_POSTGRES")
	}

	metric := args[0]
	retention_period := args[1]

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

	fmt.Printf("Setting retention period for %v to %v days\n", metric, retention_period)
	_, err = pool.Exec(context.Background(), "SELECT prom_api.set_metric_retention_period('"+metric+"', INTERVAL '1 day' * "+retention_period+")")
	if err != nil {
		return err
	}

	return nil
}
