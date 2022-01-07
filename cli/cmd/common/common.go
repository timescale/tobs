package common

import (
	"errors"
	"fmt"

	root "github.com/timescale/tobs/cli/cmd"
	"github.com/timescale/tobs/cli/pkg/helm"
	"github.com/timescale/tobs/cli/pkg/k8s"
	"github.com/timescale/tobs/cli/pkg/pgconn"
)

const (
	LISTEN_PORT_GRAFANA    = 8080
	FORWARD_PORT_GRAFANA   = 3000
	LISTEN_PORT_PROM       = 9090
	FORWARD_PORT_PROM      = 9090
	LISTEN_PORT_PROMLENS   = 8081
	FORWARD_PORT_PROMLENS  = 8080
	LISTEN_PORT_PROMSCALE  = 9201
	FORWARD_PORT_PROMSCALE = 9201
	LISTEN_PORT_TSDB       = 5432
	FORWARD_PORT_TSDB      = 5432
	FORWARD_PORT_JAEGER    = 16686
	LISTEN_PORT_JAEGER     = 16686
)

var (
	TimescaleDBBackUpKeyForValuesYaml = []string{"timescaledb-single", "backup", "enabled"}
	PrometheusLabels                  = map[string]string{"app.kubernetes.io/managed-by": "prometheus-operator", "prometheus": "tobs-kube-prometheus-prometheus"}
	DBSuperUserSecretKey              = "PATRONI_SUPERUSER_PASSWORD"
	DBReplicationSecretKey            = "PATRONI_REPLICATION_PASSWORD"
	DBAdminSecretKey                  = "PATRONI_admin_PASSWORD"
)

func GetSuperuserDBDetails(namespace, releaseName string) (*pgconn.DBDetails, error) {
	helmClient := helm.NewClient(namespace)
	defer helmClient.Close()
	// use default super user from helm release
	// the default super-user password is mapped to "PATRONI_SUPERUSER_PASSWORD" secret key
	dbDetails := &pgconn.DBDetails{ReleaseName: releaseName, Namespace: namespace, SecretKey: DBSuperUserSecretKey, Remote: FORWARD_PORT_TSDB}
	err := getDBDetails(helmClient, dbDetails)
	if err != nil {
		return dbDetails, fmt.Errorf("could not get DB secret key from helm release: %w", err)
	}

	return dbDetails, nil
}

func getDBDetails(helmClient helm.Client, dbDetails *pgconn.DBDetails) error {
	e, err := helmClient.ExportValuesFieldFromRelease(dbDetails.ReleaseName, []string{"timescaledb-single", "enabled"})
	if err != nil {
		return err
	}
	enableTimescaleDB, ok := e.(bool)
	if !ok {
		return fmt.Errorf("timescaledb-single.enabled was not a bool")
	}

	if !enableTimescaleDB {
		dbURI, err := helmClient.ExportValuesFieldFromRelease(dbDetails.ReleaseName, []string{"timescaledbExternal", "db_uri"})
		if err != nil {
			return err
		}

		if dbURI == "" {
			return fmt.Errorf("failed to find timescaledbExternal URI at timescaledbExternal.db_uri")
		}

		uriDetails, err := pgconn.ParseDBURI(fmt.Sprint(dbURI))
		if err != nil {
			return err
		}
		dbDetails.User = uriDetails.ConnConfig.User
		dbDetails.DBName = uriDetails.ConnConfig.Database
		return nil
	}

	data, err := helmClient.ExportValuesFieldFromRelease(dbDetails.ReleaseName, []string{"timescaledb-single", "patroni", "postgresql", "authentication", "superuser", "username"})
	if err != nil {
		return err
	}

	dbname, err := helmClient.ExportValuesFieldFromRelease(dbDetails.ReleaseName, []string{"promscale", "connection", "dbName"})
	if err != nil {
		return err
	}

	dbDetails.User = fmt.Sprint(data)
	dbDetails.DBName = fmt.Sprint(dbname)

	k8sClient := k8s.NewClient()
	secret, err := k8sClient.KubeGetSecret(root.Namespace, root.HelmReleaseName+"-credentials")
	if err != nil {
		return fmt.Errorf("could not get TimescaleDB password: %w", err)
	}

	if bytepass, exists := secret.Data[dbDetails.SecretKey]; exists {
		dbDetails.Password = string(bytepass)
	} else {
		return fmt.Errorf("could not get TimescaleDB password: %w", errors.New("user not found"))
	}

	return nil
}

func GetTimescaleDBLabels(releaseName string) map[string]string {
	return map[string]string{
		"app":     releaseName + "-timescaledb",
		"release": releaseName,
	}
}
