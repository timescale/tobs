# tobs - The Observability Stack for Kubernetes CLI

This is a CLI tool for installing and managing the The Observability Stack for Kubernetes.

## Quick Start

__Dependencies__: [Helm](https://helm.sh/docs/intro/install/)

Getting started with the CLI tool is a two-step process: First you install the CLI tool locally, then you use the CLI tool to install the tobs stack into your Kubernetes cluster.

### Installing the CLI tool

To download and install tobs, run the following in your terminal, then follow the on-screen instructions.

```bash
curl --proto '=https' --tlsv1.2 -sSLf  https://tsdb.co/install-tobs-sh |sh
```

Alternatively, you can download the CLI directly via [our releases page](/releases)

### Using the tobs CLI tool to deploy the stack into your Kubernetes cluster

After setting up tobs run the following to install the tobs helm charts into your Kubernetes cluster

```bash
tobs install
```

This will deploy all of the tobs components into your cluster and provide instructions as to next steps.

### Getting started by viewing your metrics in Grafana
To see your Grafana dashboards after installation run

```bash
tobs grafana change-password <new_password>
tobs grafana port-forward
```
Then, point your browser to http://127.0.0.1:8080/ and login with the `admin` username.

## Commands

The following are the commands possible with the CLI.

### Base Commands

| Command             | Description                                                      | Flags                                                |
|---------------------|------------------------------------------------------------------|------------------------------------------------------|
| `tobs install`      | Alias for `tobs helm install`.                                   | `--filename`, `-f` : file to load configuration from |
| `tobs uninstall`    | Alias for `tobs helm unintall`.                                  | None                                                 |
| `tobs port-forward` | Port-forwards TimescaleDB, Grafana, and Prometheus to localhost. | `--timescaledb`, `-t` : port for TimescaleDB <br> `--grafana`, `-g` : port for Grafana <br> `--prometheus`, `-p` : port for Prometheus |

### Helm Commands

Documentation about Helm configuration can be found in the [Helm chart directory](/chart/README.md).

| Command                 | Description                                                                  | Flags                                                |
|-------------------------|------------------------------------------------------------------------------|------------------------------------------------------|
| `tobs helm install`     | Installs Helm chart for The Observability Stack.                             | `--filename`, `-f` : file to load configuration from |
| `tobs helm uninstall`   | Uninstalls Helm chart for The Observability Stack.                           | None                                                 |
| `tobs helm show-values` | Prints the YAML configuration of the Helm chart for The Observability Stack. | None                                                 |
| `tobs helm delete-data` | Deletes persistent volume claims associated with The Observability Stack.    | None                                                 |

### TimescaleDB Commands

| Command                            | Description                                                | Flags                                       |
|------------------------------------|------------------------------------------------------------|---------------------------------------------|
| `tobs timescaledb connect`         | Connects to the Timescale database running in the cluster. | `--user`, `-U` : user to login with <br> `--master`, `-m` : directly execute session on master node |
| `tobs timescaledb port-forward`    | Port-forwards TimescaleDB to localhost.                    | `--port`, `-p` : port to listen from        |
| `tobs timescaledb get-password`    | Gets the password for a user in the Timescale database.    | `--user`, `-U` : user whose password to get |
| `tobs timescaledb change-password` | Changes the password for a user in the Timescale database. | `--user`, `-U` : user whose password to get |

### Grafana Commands

| Command                             | Description                                    | Flags                                |
|-------------------------------------|------------------------------------------------|--------------------------------------|
| `tobs grafana port-forward`         | Port-forwards the Grafana server to localhost. | `--port`, `-p` : port to listen from |
| `tobs grafana get-initial-password` | Gets the initial admin password for Grafana.   | None                                 |
| `tobs grafana change-password`      | Changes the admin password for Grafana.        | None                                 |

### Prometheus Commands

| Command                        | Description                                       | Flags                                |
|--------------------------------|---------------------------------------------------|--------------------------------------|
| `tobs prometheus port-forward` | Port-forwards the Prometheus server to localhost. | `--port`, `-p` : port to listen from |

### Metrics Commands

| Command                                   | Description                                                                          | Flags |
|-------------------------------------------|--------------------------------------------------------------------------------------|-------|
| `tobs metrics retention get`              | Gets the data retention period of a specific metric.                                 | `--user`, `-U` : database user name <br> `--dbname`, `-d` : database name to connect to |
| `tobs metrics retention set-default`      | Sets the default data retention period to the specified number of days.              | `--user`, `-U` : database user name <br> `--dbname`, `-d` : database name to connect to |
| `tobs metrics retention set`              | Sets the data retention period of a specific metric to the specified number of days. | `--user`, `-U` : database user name <br> `--dbname`, `-d` : database name to connect to |
| `tobs metrics retention reset`            | Resets the data retention period of a specific metric to the default value.          | `--user`, `-U` : database user name <br> `--dbname`, `-d` : database name to connect to |
| `tobs metrics chunk-interval get`         | Gets the chunk interval of a specific metric.                                        | `--user`, `-U` : database user name <br> `--dbname`, `-d` : database name to connect to |
| `tobs metrics chunk-interval set-default` | Sets the default chunk interval to the specified duration.                           | `--user`, `-U` : database user name <br> `--dbname`, `-d` : database name to connect to |
| `tobs metrics chunk-interval set`         | Sets the chunk interval of a specific metric to the specified duration.              | `--user`, `-U` : database user name <br> `--dbname`, `-d` : database name to connect to |
| `tobs metrics chunk-interval reset`       | Resets chunk interval of a specific metric to the default value.                     | `--user`, `-U` : database user name <br> `--dbname`, `-d` : database name to connect to |

### Volume Commands

The volume operation is available for TimescaleDB & Prometheus PVC's.

**Note**: To expand PVC's in Kubernetes cluster make sure you have configured `storageClass` with `allowVolumeExpansion: true` to allow PVC expansion.

| Command                        | Description                                       | Flags                                |
|--------------------------------|---------------------------------------------------|--------------------------------------|
| `tobs volume get`     | Displays Persistent Volume Claims sizes. | `--timescaleDB-storage`, `s`, `--timescaleDB-wal`, `w`, `prometheus-storage`, `-p`  |
| `tobs volume expand`  | Expands the Persistent Volume Claims for provided resources to specified sizes. The expansion size is allowed in `Ki`, `Mi` & `Gi` units. example: `150Gi`. | `--timescaleDB-storage`, `s`, `--timescaleDB-wal`, `w`, `prometheus-storage`, `-p` |

## Global Flags

The following are global flags that can be used with any of the above commands:

| Flag           | Description          |
|----------------|----------------------|
| `--name`, `-n` | Helm release name    |
| `--namespace`  | Kubernetes namespace |

## Advanced configuration

Documentation about Helm configuration can be found in the [Helm chart directory](/chart/README.md).
Custom values.yml files can be used with the `tobs helm install -f values.yml` command.

## Building from source

__Dependencies__: [Go](https://golang.org/doc/install), [Helm](https://helm.sh/docs/intro/install/)

To build from source, run `go build -o tobs` from inside the `cli` folder.
Then, move the `tobs` binary from the current directory to your `/bin` folder.

## Testing

__Dependencies__: [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/), [kind](https://kind.sigs.k8s.io/)

A testing suite is included in the `tests` folder. The testing suite can be run by `./e2e-tests.sh` this script will create a [kind](https://kind.sigs.k8s.io) cluster, execute the test suite, and delete the kind cluster.