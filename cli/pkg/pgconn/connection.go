package pgconn

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strconv"

	"github.com/timescale/tobs/cli/pkg/utils"

	"github.com/timescale/tobs/cli/pkg/k8s"

	"github.com/jackc/pgx/v4/pgxpool"
)

type DBDetails struct {
	Namespace   string
	ReleaseName string
	DBName      string
	User        string
	SecretKey   string
	Password    string
	Remote      int
}

func (d *DBDetails) OpenConnectionToDB() (*pgxpool.Pool, error) {
	var pool *pgxpool.Pool
	var err error

	// Suppress output
	stdout := os.Stdout
	os.Stdout = nil
	defer func() { os.Stdout = stdout }()

	k8sClient := k8s.NewClient()
	tspromPods, err := k8sClient.KubeGetPods(d.Namespace, map[string]string{"app": d.ReleaseName + "-promscale"})
	if err != nil {
		return nil, err
	}

	passBytes, err := utils.GetDBPassword(k8sClient, d.SecretKey, d.ReleaseName, d.Namespace)
	if err != nil {
		return nil, err
	}
	pass := string(passBytes)

	envs := tspromPods[0].Spec.Containers[0].Env

	var port, host, sslmode string
	for _, env := range envs {
		if env.Name == "TS_PROM_DB_PORT" {
			port = env.Value
		} else if env.Name == "TS_PROM_DB_HOST" {
			host = env.Value
		} else if env.Name == "TS_PROM_DB_SSL_MODE" {
			sslmode = env.Value
		}
	}

	dbURI, err := utils.GetTimescaleDBURI(k8sClient, d.Namespace, d.ReleaseName)
	if err != nil {
		return nil, err
	}

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
		local := int(ports[0].Local)

		pool, err = pgxpool.Connect(context.Background(), constructURI(d.User, pass, "localhost", local, d.DBName, "", 0))
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
			pool, err = pgxpool.Connect(context.Background(), constructURI(d.User, pass, host, iport, d.DBName, sslmode, 0))
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
	port := int(db.ConnConfig.Port)
	connectTimeOut := 0
	if db.ConnConfig.ConnectTimeout.String() != "0s" {
		connectTimeOut = int(db.ConnConfig.ConnectTimeout.Seconds())
	}
	res := fmt.Sprint(constructURI(db.ConnConfig.User, newpass, db.ConnConfig.Host, port, db.ConnConfig.Database, sslmode, connectTimeOut))
	return res, nil
}

func ParseDBURI(dbURI string) (*pgxpool.Config, error) {
	db, err := pgxpool.ParseConfig(dbURI)
	if err != nil {
		return db, err
	}

	return db, nil
}

func constructURI(user, password, host string, port int, dbname, sslmode string, connectTimeout int) string {
	c := new(url.URL)
	c.Scheme = "postgres"
	c.Host = fmt.Sprintf("%s:%d", host, port)
	if password != "" {
		c.User = url.UserPassword(user, password)
	} else {
		c.User = url.User(user)
	}
	c.Path = dbname
	q := c.Query()
	if sslmode != "" {
		q.Set("sslmode", sslmode)
	}
	if connectTimeout != 0 {
		q.Set("connect_timeout", strconv.Itoa(connectTimeout))
	}
	c.RawQuery = q.Encode()
	return c.String()
}
