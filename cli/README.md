# tobs - The Observability Stack for Kubernetes CLI

This is a CLI tool for installing and managing the The Observability Stack for Kubernetes.

## Quick Start

### Getting started by viewing your metrics in Grafana
To see your Grafana dashboards after installation run

```bash
tobs grafana get-password
tobs grafana port-forward
```
Then, point your browser to http://127.0.0.1:8080/ and login with the `admin` username.

## Usage guide

The usage guide provides a good high-level overview of what tobs CLI can do.

## Commands

The following are the commands possible with the CLI.

### Base Commands

| Command             | Description                                                      | Flags                                                |
|---------------------|------------------------------------------------------------------|------------------------------------------------------|
| `tobs install`      | Alias for `tobs helm install`.                                   | `--filename`, `-f` : file to load configuration from <br> `--chart-reference`, `-c` : helm chart reference (default "timescale/tobs") <br>  `--external-timescaledb-uri`, `-e`: external database URI, TimescaleDB installation will be skipped & Promscale connects to the provided database <br> `--enable-prometheus-ha` : option to enable prometheus and promscale high-availability, by default scales promscale to 3 replicas and prometheus to 2<br> `--enable-timescaledb-backup`, `-b` : option to enable TimescaleDB S3 backup <br> `--only-secrets` :  option to create only TimescaleDB secrets <br> `--skip-wait` : option to do not wait for pods to get into running state (useful for faster tobs installation) <br> `--timescaledb-tls-cert` : option to provide your own tls certificate for TimescaleDB <br> `--timescaledb-tls-key` : option to provide your own tls key for TimescaleDB <br> `--version` : option to provide tobs helm chart version, if not provided will install the latest tobs chart available <br> `--tracing` : option to enable tracing components   |
| `tobs uninstall`    | Alias for `tobs helm unintall`.                                  | `--delete-data`: option to delete persistent volume claims |
| `tobs port-forward` | Port-forwards TimescaleDB, Grafana, and Prometheus to localhost. | `--timescaledb`, `-t` : port for TimescaleDB <br> `--grafana`, `-g` : port for Grafana <br> `--prometheus`, `-p` : port for Prometheus <br> `--promscale`, `-c` : port for Promscale <br> `--promlens`, `-l` : port for Promlens |
| `tobs version`      | Shows the version of tobs CLI and latest helm chart              | `--deployed-chart`, `-d` : option to show the deployed helm chart version alongside tobs CLI version   |

### Helm Commands

Documentation about Helm configuration can be found in the [Helm chart directory](/chart/README.md).

| Command                 | Description                                                                  | Flags                                                |
|-------------------------|------------------------------------------------------------------------------|------------------------------------------------------|
| `tobs helm show-values` | Prints the YAML configuration of the Helm chart for The Observability Stack. | `--filename`, `-f` : file to load configuration from <br> `--chart-reference`, `-c` : helm chart reference (default "timescale/tobs") |

### TimescaleDB Commands

| Command                            | Description                                                | Flags                                       |
|------------------------------------|------------------------------------------------------------|---------------------------------------------|
| `tobs timescaledb connect`         | Connects to the Timescale database with the provided user. | `--dbname`, `-d` : database name to connect to, defaults to dbname from the helm release <br> `--master`, `-m` : directly execute session on master node |
| `tobs timescaledb port-forward`    | Port-forwards TimescaleDB to localhost.                    | `--port`, `-p` : port to listen from        |

**TimescaleDB superuser Commands**

| Command                                       | Description                                                  | Flags                                                       |
|-----------------------------------------------|--------------------------------------------------------------|-------------------------------------------------------------|
| `tobs timescaledb superuser get-password`     | Gets the password of superuser in the Timescale database.    | None                                                        |
| `tobs timescaledb superuser change-password`  | Changes the password of superuser in the Timescale database. | None                                                        |
| `tobs timescaledb superuser connect`          | Connects to the TimescaleDB database using super-user        | `--master`, `-m` : directly execute session on master node  |

### Grafana Commands

| Command                             | Description                                    | Flags                                |
|-------------------------------------|------------------------------------------------|--------------------------------------|
| `tobs grafana port-forward`         | Port-forwards the Grafana server to localhost. | `--port`, `-p` : port to listen from |
| `tobs grafana get-password`         | Gets the admin password for Grafana.           | None                                 |
| `tobs grafana change-password`      | Changes the admin password for Grafana.        | None                                 |

### Prometheus Commands

| Command                        | Description                                       | Flags                                |
|--------------------------------|---------------------------------------------------|--------------------------------------|
| `tobs prometheus port-forward` | Port-forwards the Prometheus server to localhost. | `--port`, `-p` : port to listen from |

### Jaeger Commands

Jaeger cmds are only supported if tracing is enabled in tobs installation

| Command                        | Description                                       | Flags                                |
|--------------------------------|---------------------------------------------------|--------------------------------------|
| `tobs jaeger port-forward`     | Port-forwards the jaeger query to localhost.      | `--port`, `-p` : port to listen from |

### Metrics Commands

| Command                                   | Description                                                                          | Flags |
|-------------------------------------------|--------------------------------------------------------------------------------------|-------|
| `tobs metrics retention get`              | Gets the data retention period of a specific metric.                                 | None  |
| `tobs metrics retention set-default`      | Sets the default data retention period to the specified number of days.              | None  |
| `tobs metrics retention set`              | Sets the data retention period of a specific metric to the specified number of days. | None  |
| `tobs metrics retention reset`            | Resets the data retention period of a specific metric to the default value.          | None  |
| `tobs metrics chunk-interval get`         | Gets the chunk interval of a specific metric.                                        | None  |
| `tobs metrics chunk-interval set-default` | Sets the default chunk interval to the specified duration.                           | None  |
| `tobs metrics chunk-interval set`         | Sets the chunk interval of a specific metric to the specified duration.              | None  |
| `tobs metrics chunk-interval reset`       | Resets chunk interval of a specific metric to the default value.                     | None  |

### Volume Commands

The volume operation is available for TimescaleDB & Prometheus PVC's.

**Note**: To expand PVC's in Kubernetes cluster make sure you have configured `storageClass` with `allowVolumeExpansion: true` to allow PVC expansion.

| Command                        | Description                                       | Flags                                |
|--------------------------------|---------------------------------------------------|--------------------------------------|
| `tobs volume get`     | Displays Persistent Volume Claims sizes. | `--timescaleDB-storage`, `s`, `--timescaleDB-wal`, `w`, `prometheus-storage`, `-p`  |
| `tobs volume expand`  | Expands the Persistent Volume Claims for provided resources to specified sizes. The expansion size is allowed in `Ki`, `Mi` & `Gi` units. example: `150Gi`. | `--timescaleDB-storage`, `s`, `--timescaleDB-wal`, `w`, `prometheus-storage`, `-p`, `--restart-pods`, `-r` to restart pods bound to PVC after PVC expansion. |

### Upgrade Command

The upgrade cmd helps to upgrade the existing tobs deployment. You can upgrade the tobs to latest helm chart provided the helm chart is released to timescale helm repository. 
You can also upgrade your existing tobs deployment to latest `values.yaml` configuration. This internally uses the `helm upgrade` utility.

| Command                        | Description                                       | Flags                                |
|--------------------------------|---------------------------------------------------|--------------------------------------|
| `tobs upgrade`     | Upgrades the tobs deployment if new helm chart is available. Also, upgrades tobs if updated `values.yaml` is provided. | `--filename`, `-f` : file to load configuration from <br> `--chart-reference`, `-c` : helm chart reference (default "timescale/tobs") <br> `--reuse-values` : native helm upgrade flag to use existing values from release <br> `--reset-values` : native helm flag to reset values to default helm chart values <br> `--confirm`, `-y` : approve upgrade action <br> `--same-chart` : option to upgrade the helm release with latest values.yaml but the chart remains the same. <br> `--skip-crds` : option to skip creating CRDs on upgrade  |

## Global Flags

The following are global flags that can be used with any of the above commands:

| Flag                 | Description                                    |
|----------------------|------------------------------------------------|
| `--name`             | Helm release name                              |
| `--namespace`, `-n`  | Kubernetes namespace                           |
| `--config`           | Tobs config file (default is $HOME/.tobs.yaml) |

## Advanced configuration

Documentation about Helm configuration can be found in the [Helm chart directory](/chart/README.md).
Custom values.yml files can be used with the `tobs helm install -f values.yml` command.

## Building from source

__Dependencies__: [Go](https://golang.org/doc/install)

To build from source, run `go build -o tobs` from inside the `cli` folder.
Then, move the `tobs` binary from the current directory to your `/bin` folder.

## Testing

__Dependencies__: [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/), [kind](https://kind.sigs.k8s.io/)

A testing suite is included in the `tests` folder. The testing suite can be run by `./e2e-tests.sh` this script will create a [kind](https://kind.sigs.k8s.io) cluster, execute the test suite, and delete the kind cluster.
