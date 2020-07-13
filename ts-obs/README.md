# Timescale Observability CLI

This is a CLI tool for installing and managing the Timescale Observability Helm chart. 

## Quick Start

__Dependencies__: [Go](https://golang.org/doc/install), [Helm](https://helm.sh/docs/intro/install/)

To install the CLI, run `go install` from inside the `ts-obs` folder. 
Then, copy the `ts-obs` binary from `$GOPATH/bin` to your `/bin` folder. 

## Commands

The following are the commands possible with the CLI. 

### Base Commands

| Command               | Description                                                      | Flags                                                |
|-----------------------|------------------------------------------------------------------|------------------------------------------------------|
| `ts-obs install`      | Alias for `ts-obs helm install`.                                 | `--filename`, `-f` : file to load configuration from |
| `ts-obs uninstall`    | Alias for `ts-obs helm unintall`.                                | None                                                 |
| `ts-obs port-forward` | Port-forwards TimescaleDB, Grafana, and Prometheus to localhost. | `--timescaledb`, `-t` : port for TimescaleDB <br> `--grafana`, `-g` : port for Grafana <br> `--prometheus`, `-p` : port for Prometheus |

### Helm Commands

| Command                   | Description                                                               | Flags                                                |
|---------------------------|---------------------------------------------------------------------------|------------------------------------------------------|
| `ts-obs helm install`     | Installs the Timescale Observability Helm chart.                          | `--filename`, `-f` : file to load configuration from |
| `ts-obs helm uninstall`   | Uninstalls the Timescale Observability Helm chart.                        | None                                                 |
| `ts-obs helm show-values` | Prints the YAML configuration of the Timescale Observability Helm chart.  | None                                                 |
| `ts-obs helm delete-data` | Deletes persistent volume claims associated with Timescale Observability. | None                                                 |

### TimescaleDB Commands

| Command                           | Description                                                | Flags                                       |
|-----------------------------------|------------------------------------------------------------|---------------------------------------------|
| `ts-obs timescaledb connect`         | Connects to the Timescale database running in the cluster. | `--user`, `-U` : user to login with <br> `--master`, `-m` : directly execute session on master node |
| `ts-obs timescaledb port-forward`    | Port-forwards TimescaleDB to localhost.                    | `--port`, `-p` : port to listen from        |
| `ts-obs timescaledb get-password`    | Gets the password for a user in the Timescale database.    | `--user`, `-U` : user whose password to get |
| `ts-obs timescaledb change-password` | Changes the password for a user in the Timescale database. | `--user`, `-U` : user whose password to get |

### Grafana Commands

| Command                               | Description                                    | Flags                                |
|---------------------------------------|------------------------------------------------|--------------------------------------|
| `ts-obs grafana port-forward`         | Port-forwards the Grafana server to localhost. | `--port`, `-p` : port to listen from |
| `ts-obs grafana get-initial-password` | Gets the initial admin password for Grafana.   | None                                 |
| `ts-obs grafana change-password`      | Changes the admin password for Grafana.        | None                                 |

### Prometheus Commands

| Command                          | Description                                       | Flags                                |
|----------------------------------|---------------------------------------------------|--------------------------------------|
| `ts-obs prometheus port-forward` | Port-forwards the Prometheus server to localhost. | `--port`, `-p` : port to listen from |

### Metrics Commands

| Command                                     | Description                                                                          | Flags |
|---------------------------------------------|--------------------------------------------------------------------------------------|-------|
| `ts-obs metrics retention get`              | Gets the data retention period of a specific metric.                                 | `--user`, `-U` : database user name <br> `--dbname`, `-d` : database name to connect to |
| `ts-obs metrics retention set-default`      | Sets the default data retention period to the specified number of days.              | `--user`, `-U` : database user name <br> `--dbname`, `-d` : database name to connect to |
| `ts-obs metrics retention set`              | Sets the data retention period of a specific metric to the specified number of days. | `--user`, `-U` : database user name <br> `--dbname`, `-d` : database name to connect to |
| `ts-obs metrics retention reset`            | Resets the data retention period of a specific metric to the default value.          | `--user`, `-U` : database user name <br> `--dbname`, `-d` : database name to connect to |
| `ts-obs metrics chunk-interval get`         | Gets the chunk interval of a specific metric.                                        | `--user`, `-U` : database user name <br> `--dbname`, `-d` : database name to connect to |
| `ts-obs metrics chunk-interval set-default` | Sets the default chunk interval to the specified duration.                           | `--user`, `-U` : database user name <br> `--dbname`, `-d` : database name to connect to |
| `ts-obs metrics chunk-interval set`         | Sets the chunk interval of a specific metric to the specified duration.              | `--user`, `-U` : database user name <br> `--dbname`, `-d` : database name to connect to |
| `ts-obs metrics chunk-interval reset`       | Resets chunk interval of a specific metric to the default value.                     | `--user`, `-U` : database user name <br> `--dbname`, `-d` : database name to connect to |

## Global Flags

The following are global flags that can be used with any of the above commands: 

| Flag           | Description          |
|----------------|----------------------|
| `--name`, `-n` | Helm release name    |
| `--namespace`  | Kubernetes namespace |

## Testing

A testing suite is included in the `tests` folder. This testing suite has additional dependencies on [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) and [minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/).

The testing suite can be run by calling `go test -timeout 30m` from within the `tests` folder. 
