# tobs - The Observability Stack for Kubernetes CLI

This is a CLI tool for installing and managing the The Observability Stack for Kubernetes.

## Quick Start

### Installing the CLI tool

To download and install tobs, run the following in your terminal, then follow the on-screen instructions.

```bash
curl --proto '=https' -A 'tobs' --tlsv1.2 -sSLf  https://tsdb.co/install-tobs-sh |sh
```

Alternatively, you can download the CLI directly via [our releases page](https://github.com/timescale/tobs/releases/latest)

Getting started with the CLI tool is a two-step process: First you install the CLI tool locally, then you use the CLI tool to install the tobs stack into your Kubernetes cluster.

### Using the tobs CLI tool to deploy the stack into your Kubernetes cluster

After setting up tobs run the following to install the tobs helm charts into your Kubernetes cluster

```bash
tobs install
```

This will deploy all of the tobs components into your cluster and provide instructions as to next steps.

#### Tracing support

From `0.7.0` release tobs supports installation of tracing components. To install tracing components use

```
tobs install --tracing
```

For more details on tracing support visit [Promscale tracing docs](https://github.com/timescale/promscale/blob/master/docs/tracing.md).

### Getting started by viewing your metrics in Grafana

To see your Grafana dashboards after installation run

```bash
tobs grafana get-password
tobs grafana port-forward
```

Then, point your browser to http://127.0.0.1:8080/ and login with the `admin` username.

## Usage guide

The usage guide provides a good high-level overview of what tobs CLI can do.

## Global Flags

The following are global flags that can be used with any of the commands listed below:

| Flag                | Description                                    |
|---------------------|------------------------------------------------|
| `--name`            | Helm release name                              |
| `--namespace`, `-n` | Kubernetes namespace                           |
| `--config`          | Tobs config file (default is $HOME/.tobs.yaml) |

## Commands

The following are the commands possible with the CLI.

### Base Commands

#### `tobs install`

Installs tobs observability platform. Internally uses `helm install`.

| Flag                          | Short Flag | Description                                                                                                                |
|-------------------------------|------------|----------------------------------------------------------------------------------------------------------------------------|
| `--filename`                  | `-f`       | file to load configuration from                                                                                            |
| `--chart-reference`           | `-c`       | helm chart reference (default "timescale/tobs")                                                                            |
| `--external-timescaledb-uri`  | `-e`       | external database URI, TimescaleDB installation will be skipped & Promscale connects to the provided database              |
| `--enable-prometheus-ha`      |            | option to enable prometheus and promscale high-availability, by default scales promscale to 3 replicas and prometheus to 2 |
| `--enable-timescaledb-backup` | `-b`       | option to enable TimescaleDB S3 backup                                                                                     |
| `--only-secrets`              |            | option to create only TimescaleDB secrets                                                                                  |
| `--skip-wait`                 |            | option to do not wait for pods to get into running state (useful for faster tobs installation)                             |
| `--timescaledb-tls-cert`      |            | option to provide your own tls certificate for TimescaleDB                                                                 |
| `--timescaledb-tls-key`       |            | option to provide your own tls key for TimescaleDB                                                                         |
| `--version`                   |            | option to provide tobs helm chart version, if not provided will install the latest tobs chart available                    |
| `--tracing`                   |            | option to enable tracing components                                                                                        |

#### `tobs uninstall`

Uninstalls the stack. Internally uses `helm uninstall`.

| Flag            | Short Flag | Description                               |
|-----------------|------------|-------------------------------------------|
| `--delete-data` |            | option to delete persistent volume claims |

#### `tobs upgrade`

Upgrades the stack. Internally uses `helm upgrade`.

| Flag                | Short Flag | Description                                                                                |
|---------------------|------------|--------------------------------------------------------------------------------------------|
| `--filename`        | `-f`       | file to load configuration from                                                            |
| `--chart-reference` | `-c`       | helm chart reference (default "timescale/tobs")                                            |
| `--reuse-values`    |            | native helm upgrade flag to use existing values from release                               |
| `--reset-values`    |            | native helm flag to reset values to default helm chart values                              |
| `--confirm`         | `-y`       | approve upgrade action                                                                     |
| `--same-chart`      |            | option to upgrade the helm release with latest values.yaml but the chart remains the same. |
| `--skip-crds`       |            | option to skip creating CRDs on upgrade                                                    |

#### `tobs port-forward`

Port-forwards TimescaleDB, Grafana, and Prometheus to localhost.

| Flag            | Short Flag | Description          |
|-----------------|------------|----------------------|
| `--timescaledb` | `-t`       | port for TimescaleDB |
| `--grafana`     | `-g`       | port for Grafana     |
| `--prometheus`  | `-p`       | port for Prometheus  |
| `--promscale`   | `-c`       | port for Promscale   |
| `--promlens`    | `-l`       | port for Promlens    |

#### `tobs version`

Shows the version of tobs CLI and latest helm chart

| Flag               | Short Flag | Description                                                               |
|--------------------|------------|---------------------------------------------------------------------------|
| `--deployed-chart` | `-d`       | option to show the deployed helm chart version alongside tobs CLI version |

### Helm Commands

#### `tobs helm show-values`

Documentation about Helm configuration can be found in the [Helm chart directory](/chart/README.md).

| Flag                | Short Flag | Description                                     |
|---------------------|------------|-------------------------------------------------|
| `--filename`        | `-f`       | file to load configuration from                 |
| `--chart-reference` | `-c`       | helm chart reference (default "timescale/tobs") |

### Volume Commands

The volume operation is available for TimescaleDB & Prometheus PVC's.

**Note**: To expand PVC's in Kubernetes cluster make sure you have configured `storageClass` with `allowVolumeExpansion: true` to allow PVC expansion.

#### `tobs volume get`

Displays Persistent Volume Claims sizes.

| Flag                    | Short Flag | Description |
|-------------------------|------------|-------------|
| `--timescaleDB-storage` | `-s`       |             |
| `--timescaleDB-wal`     | `-w`       |             |
| `--prometheus-storage`  | `-p`       |             |

#### `tobs volume expand`

Expands the Persistent Volume Claims for provided resources to specified sizes. The expansion size is allowed in `Ki`, `Mi` & `Gi` units. example: `150Gi`.

| Flag                    | Short Flag | Description                                    |
|-------------------------|------------|------------------------------------------------|
| `--timescaleDB-storage` | `-s`       |                                                |
| `--timescaleDB-wal`     | `-w`       |                                                |
| `--prometheus-storage`  | `-p`       |                                                |
| `--restart-pods`        | `-r`       | restart pods bound to PVC after PVC expansion. |

### TimescaleDB Commands

#### `tobs timescaledb connect`

| Flag       | Short Flag | Description                                                           |
|------------|------------|-----------------------------------------------------------------------|
| `--dbname` | `-d`       | database name to connect to, defaults to dbname from the helm release |
| `--master` | `-m`       | directly execute session on master node                               |

#### `tobs timescaledb port-forward`

| Flag     | Short Flag | Description         |
|----------|------------|---------------------|
| `--port` | `-p`       | port to listen from |

#### TimescaleDB superuser Commands

| Command                                      | Description                                                  | Flags                                                      |
|----------------------------------------------|--------------------------------------------------------------|------------------------------------------------------------|
| `tobs timescaledb superuser get-password`    | Gets the password of superuser in the Timescale database.    | None                                                       |
| `tobs timescaledb superuser change-password` | Changes the password of superuser in the Timescale database. | None                                                       |
| `tobs timescaledb superuser connect`         | Connects to the TimescaleDB database using super-user        | `--master`, `-m` : directly execute session on master node |

### Grafana Commands

| Command                        | Description                                    | Flags                                |
|--------------------------------|------------------------------------------------|--------------------------------------|
| `tobs grafana port-forward`    | Port-forwards the Grafana server to localhost. | `--port`, `-p` : port to listen from |
| `tobs grafana get-password`    | Gets the admin password for Grafana.           | None                                 |
| `tobs grafana change-password` | Changes the admin password for Grafana.        | None                                 |

### Prometheus Commands

| Command                        | Description                                       | Flags                                |
|--------------------------------|---------------------------------------------------|--------------------------------------|
| `tobs prometheus port-forward` | Port-forwards the Prometheus server to localhost. | `--port`, `-p` : port to listen from |

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

## Advanced configuration

Documentation about Helm configuration can be found in the [Helm chart directory](/chart/README.md).
Custom values.yml files can be used with the `tobs helm install -f values.yml` command.

## Building from source

**Dependencies**: [Go](https://golang.org/doc/install)

To build from source, run `go build -o tobs` from inside the `cli` folder.
Then, move the `tobs` binary from the current directory to your `/bin` folder.

## Testing

**Dependencies**: [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/), [kind](https://kind.sigs.k8s.io/)

A testing suite is included in the `tests` folder. The testing suite can be run by `./e2e-tests.sh` this script will create a [kind](https://kind.sigs.k8s.io) cluster, execute the test suite, and delete the kind cluster.
