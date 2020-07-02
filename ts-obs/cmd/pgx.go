package cmd

import (
	"context"
	"errors"
	"os"
	"strconv"

	"github.com/jackc/pgx/v4/pgxpool"
)

func OpenConnectionToDB(namespace, name, user, dbname string, remote int) (*pgxpool.Pool, error) {
	var err error

	// Suppress output
	stdout := os.Stdout
	os.Stdout = nil
	defer func() { os.Stdout = stdout }()

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

	podName, err := KubeGetPodName(namespace, map[string]string{"release": name, "role": "master"})
	if err != nil {
		return nil, err
	}

	pf, err := KubePortForwardPod(namespace, podName, 0, remote)
	if err != nil {
		return nil, err
	}

	ports, err := pf.GetPorts()
	local := int(ports[0].Local)

	pool, err := pgxpool.Connect(context.Background(), "postgres://"+user+":"+pass+"@localhost:"+strconv.Itoa(local)+"/"+dbname)
	if err != nil {
		return nil, err
	}

	return pool, nil
}
