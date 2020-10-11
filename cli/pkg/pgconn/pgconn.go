package pgconn

import (
	"context"
	"errors"
	"os"
	"strconv"

	"github.com/timescale/tobs/cli/pkg/k8s"

	"github.com/jackc/pgx/v4/pgxpool"
)

func OpenConnectionToDB(namespace, name, user, dbname string, remote int) (*pgxpool.Pool, error) {
	var pool *pgxpool.Pool
	var err error

	// Suppress output
	stdout := os.Stdout
	os.Stdout = nil
	defer func() { os.Stdout = stdout }()

	tspromPods, err := k8s.KubeGetPods(namespace, map[string]string{"app": name + "-promscale"})
	if err != nil {
		return nil, err
	}

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

	secret, err := k8s.KubeGetSecret(namespace, name+"-timescaledb-passwords")
	if err != nil {
		return nil, err
	}

	var pass string
	if bytepass, exists := secret.Data[user]; exists {
		pass = string(bytepass)
	} else {
		return nil, errors.New("user not found")
	}

	tsdbPods, err := k8s.KubeGetPods(namespace, map[string]string{"release": name, "role": "master"})
	if err != nil {
		return nil, err
	}

	if len(tsdbPods) != 0 {
		pf, err := k8s.KubePortForwardPod(namespace, tsdbPods[0].Name, 0, remote)
		if err != nil {
			return nil, err
		}

		ports, err := pf.GetPorts()
		if err != nil {
			return nil, err
		}
		local := int(ports[0].Local)

		pool, err = pgxpool.Connect(context.Background(), "postgres://"+user+":"+pass+"@localhost:"+strconv.Itoa(local)+"/"+dbname)
		if err != nil {
			return nil, err
		}
	} else {
		pool, err = pgxpool.Connect(context.Background(), "postgres://"+user+":"+pass+"@"+host+":"+port+"/tsdb?sslmode="+sslmode)
		if err != nil {
			return nil, err
		}
	}

	return pool, nil
}
