# tobs Helm Charts

A Helm chart for deploying Prometheus configured to use TimescaleDB as compressed
long-term store for time-series metrics through the Promscale.

## Table of contents

- **[Install](#install)**
  - [Installing the helm chart](#installing-the-helm-chart)
- **[Configuring Helm Chart](#configuring-helm-chart)**
  - **[TimescaleDB](#timescaledb)**
    - [Configuring an external TimescaleDB](#configuring-an-external-timescaledb)
    - [Additional configuration for TimescaleDB](#additional-configuration-for-timescaledb)
  - **[Promscale](#promscale)**
    - [Additional configuration for Promscale](#additional-configuration-for-promscale)
    - [Prometheus, Promscale High-Availability](../docs/values-config.md#prometheus-high-availability)
  - **[Prometheus](#prometheus)**
    - [Additional configuration for Prometheus](#additional-configuration-for-prometheus)
  - **[Grafana](#grafana)**
    - [Additional configuration for Grafana](#additional-configuration-for-grafana)
- **[Upgrading Helm Chart](../docs/upgrades.md)**
- **[Uninstall](#uninstall)**
  - [TimescaleDB secrets](#timescaledb-secrets)
  - [Kube-Prometheus secret](#kube-prometheus-secret)
  - [TimescaleDB PVCs and Backup](#timescaledb-pvcs-and-backup)
  - [Prometheus PVCs](#prometheus-pvcs)

## Install

### Prerequisites

Using tobs to install full observability stack with openTelemetry support
currently requires installation of cert-manager. To do install it please follow
[cert-manager documentation](https://cert-manager.io/docs/installation/).

_Note_: cert-manager is not required when using tobs with opentelemetry support disabled.

### Installing the helm chart

The following command will install Kube-Prometheus, TimescaleDB, OpenTelemetry
Operator, and Promscale into your Kubernetes cluster **It is recommended you
install tobs into its own Namespace**:

```shell
RELEASE=<release name>
NAMESPACE=<namespace>

kubectl create ns $NAMESPACE
helm repo add timescale https://charts.timescale.com/
helm repo update
helm install --wait --timeout 15m $RELEASE timescale/tobs -n $NAMESPACE
```

_Note_: `--wait` flag is necessary for successfull installation as tobs helm
chart can create opentelemetry Custom Resources only after opentelemetry-operator
is up and running. This flag can be omited when using tobs without opentelemetry
support.

## Uninstall

Due to some quirkiness with Helm, if you wish to uninstall tobs you will need to
follow these steps.

To uninstall a release you can run:

```shell
RELEASE=<release name>
NAMESPACE=<namespace>
helm uninstall $RELEASE -n $NAMESPACE
```

After uninstalling helm release some objects will be left over. To remove them
follow next sections.

### TimescaleDB secrets

TimescaleDB secret's created with the deployment aren't deleted. These secrets
need to be manually deleted:

```shell
RELEASE=<release_name>
NAMESPACE=<namespace>
kubectl delete -n $NAMESPACE $(kubectl get secrets -n $NAMESPACE -l "app=$RELEASE-timescaledb" -o name)
```

### Promscale configmap

Promscale has a configmap that is created that isn't deleted

```shell
RELEASE=<release_name>
NAMESPACE=<namespace>
kubectl delete -n $NAMESPACE $(kubectl get configmap -n $NAMESPACE -l "app=$RELEASE-promscale" -o name)
```

### tobs secrets

tobs installs various secrets and configmaps that need to be cleaned up as well.

```shell
RELEASE=<release_name>
NAMESPACE=<namespace>
kubectl delete -n $NAMESPACE $(kubectl get secrets -n $NAMESPACE -l "app=$RELEASE-tobs" -o name)
```

### Kube-Prometheus secret

One of the Kube-Prometheus secrets created with the deployment isn't deleted.
This secret needs to be manually deleted:

```shell
RELEASE=<release_name>
NAMESPACE=<namespace>
kubectl delete secret -n $NAMESPACE $RELEASE-kube-prometheus-stack-admission
```

### TimescaleDB PVCs and Backup

Removing the deployment does not remove the Persistent Volume
Claims (pvc) belonging to the release. For a full cleanup run:

```shell
RELEASE=<release_name>
NAMESPACE=<namespace>
kubectl delete -n $NAMESPACE $(kubectl get pvc -n $NAMESPACE -l release=$RELEASE -o name)
```

If you had TimescaleDB backups enabled please check the guide for cleaning them
at the [TimescaleDB Helm Chart repo](https://github.com/timescale/timescaledb-kubernetes/blob/master/charts/timescaledb-single/admin-guide.md#optional-delete-the-s3-backups)

### Prometheus PVCs

Removing the deployment does not remove the Persistent Volume
Claims (pvc) of Prometheus belonging to the release. For a full cleanup run:

```shell
RELEASE=<release_name>
NAMESPACE=<namespace>
kubectl delete -n $NAMESPACE $(kubectl get pvc -n $NAMESPACE -l operator.prometheus.io/name=$RELEASE-kube-prometheus-stack-prometheus -o name)
```

### Opentelemetry Collector

Removing the deployment does not remove the `OpentelemetryCollector` CR object
For a full cleanup run:

```shell
NAMESPACE=<namespace>
kubectl delete -n $NAMESPACE $(kubectl get opentelemetrycollectors -n $NAMESPACE -l app.kubernetes.io/managed-by=opentelemetry-operator -o name)
kubectl delete secret -n $NAMESPACE opentelemetry-operator-controller-manager-service-cert
```

### Delete Namespace

Since it was recommended for you to install tobs into its own specific namespace
you can go ahead and remove that as well.

```shell
NAMESPACE=<namespace>
kubectl delete ns $NAMESPACE
```

## Configuring Helm Chart

To get a fully-documented configuration file for `tobs`, please run:

```shell
helm show values timescale/tobs > my_values.yml
```

You can then edit `my_values.yml` and deploy the release with the following command:

```shell
helm upgrade --wait --install <release_name> --values my_values.yml timescale/tobs
```

The properties described in the tables below are only those that this chart overrides
for each of the sub-charts it depends on. You can additionally change any of the
configurable properties of each sub-chart.

The chart has the following properties in the `values.yaml` file:

## TimescaleDB

| Parameter                                        | Description                                       | Default             |
| ------------------------------------------------ | ------------------------------------------------- | ------------------- |
| `timescaledb-single.enabled`                     | If false TimescaleDB will not be created          | `true`              |
| `timescaledb-single.image.tag`                   | Docker image tag to use for TimescaleDB           | `pg14.4-ts2.7.2-p0` |
| `timescaledb-single.replicaCount`                | Number of pods for DB, set to 3 for HA            | `1`                 |
| `timescaledb-single.backup.enabled`              | TimescaleDB backup option by default set to false | `false`             |
| `timescaledb-single.persistentVolumes.data.size` | Size of the volume for the database               | `150Gi`             |
| `timescaledb-single.persistentVolumes.wal.size`  | Size of the volume for the WAL disk               | `20Gi`              |
| `resources.requests.cpu`                         | Resource request for cpu                          | `100m`              |
| `resources.requests.memory`                      | Resource request for memory                       | `2Gi`               |

### Additional configuration for TimescaleDB

By default, the `tobs` Helm chart sets up a single-instance of TimescaleDB; if
you are interested in a replicated setup for high-availability with automated backups,
please see [this github repo](https://github.com/timescale/timescaledb-kubernetes/tree/master/charts/timescaledb-single)
for additional instructions.

You can set up the credentials, nodeSelector, volume sizes (default volumes
created are 1GB for WAL and 2GB for storage).

### Configuring an external TimescaleDB

To configure tobs to connect with an external TimescaleDB you need to modify a
few fields in the default values.yaml while performing the installation

Below is the helm command to disable the TimescaleDB installation and set
external db uri details:

```shell
helm install --wait <release-name> timescale/tobs \
--set timescaledb-single.enabled=false,promscale.connection.uri=<timescaledb-uri>
```

## Promscale

| Parameter                             | Description                                                                                   | Default                                              |
| ------------------------------------- | --------------------------------------------------------------------------------------------- | ---------------------------------------------------- |
| `promscale.enabled`                   | If false Promscale will not be started                                                        | `true`                                               |
| `promscale.image`                     | Docker image to use for the Promscale                                                         | `timescale/promscale:0.8.0`                          |
| `promscale.connection.dbName`         | Database to store the metrics in                                                              | `postgres`                                           |
| `promscale.connection.user`           | User used for connection to db                                                                | `postgres`                                           |
| `promscale.connection.uri`            | TimescaleDB URI                                                                               | ``                                                   |
| `promscale.connection.password`       | Assign the TimescaleDB password from `tobs-credentials` from key `PATRONI_SUPERUSER_PASSWORD` | ``                                                   |
| `promscale.connection.host`           | TimescaleDB host address                                                                      | `"{{ .Release.Name }}.{{ .Release.Namespace }}.svc"` |
| `promscale.service.type`              | Configure the service type for Promscale                                                      | `ClusterIP`                                          |
| `promscale.resources.requests.memory` | Amount of memory for the Promscale pod                                                        | `2Gi`                                                |
| `promscale.resources.requests.cpu`    | Number of vCPUs for the Promscale pod                                                         | `1`                                                  |

### Additional configuration for Promscale

The Promscale is configured to connect to the TimescaleDB instance deployed with
this chart. But it can be configured to connect to [any TimescaleDB host](#configuring-an-external-timescaledb),
and expose whichever port you like. For more details about how to configure the
Promscale please see the [Helm chart directory](https://github.com/timescale/promscale/tree/master/deploy/helm-chart)
of the [Promscale](https://github.com/timescale/promscale) repo.

## Kube-Prometheus

### Prometheus

| Parameter                                                                                                         | Description                                                                                                                                        | Default                                                                          |
| ----------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------- |
| `kube-prometheus-stack.enabled`                                                                                   | If false, none of the Kube-Prometheus resources will be created                                                                                    | `true`                                                                           |
| `kube-prometheus-stack.alertManager.enabled`                                                                      | Enable AlertManager                                                                                                                                | `true`                                                                           |
| `kube-prometheus-stack.alertManager.config`                                                                       | AlertManager config, By default the alert manager config is from Kube-Prometheus                                                                   | ``                                                                               |
| `kube-prometheus-stack.fullnameOverride`                                                                          | If false, none of the Kube-Prometheus resources will be created                                                                                    | `true`                                                                           |
| `kube-prometheus-stack.prometheus.prometheusSpec.scrapeInterval`                                                  | Prometheus scrape interval                                                                                                                         | `1m`                                                                             |
| `kube-prometheus-stack.prometheus.prometheusSpec.scrapeTimeout`                                                   | Prometheus scrape timeout                                                                                                                          | `10s`                                                                            |
| `kube-prometheus-stack.prometheus.prometheusSpec.evaluationInterval`                                              | Prometheus evaluation interval                                                                                                                     | `1m`                                                                             |
| `kube-prometheus-stack.prometheus.prometheusSpec.retention`                                                       | Prometheus data retention                                                                                                                          | `1d`                                                                             |
| `kube-prometheus-stack.prometheus.prometheusSpec.remoteWrite[0].queueConfig.batchSendDeadline`                    | BatchSendDeadline is the maximum time a sample will wait in buffer.                                                                                | `"30s"`                                                                          |
| `kube-prometheus-stack.prometheus.prometheusSpec.remoteWrite[0].queueConfig.capacity`                             | Capacity is the number of samples to buffer per shard before we start dropping them.                                                               | `100000`                                                                         |
| `kube-prometheus-stack.prometheus.prometheusSpec.remoteWrite[0].queueConfig.maxBackoff`                           | MaxBackoff is the maximum retry delay.                                                                                                             | `"10s"`                                                                          |
| `kube-prometheus-stack.prometheus.prometheusSpec.remoteWrite[0].queueConfig.maxSamplesPerSecond`                  | MaxSamplesPerSend is the maximum number of samples per send.                                                                                       | `10000`                                                                          |
| `kube-prometheus-stack.prometheus.prometheusSpec.remoteWrite[0].queueConfig.maxShards`                            | MaxShards is the maximum number of shards, i.e. amount of concurrency.                                                                             | `20`                                                                             |
| `kube-prometheus-stack.prometheus.prometheusSpec.remoteWrite[0].queueConfig.minBackoff`                           | MinBackoff is the initial retry delay. Gets doubled for every retry.                                                                               | `"100ms"`                                                                        |
| `kube-prometheus-stack.prometheus.prometheusSpec.remoteWrite[0].queueConfig.minShards`                            | MinShards is the minimum number of shards, i.e. amount of concurrency.                                                                             | `20`                                                                             |
| `kube-prometheus-stack.prometheus.prometheusSpec.remoteWrite[0].remoteTimeout`                                    | Timeout for requests to the remote write endpoint.                                                                                                 | `"100s"`                                                                         |
| `kube-prometheus-stack.prometheus.prometheusSpec.remoteWrite[0].url`                                              | The Prometheus URL of the endpoint to send samples to.                                                                                             | `"http://{{ .Release.Name }}-promscale.{{ .Release.Namespace }}.svc:9201/write"` |
| `kube-prometheus-stack.prometheus.prometheusSpec.storageSpec.volumeClaimTemplate.spec.resources.requests.storage` | Prometheus persistent volume storage                                                                                                               | `8Gi`                                                                            |
| `kube-prometheus-stack.prometheus.prometheusSpec.additionalScrapeConfigs`                                         | Prometheus additional scrape config, By default additional scrape config is set scrape all pods, services and endpoint with prometheus annotations |                                                                                  |

#### Additional configuration for Prometheus

The Kube-Prometheus Community chart is used as a dependency for deploying
Prometheus. We specify Promscale as a `remote_write` and `remote_read` endpoint
in the `values.yaml` that is still compatible and respects all the configuration
properties for the kube-prometheus chart, so no functionality is lost.

The Promscale connection is set using the values in `kube-prometheus-stack.prometheus.prometheusSpec.remoteWrite`.
This doesn't change the way the `kube-prometheus-stack.prometheus.prometheusSpec.remoteWrite`
configuration is handled. The configuration is separate so we can use templating
and set the endpoint properly when deploying Promscale and Prometheus in the same
release. If you specify more endpoints in `prometheus.server.remoteWrite` (or `remoteRead`)
They will be added additionally.

For all the properties that can be configured and more details on how to set up
the Prometheus deployment see the [Kube Prometheus Community Chart Repo](https://prometheus-community.github.io/helm-charts).

For more information about the `remote_write` configuration that can be set with
`kube-prometheus-stack.prometheus.prometheusSpec.` visit the Prometheus
[Remote Write Tuning](https://prometheus.io/docs/practices/remote_write/) guide.

This Helm chart utilizes our recommended Prometheus `remote-write` [configuration](https://docs.timescale.com/promscale/latest/recommendations/resource-recomm/#prometheus-remote-write) by default

### Grafana

| Parameter                                                     | Description                                                                                       | Default                                                                      |
| ------------------------------------------------------------- | ------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------- |
| `kube-prometheus-stack.grafana.enabled`                       | If false, Grafana will not be created                                                             | `true`                                                                       |
| `kube-prometheus-stack.grafana.sidecar.datasources.enabled`   | If false, no data sources will be provisioned                                                     | `true`                                                                       |
| `kube-prometheus-stack.grafana.sidecar.dashboards.enabled`    | If false, no dashboards will be provisioned by default                                            | `true`                                                                       |
| `kube-prometheus-stack.grafana.sidecar.dashboards.files`      | Files with dashboard definitions (in JSON) to be provisioned                                      | `['dashboards/k8s-cluster.json','dashboards/k8s-hardware.json']`             |
| `kube-prometheus-stack.grafana.prometheus.datasource.enabled` | If false, a Prometheus data source will not be provisioned                                        | true                                                                         |
| `kube-prometheus-stack.grafana.prometheus.datasource.url`     | Template parsed to the url of the Prometheus API. Defaults to Prometheus deployed with this chart | `http://{{ .Release.Name }}-prometheus-service.{{ .Release.Namespace }}.svc` |
| `kube-prometheus-stack.grafana.timescale.datasource.host`     | Hostname (templated) of database, defaults to host deployed with this chart                       | `"{{ .Release.Name }}.{{ .Release.Namespace}}.svc`                           |
| `kube-prometheus-stack.grafana.timescale.datasource.enabled`  | If false a TimescaleDB data source will not be provisioned                                        | `true`                                                                       |
| `kube-prometheus-stack.grafana.timescale.datasource.user`     | User to connect with                                                                              | `grafana`                                                                    |
| `kube-prometheus-stack.grafana.timescale.datasource.pass`     | Pass for user                                                                                     | `grafana`                                                                    |
| `kube-prometheus-stack.grafana.timescale.datasource.dbName`   | Database storing the metrics (Should be same with `promscale.connection.dbName`)                  | `postgres`                                                                   |
| `kube-prometheus-stack.grafana.timescale.datasource.sslMode`  | SSL mode for connection                                                                           | `require`                                                                    |
| `kube-prometheus-stack.grafana.adminPassword`                 | Grafana admin password, By default generates a random password                                    | ``                                                                           |

#### TimescaleDB user for a provisioned Data Source in Grafana

The chart is configured to provision a TimescaleDB data source. This is controlled
with the `grafana.timescale.datasource.enabled` If enabled it will add
timescaleDB SQL initialization script that creates a user (as specified with
`kube-prometheus-stack.grafana.timescale.datasource.user`) and grant read-only
access to the promscale schemas.

_Note: For security reasons this feature works only with TimescaleDB provisioned
with tobs. For external DB you need to provision that user and password by
yourself using instructions from [../docs/upgrades.md#SQL-Datasource-credential-handling-improvements](docs/upgrades.md#sql-datasource-credential-handling-improvements)_

#### Additional configuration for Grafana

The Kube-Prometheus Community chart is used as a dependency for deploying
Grafana. We specify a Secret that sets up the Prometheus Server and TimescaleDB
as provisioned data sources (if they are enabled).

To get the initial password for the `admin` user after deployment run the following
command

```shell
kubectl get secret --namespace <namespace> <release_name>-grafana -o jsonpath="{.data.admin-password}" | base64 --decode
```

By default Grafana is accessible on port 80 through the `<release_name>-grafana` service. You can use port-forwarding
to access it in your browser locally with

```shell
kubectl port-forward svc/<release_name>-grafana 8080:80
```

And then navigate to [http://localhost:8080](http://localhost:8080).

For all the properties that can be configured and more details on how to set up
the Grafana deployment see the [Grafana Community Chart Repo](https://grafana.github.io/helm-charts)
