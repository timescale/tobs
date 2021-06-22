package pgconn

import (
	"context"
	"fmt"
	"github.com/timescale/tobs/cli/pkg/k8s"
	"os"
	"strconv"

	"github.com/jackc/pgx/v4/pgxpool"
)

type DBDetails struct {
	Namespace string
	Name      string
	DBName    string
	User      string
	SecretKey string
	Remote    int
}

func (d *DBDetails) OpenConnectionToDB() (*pgxpool.Pool, error) {
	var pool *pgxpool.Pool
	var err error

	// Suppress output
	stdout := os.Stdout
	os.Stdout = nil
	defer func() { os.Stdout = stdout }()

	kubeClient, _ := k8s.NewClient()
	tspromPods, err := kubeClient.KubeGetPods(d.Namespace, map[string]string{"app": d.Name + "-promscale"})
	if err != nil {
		return nil, err
	}

	passBytes, err := kubeClient.GetDBPassword(d.SecretKey, d.Name, d.Namespace)
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

	dbURI, err := kubeClient.GetTimescaleDBURI(d.Namespace, d.Name)
	if err != nil {
		return nil, err
	}

	tsdbPods, err := kubeClient.KubeGetPods(d.Namespace, map[string]string{"release": d.Name, "role": "master"})
	if err != nil {
		return nil, err
	}

	if len(tsdbPods) != 0 {
		pf, err := kubeClient.KubePortForwardPod(d.Namespace, tsdbPods[0].Name, 0, d.Remote)
		if err != nil {
			return nil, err
		}

		ports, err := pf.GetPorts()
		if err != nil {
			return nil, err
		}
		local := int(ports[0].Local)

		pool, err = pgxpool.Connect(context.Background(), fmt.Sprint("postgres://"+d.User+":"+pass+"@localhost:"+strconv.Itoa(local)+"/"+d.DBName))
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
			pool, err = pgxpool.Connect(context.Background(), fmt.Sprint("postgres://"+d.User+":"+pass+"@"+host+":"+port+"/"+d.DBName+"?sslmode="+sslmode))
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
	port := strconv.Itoa(int(db.ConnConfig.Port))
	connectTimeOut := ""
	if db.ConnConfig.ConnectTimeout.String() != "0s" {
		connectTimeOut = "&connect_timeout=" + fmt.Sprintf("%.f", db.ConnConfig.ConnectTimeout.Seconds())
	}
	res := fmt.Sprint("postgres://" + db.ConnConfig.User + ":" + newpass + "@" + db.ConnConfig.Host + ":" + port + "/" + db.ConnConfig.Database + "?sslmode=" + sslmode + connectTimeOut)
	return res, nil
}

func ParseDBURI(dbURI string) (*pgxpool.Config, error) {
	db, err := pgxpool.ParseConfig(dbURI)
	if err != nil {
		return db, err
	}

	return db, nil
}
