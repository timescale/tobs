# tobs Helm Charts

A Helm chart for deploying Prometheus configured to use TimescaleDB as compressed long-term store for time-series metrics through the Promscale Connector.

### Table of contents
* **[Installing](#installing)**
  * **[Prometheus, Promscale High-Availability](./docs/ha-setup.md)**
  * **[Cleanup](#cleanup)**
    * [TimescaleDB PVCs and Backup](#timescaledb-pvcs-and-backup)
    * [TimescaleDB config service](#timescaledb-config-service)
    * [Connecting to External TimescaleDB](#timescaledb-external-integration)
* **[Configuring Helm Chart](#configuring-helm-chart)**
  * **[TimescaleDB related values](#timescaledb-related-values)**
    * [Additional configuration for TimescaleDB](#additional-configuration-for-timescaledb)
  * **[Promscale related values](#promscale-related-values)**
    * [Additional configuration for Promscale Connector](#additional-configuration-for-promscale-connector)
  * **[Prometheus related values](#prometheus-related-values)**
    * [Additional configuration for Prometheus](#additional-configuration-for-prometheus)
  * **[Grafana related values](#grafana-related-values)**
    * [Additional configuration for Grafana](#additional-configuration-for-grafana)
* **[Upgrading Helm Chart](./docs/upgrades.md)**

# Installing

The recommended way to deploy tobs is through the [CLI tool](/cli). However, we also
support helm deployments that do not use the CLI.

## Creating Secrets

By default, timescaledb helm chart doesn't create its own secrets. So please follow [these instructions][timescaledb-secrets] to create timescaledb-secrets. If you use tobs CLI this is taken care for you.

## Installing the helm chart

The following command will install Kube-Prometheus, TimescaleDB, and Promscale Connector
into your Kubernetes cluster:
```
helm repo add timescale https://charts.timescale.com/
helm repo update
helm install --devel <release_name> timescale/tobs
```

Note: The `--devel` option is needed until the chart is in beta -- helm won't find the chart withot this option.  See the helm docs for more.

## Cleanup

To uninstall a release you can run:
```
helm uninstall <release_name>
```

### TimescaleDB PVCs and Backup

Removing the deployment does not remove the Persistent Volume
Claims (pvc) belonging to the release. For a full cleanup run:
```
RELEASE=<release_name>
kubectl delete $(kubectl get pvc -l release=$RELEASE -o name)
```

If you had TimescaleDB backups enabled please check the guide for cleaning them at the [TimescaleDB Helm Chart repo][timescaledb-helm-cleanup]

### TimescaleDB config service

Sometimes one of the services created with the deployment is not deleted. The `<release_name>-config` service and `<release_name>` endpoint
may need to be manually deleted with
```
kubectl delete svc <release_name>-config
kubectl delete endpoints <release_name>
```

# Configuring Helm Chart

To get a fully-documented configuration file for `tobs`, please run:

```
helm show values --devel timescale/tobs > my_values.yml
```

You can then edit `my_values.yml` and deploy the release with the following command:

```
helm upgrade --install <release_name> --values my_values.yml timescale/tobs
```

The properties described in the tables below are only those that this chart overrides for each of the sub-charts it depends on.
You can additionally change any of the configurable properties of each sub-chart.

The chart has the following properties in the `values.yaml` file:

## TimescaleDB related values
| Parameter                                           | Description                                           | Default     |
|-----------------------------------------------------|-------------------------------------------------------|-------------|
| `timescaledb-single.enabled`                        | If false TimescaleDB will not be created              | `true`      |
| `timescaledb-single.image.tag`                      | Docker image tag to use for TimescaleDB               | `pg12-ts2.1-latest`|
| `timescaledb-single.loadBalancer.enabled`           | Create a LB for the DB instead of a ClusterIP         | `false`     |
| `timescaledb-single.replicaCount`                   | Number of pods for DB, set to 3 for HA                | `1`         |
| `timescaledb-single.backup.enabled`                 | TimescaleDB backup option by default set to false     | `false`     |

### Additional configuration for TimescaleDB

By default, the `tobs` Helm chart sets up a single-instance of TimescaleDB; if you are
interested in a replicated setup for high-availability with automated backups, please see
[this github repo][timescaledb-helm-repo] for additional instructions.

You can set up the credentials, nodeSelector, volume sizes (default volumes created are 1GB for WAL and 2GB for storage).

### Configuring External TimescaleDB

To configure tobs to connect with an external TimescaleDB you need to modify few fields in the default values.yaml while performing the installation

Below is the helm command to disable TimescaleDB installation and to set external db uri details:
```
helm install <release-name> timescale/tobs \
--set timescaledb-single.enabled=false,timescaledbExternal.enabled=true,timescaledbExternal.db_uri=<timescaledb-uri>, \
promscale.connection.uri.secretTemplate=<release-name>-timescaledb-uri
```

## Promscale related values
| Parameter                                           | Description                                           | Default     |
|-----------------------------------------------------|-------------------------------------------------------|-------------|
| `promscale.enabled`                      | If false Promscale Connector will not be created| `true` |
| `promscale.image`                        | Docker image to use for the Connector                 | `timescale/promscale:0.4.0` |
| `promscale.connection.dbName`            | Database to store the metrics in                      | `postgres`  |
| `promscale.connection.user`              | User used for connection to db | `postgres` |
| `promscale.connection.dbURI.secretTemplate` | The template for generating the name of a secret object which will hold the db URI | `` |
| `promscale.connection.password.secretTemplate` | Name (templated) of secret object containing the connection password. Key must be value of `PATRONI_SUPERUSER_PASSWORD` as this is used in TimescaleDB helm chart as reference to user `postgres`. | `"{{ .Release.Name }}-credentials"` |
| `promscale.connection.host.nameTemplate` | Host name (templated) of the database instance. Defaults to service created in `timescaledb-single` | `"{{ .Release.Name }}.{{ .Release.Namespace }}.svc.cluster.local"` |
| `promscale.service.loadBalancer.enabled` | Create a LB for the connector instead of a Cluster IP | `false`     |
| `promscale.resources.requests.memory`    | Amount of memory for the Connector pod                | `2Gi`       |
| `promscale.resources.requests.cpu`       | Number of vCPUs for the Connector pod                 | `1`         |

### Additional configuration for Promscale Connector

The connector is configured to connect to the TimescaleDB instance deployed with this chart.
But it can be configured to connect to any TimescaleDB host, and expose whichever port you like.
For more details about how to configure the Promscale connector please see the
[Helm chart directory][promscale-helm] of the [Promscale][promscale-repo] repo.

## Kube-Prometheus 

### Prometheus related values

| Parameter                                           | Description                                           | Default     |
|-----------------------------------------------------|-------------------------------------------------------|-------------|
| `kube-prometheus-stack.enabled`                     | If false, none of the Kube-Prometheus resources will be created | `true` |
| `kube-prometheus-stack.alertManager.enabled` | Enable AlertManager  |      `true` |
| `kube-prometheus-stack.alertManager.config` | AlertManager config, By default the alert manager config is from Kube-Prometheus |      `` |
| `kube-prometheus-stack.fullnameOverride`            | If false, none of the Kube-Prometheus resources will be created | `true` |
| `kube-prometheus-stack.prometheus.prometheusSpec.scrapeInterval` | Prometheus scrape interval |      `1m` |
| `kube-prometheus-stack.prometheus.prometheusSpec.scrapeTimeout` | Prometheus scrape timeout |      `10s` |
| `kube-prometheus-stack.prometheus.prometheusSpec.evaluationInterval` | Prometheus evaluation interval |    `1m` |
| `kube-prometheus-stack.prometheus.prometheusSpec.remoteRead` | Prometheus remote read config |  `url: http://{{ .Release.Name }}-promscale-connector.{{ .Release.Namespace }}.svc.cluster.local:9201/read` and `readRecent: true` |
| `kube-prometheus-stack.prometheus.prometheusSpec.remoteWrite` | Prometheus remote write config |  `url: http://{{ .Release.Name }}-promscale-connector.{{ .Release.Namespace }}.svc.cluster.local:9201/write` |
| `kube-prometheus-stack.prometheus.prometheusSpec.storageSpec.volumeClaimTemplate.spec.resources.requests.storage` | Prometheus persistent volume storage |  `8Gi` |
| `kube-prometheus-stack.prometheus.prometheusSpec.additionalScrapeConfigs` | Prometheus additional scrape config, By default additional scrape config is set scrape all pods, services and endpoint with prometheus annotations  |   |

#### Additional configuration for Prometheus

The Kube-Prometheus Community chart is used as a dependency for deploying Prometheus. We specify
the Promscale Connector as a `remote_write` and `remote_read` endpoint in the `values.yaml` that is still compatible and respects all the configuration
properties for the kube-prometheus chart, so no functionality is lost.

The Promscale connection is set using the values in `kube-prometheus-stack.prometheus.prometheusSpec.remoteWrite`.
This doesn't change the way the `kube-prometheus-stack.prometheus.prometheusSpec.remoteWrite` configuration is handled. The configuration
is separate so we can use templating and set the endpoint properly when deploying Promscale and
Prometheus in the same release. If you specify more endpoints in `prometheus.server.remoteWrite` (or `remoteRead`)
They will be added additionally.

For all the properties that can be configured and more details on how to set up the Prometheus
deployment see the [Kube Prometheus Community Chart Repo][kube-prometheus-helm-hub].

For more information about the `remote_write` configuration that can be set with
`kube-prometheus-stack.prometheus.prometheusSpec.` visit the Prometheus [Remote Write Tuning][prometheus-remote-tune] guide.

### Grafana related values

| Parameter                                                           | Description                                           | Default     |
|---------------------------------------------------------------------|-------------------------------------------------------|-------------|
| `kube-prometheus-stack.grafana.enabled`                             | If false, Grafana will not be created                 | `true`      |
| `kube-prometheus-stack.grafana.sidecar.datasources.enabled`         | If false, no data sources will be provisioned         | `true`      |
| `kube-prometheus-stack.grafana.sidecar.dashboards.enabled`          | If false, no dashboards will be provisioned by default | `true`     |
| `kube-prometheus-stack.grafana.sidecar.dashboards.files`            | Files with dashboard definitions (in JSON) to be provisioned | `['dashboards/k8s-cluster.json','dashboards/k8s-hardware.json']` |
| `kube-prometheus-stack.grafana.prometheus.datasource.enabled`       | If false, a Prometheus data source will not be provisioned | true |
| `kube-prometheus-stack.grafana.prometheus.datasource.url`           | Template parsed to the url of the Prometheus API. Defaults to Prometheus deployed with this chart  | `http://{{ .Release.Name }}-prometheus-service.{{ .Release.Namespace }}.svc.cluster.local` |
| `kube-prometheus-stack.grafana.timescale.database.enabled`          | If false, TimescaleDB will not be configured as a database, default sqllite will be used | `true` |
| `kube-prometheus-stack.grafana.timescale.database.host`             | Hostname (templated) of database, defaults to db deployed with this chart | `"{{ .Release.Name }}.{{ .Release.Namespace}}.svc.cluster.local` |
| `kube-prometheus-stack.grafana.timescale.database.user`             | User to connect to the db with (will be created ) | `grafanadb` |
| `kube-prometheus-stack.grafana.timescale.database.pass`             | Password for the user | `grafanadb` |
| `kube-prometheus-stack.grafana.timescale.database.dbName`           | Database where to store the data | `postgres` |
| `kube-prometheus-stack.grafana.timescale.database.schema`           | Schema to use (will be created) | `grafanadb` |
| `kube-prometheus-stack.grafana.timescale.database.sslMode`          | SSL mode for connection | `require` |
| `kube-prometheus-stack.grafana.timescale.datasource.host`           | Hostname (templated) of database, defaults to host deployed with this chart | `"{{ .Release.Name }}.{{ .Release.Namespace}}.svc.cluster.local` |
| `kube-prometheus-stack.grafana.timescale.datasource.enabled`        | If false a TimescaleDB data source will not be provisioned | `true` |
| `kube-prometheus-stack.grafana.timescale.datasource.user`           | User to connect with | `grafana` |
| `kube-prometheus-stack.grafana.timescale.datasource.pass`           | Pass for user | `grafana` |
| `kube-prometheus-stack.grafana.timescale.datasource.dbName`         | Database storing the metrics (Should be same with `promscale.connection.dbName`) | `postgres` |
| `kube-prometheus-stack.grafana.timescale.datasource.sslMode`        | SSL mode for connection | `require` |
| `kube-prometheus-stack.grafana.timescale.adminUser`                 | Admin user to create the users and schemas with | `postgres` |
| `kube-prometheus-stack.grafana.timescale.adminPassSecret`           | Name (templated) of secret containing password for admin user | `"{{ .Release.Name }}-credentials"` |
| `kube-prometheus-stack.grafana.adminPassword`                       | Grafana admin password, By default generates a random password | `` |

#### TimescaleDB user for the Grafana Database

This chart is configured to deploy Grafana so that it uses a TimescaleDB/PostgreSQL instance for it's database.
This is controlled with the `kube-prometheus-stack.grafana.timescale.database.enabled` value. If enabled it will run a Job that creates
a user (as specified with `kube-prometheus-stack.grafana.timescale.database.user`) and a separate schema (`kube-prometheus-stack.grafana.timescale.database.schema`).
This user is created as the owner of the schema, and will not have access to any other schemas/tables in the specified
database (`kube-prometheus-stack.grafana.timescale.database.dbName`). In order for the user and schema to be created, the `kube-prometheus-stack.grafana.timescale.adminUser`
must be set to a db user with the ability to do so (e.g. postgres), and `kube-prometheus-stack.grafana.timescale.adminPassSecret` must be
the name of a secret that contains the password for this user.

#### TimescaleDB user for a provisioned Data Source in Grafana

The chart is configured to provision a TimescaleDB data source. This is controlled with the `grafana.timescale.datasource.enabled`
If enabled it will run a Job that creates a user (as specified with `kube-prometheus-stack.grafana.timescale.datasource.user`) and grant read-only
access to the promscale schemas. In order for the user and schema to be created, the `kube-prometheus-stack.grafana.timescale.adminUser`
must be set to a db user with the ability to do so (e.g. postgres), and `kube-prometheus-stack.grafana.timescale.adminPassSecret` must be
the name of a secret that contains the password for this user.

#### Additional configuration for Grafana

The Kube-Prometheus Community chart is used as a dependency for deploying Grafana. We specify a Secret that
sets up the Prometheus Server and TimescaleDB as provisioned data sources (if they are enabled).

To get the initial password for the `admin` user after deployment execute
```
kubectl get secret --namespace <namespace> <release_name>-grafana -o jsonpath="{.data.admin-password}" | base64 --decode
```

By default Grafana is accessible on port 80 through the `<release_name>-grafana` service. You can use port-forwarding
to access it in your browser locally with

```
kubectl port-forward svc/<release_name>-grafana 8080:80
```

And then navigate to http://localhost:8080.

For all the properties that can be configured and more details on how to set up the Grafana deployment see the [Grafana Community Chart Repo][grafana-helm-hub]

[design-doc]: https://tsdb.co/prom-design-doc
[timescaledb-helm-cleanup]: https://github.com/timescale/timescaledb-kubernetes/blob/master/charts/timescaledb-single/admin-guide.md#optional-delete-the-s3-backups
[timescaledb-helm-repo]: https://github.com/timescale/timescaledb-kubernetes/tree/master/charts/timescaledb-single
[promscale-repo]: https://github.com/timescale/promscale
[promscale-helm]: https://github.com/timescale/promscale/tree/master/helm-chart
[kube-prometheus-helm-hub]: https://prometheus-community.github.io/helm-charts
[prometheus-remote-tune]: https://prometheus.io/docs/practices/remote_write/
[grafana-helm-hub]: https://grafana.github.io/helm-charts
[timescaledb-secrets]: https://github.com/timescale/timescaledb-kubernetes/blob/master/charts/timescaledb-single/admin-guide.md#creating-the-secrets