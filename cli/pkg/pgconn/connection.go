package pgconn

import (
	"context"
	"fmt"

	"net/url"
	"os"
	"strconv"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/timescale/tobs/cli/pkg/helm"
	"github.com/timescale/tobs/cli/pkg/k8s"
)

type DBDetails struct {
	Namespace           string
	ReleaseName         string
	DBName              string
	User                string
	SecretKey           string
	PromscaleSecretName string
	Password            string
	Remote              int
}

func (d *DBDetails) OpenConnectionToDB() (*pgxpool.Pool, error) {
	var pool *pgxpool.Pool
	var err error

	// Suppress output
	stdout := os.Stdout
	os.Stdout = nil
	defer func() { os.Stdout = stdout }()

	k8sClient := k8s.NewClient()

	secName, err := GetPromscaleSecretName(d.ReleaseName, d.Namespace)
	if err != nil {
		return nil, err
	}
	promscaleSecret, err := k8sClient.KubeGetSecret(d.Namespace, secName)
	if err != nil {
		return nil, err
	}

	secs := promscaleSecret.Data

	var port, host, sslmode, dbURI, pass string
	port = string(secs["PROM_DB_PORT"])
	host = string(secs["PROMSCALE_DB_HOST"])
	sslmode = string(secs["PROMSCALE_DB_SSL_MODE"])
	dbURI = string(secs["PROMSCALE_DB_URI"])
	pass = string(secs["PROMSCALE_DB_PASSWORD"])

	tsdbPods, err := k8sClient.KubeGetPods(d.Namespace, map[string]string{"release": d.ReleaseName, "role": "master"})
	if err != nil {
		return nil, err
	}

	if len(tsdbPods) != 0 {
		pf, err := k8sClient.KubePortForwardPod(d.Namespace, tsdbPods[0].Name, 0, d.Remote)
		if err != nil {
			return nil, err
		}

		ports, err := pf.GetPorts()
		if err != nil {
			return nil, err
		}

		connDetails := pgconn.Config{
			Host:     "localhost",
			Port:     ports[0].Local,
			Database: d.DBName,
			User:     d.User,
			Password: pass,
		}
		pool, err = pgxpool.Connect(context.Background(), ConstructURI(connDetails, sslmode))
		if err != nil {
			return nil, err
		}
	} else {
		if dbURI != "" {
			pool, err = pgxpool.Connect(context.Background(), dbURI)
			if err != nil {
				return nil, err
			}
		} else {
			iport, err := strconv.Atoi(port)
			if err != nil {
				return nil, err
			}

			connDetails := pgconn.Config{
				Host:     host,
				Port:     uint16(iport),
				Database: d.DBName,
				User:     d.User,
				Password: pass,
			}
			pool, err = pgxpool.Connect(context.Background(), ConstructURI(connDetails, sslmode))
			if err != nil {
				return nil, err
			}
		}
	}

	return pool, nil
}

func UpdatePasswordInDBURI(dburi, newpass string) (string, error) {
	db, err := pgxpool.ParseConfig(dburi)
	if err != nil {
		return "", err
	}

	var sslmode string
	if db.ConnConfig.TLSConfig == nil {
		sslmode = "allow"
	} else {
		sslmode = "require"
	}
	db.ConnConfig.Config.Password = newpass
	res := fmt.Sprint(ConstructURI(db.ConnConfig.Config, sslmode))
	return res, nil
}

func ParseDBURI(dbURI string) (*pgxpool.Config, error) {
	db, err := pgxpool.ParseConfig(dbURI)
	if err != nil {
		return db, err
	}

	return db, nil
}

func ConstructURI(connDetails pgconn.Config, sslmode string) string {
	c := new(url.URL)
	c.Scheme = "postgres"
	c.Host = fmt.Sprintf("%s:%d", connDetails.Host, connDetails.Port)
	if connDetails.Password != "" {
		c.User = url.UserPassword(connDetails.User, connDetails.Password)
	} else {
		c.User = url.User(connDetails.User)
	}
	c.Path = connDetails.Database
	q := c.Query()
	if sslmode != "" {
		q.Set("sslmode", sslmode)
	}

	if connDetails.ConnectTimeout != 0 {
		q.Set("connect_timeout", strconv.Itoa(int(connDetails.ConnectTimeout.Seconds())))
	}
	c.RawQuery = q.Encode()
	return c.String()
}

func GetPromscaleSecretName(releaseName, namespace string) (string, error) {
	helmClient := helm.NewClient(namespace)
	defer helmClient.Close()
	eS, err := helmClient.ExportValuesFieldFromRelease(releaseName, []string{"promscale", "connectionSecretName"})
	if err != nil {
		return "", err
	}
	secretProvided, ok := eS.(string)
	if !ok {
		return "", fmt.Errorf("promscale.connectionSecretName is not a string")
	}

	var secretName string
	if secretProvided == "" {
		secretName = releaseName + "-promscale"
	} else {
		secretName = secretProvided
	}

	return secretName, nil
}
