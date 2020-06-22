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

// retentionResetCmd represents the retention reset command
var retentionResetCmd = &cobra.Command{
	Use:   "reset <metric>",
	Short: "Resets data retention period to default for a specific metric",
	Args:  cobra.ExactArgs(1),
	RunE:  retentionReset,
}

func init() {
	retentionCmd.AddCommand(retentionResetCmd)
	retentionResetCmd.Flags().StringP("user", "U", "postgres", "database user name")
	retentionResetCmd.Flags().StringP("dbname", "d", "postgres", "database name to connect to")
}

func retentionReset(cmd *cobra.Command, args []string) error {
	var err error

	if os.Getenv("PGPASSWORD") == "" {
		return errors.New("Password for user must be set in environment variable PGPASSWORD")
	}

	metric := args[0]

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

	fmt.Printf("Resetting retention period for %v back to default\n", metric)
	_, err = pool.Exec(context.Background(), "SELECT prom_api.reset_metric_retention_period($1)", metric)
	if err != nil {
		return err
	}

	return nil

}
