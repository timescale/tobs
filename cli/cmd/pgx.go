package cmd

import (
	"context"
	"errors"
	"os"
	"strconv"

	"github.com/jackc/pgx/v4/pgxpool"
)

func OpenConnectionToDB(namespace, name, user, dbname string, remote int) (*pgxpool.Pool, error) {
	var pool *pgxpool.Pool
	var err error

	// Suppress output
	stdout := os.Stdout
	os.Stdout = nil
	defer func() { os.Stdout = stdout }()

	tspromPods, err := KubeGetPods(namespace, map[string]string{"app": name + "-promscale"})
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

	secret, err := KubeGetSecret(namespace, name+"-timescaledb-passwords")
	if err != nil {
		return nil, err
	}

	var pass string
	if bytepass, exists := secret.Data[user]; exists {
		pass = string(bytepass)
	} else {
		return nil, errors.New("user not found")
	}

	tsdbPods, err := KubeGetPods(namespace, map[string]string{"release": name, "role": "master"})
	if err != nil {
		return nil, err
	}

	if len(tsdbPods) != 0 {
		pf, err := KubePortForwardPod(namespace, tsdbPods[0].Name, 0, remote)
		if err != nil {
			return nil, err
		}

		ports, err := pf.GetPorts()
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
