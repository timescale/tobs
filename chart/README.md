# tobs Helm Charts

A Helm chart for deploying Prometheus configured to use TimescaleDB as compressed long-term store for time-series metrics through the Promscale Connector.

### Table of contents
* **[Installing](#installing)**
  * **[Cleanup](#cleanup)**
    * [TimescaleDB PVCs and Backup](#timescaledb-pvcs-and-backup)
    * [TimescaleDB config service](#timescaledb-config-service)
* **[Configuring Helm Chart](#configuring-helm-chart)**
  * **[TimescaleDB related values](#timescaledb-related-values)**
    * [Additional configuration for TimescaleDB](#additional-configuration-for-timescaledb)
  * **[Promscale related values](#promscale-related-values)**
    * [Additional configuration for Promscale Connector](#additional-configuration-for-promscale-connector)
  * **[Prometheus related values](#prometheus-related-values)**
    * [Additional configuration for Prometheus](#additional-configuration-for-prometheus)
  * **[Grafana related values](#grafana-related-values)**
    * [Additional configuration for Grafana](#additional-configuration-for-grafana)

# Installing

The recommended way to deploy tobs is through the [CLI tool](/cli). However, we also
support helm deployments that do not use the CLI.

The following command will install Prometheus, TimescaleDB, Promscale Connector, and Grafana
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
| `timescaledb-single.image.tag`                      | Docker image tag to use for TimescaleDB               | `pg12-ts1.7`|
| `timescaledb-single.loadBalancer.enabled`           | Create a LB for the DB instead of a ClusterIP         | `false`     |
| `timescaledb-single.replicaCount`                   | Number of pods for DB, set to 3 for HA                | `1`         |

### Additional configuration for TimescaleDB

By default, the `tobs` Helm chart sets up a single-instance of TimescaleDB; if you are
interested in a replicated setup for high-availability with automated backups, please see
[this github repo][timescaledb-helm-repo] for additional instructions.

You can set up the credentials, nodeSelector, volume sizes (default volumes created are 1GB for WAL and 2GB for storage).

## Promscale related values
| Parameter                                           | Description                                           | Default     |
|-----------------------------------------------------|-------------------------------------------------------|-------------|
| `promscale.enabled`                      | If false Promscale Connector will not be created| `true` |
| `promscale.image`                        | Docker image to use for the Connector                 | `timescale/promscale:0.1.0-alpha.2` |
| `promscale.connection.dbName`            | Database to store the metrics in                      | `postgres`  |
| `promscale.connection.user`              | User used for connection to db | `postgres` |
| `promscale.connection.password.secretTemplate` | Name (templated) of secret object containing the connection password. Key must be value of `promscale.connection.user`. Defaults to secret created by timescaledb-single chart | `"{{ .Release.Name }}-timescaledb-passwords"` |
| `promscale.connection.host.nameTemplate` | Host name (templated) of the database instance. Defaults to service created in `timescaledb-single` | `"{{ .Release.Name }}.{{ .Release.Namespace }}.svc.cluster.local"` |
| `promscale.service.loadBalancer.enabled` | Create a LB for the connector instead of a Cluster IP | `false`     |
| `promscale.resources.requests.memory`    | Amount of memory for the Connector pod                | `2Gi`       |
| `promscale.resources.requests.cpu`       | Number of vCPUs for the Connector pod                 | `1`         |

### Additional configuration for Promscale Connector

The connector is configured to connect to the TimescaleDB instance deployed with this chart.
But it can be configured to connect to any TimescaleDB host, and expose whichever port you like.
For more details about how to configure the Promscale connector please see the
[Helm chart directory][promscale-helm] of the [Promscale][promscale-repo] repo.

## Prometheus related values

| Parameter                                           | Description                                           | Default     |
|-----------------------------------------------------|-------------------------------------------------------|-------------|
| `prometheus.enabled`                                | If false, none of the Prometheus resources will be created | `true` |
| `prometheus.alertmanager.enabled`                   | If true will create the Prometheus Alert Manager       | `false`    |
| `prometheus.pushgateway.enabled`                    | If true will create the Prometheus Push Gateway        | `false`    |
| `prometheus.server.configMapOverrideName`           | The name of the ConfigMap that provides the Prometheus config. Resolves to `{{ .Release.Name }}-{{ .Values.prometheus.server.configMapOverrideName }}` | `prometheus-config` |
| `prometheus.server.timescaleRemote.host` | Templated hostname of Promscale connector to be used as Long Term Storage | `{{ .Release.Name }}-promscale.{{ .Release.Namespace }}.svc.cluster.local` |
| `prometheus.server.timescaleRemote.protocol` | Protocol to use to send the metrics to Promscale | `http` |
| `prometheus.server.timescaleRemote.port` | Listening Port of Promscale Connector | `9201` |
| `prometheus.server.timescaleRemote.write.enabled` | If false, Promscale Connector will not be set up as remote_write| `true` |
| `prometheus.server.timescaleRemote.write.endpoint` | Write endpoint of Promscale. Used to generate url of remote_write as {protocol}://{host}:{port}/{endpoint} | `write` |
| `prometheus.server.timescaleRemote.write.queue` | remote_write queue config | `{max_shards: 30}`
| `prometheus.server.timescaleRemote.read.enabled` | If false Promscale Connector will not be set up as remote_read | `true` |
| `prometheus.server.timescaleRemote.write.endpoint` | Read endpoint of Promscale. Used to generate url of remote_read as {protocol}://{host}:{port}/{endpoint} | `read` |

### Additional configuration for Prometheus

The Prometheus Community chart is used as a dependency for deploying Prometheus. We specify
a ConfigMap override where the Promscale Connector is already configured as a `remote_write`
and `remote_read` endpoint. We create a ConfigMap that is still compatible and respects all the configuration
properties for the prometheus chart, so no functionality is lost.

The Promscale connection is set using the values in `prometheus.server.timescaleRemote`.
This doesn't change the way the `prometheus.server.remoteWrite` configuration is handled. The configuration
is separate so we can use templating and set the endpoint properly when deploying Promscale and
Prometheus in the same release. If you specify more endpoints in `prometheus.server.remoteWrite` (or `remoteRead`)
They will be added additionally.

For all the properties that can be configured and more details on how to set up the Prometheus
deployment see the [Prometheus Community Chart Repo][prometheus-helm-hub].

For more information about the `remote_write` configuration that can be set with
`promscale.remote.queue` visit the Prometheus [Remote Write Tuning][prometheus-remote-tune] guide.

## Grafana related values

| Parameter                                           | Description                                           | Default     |
|-----------------------------------------------------|-------------------------------------------------------|-------------|
| `grafana.enabled`                                   | If false, Grafana will not be created                 | `true`      |
| `grafana.sidecar.datasources.enabled`               | If false, no data sources will be provisioned         | `true`      |
| `grafana.sidecar.dashboards.enabled`                | If false, no dashboards will be provisioned by default | `true`     |
| `grafana.sidecar.dashboards.files`                  | Files with dashboard definitions (in JSON) to be provisioned | `['dashboards/k8s-cluster.json','dashboards/k8s-hardware.json']` |
| `grafana.prometheus.datasource.enabled` | If false, a Prometheus data source will not be provisioned | true |
| `grafana.prometheus.datasource.url` | Template parsed to the url of the Prometheus API. Defaults to Prometheus deployed with this chart  | `http://{{ .Release.Name }}-prometheus-service.{{ .Release.Namespace }}.svc.cluster.local` |
| `grafana.timescale.database.enabled` | If false, TimescaleDB will not be configured as a database, default sqllite will be used | `true` |
| `grafana.timescale.database.host` | Hostname (templated) of database, defaults to db deployed with this chart | `"{{ .Release.Name }}.{{ .Release.Namespace}}.svc.cluster.local` |
| `grafana.timescale.database.user`                     | User to connect to the db with (will be created ) | `grafanadb` |
| `grafana.timescale.database.pass`                     | Password for the user | `grafanadb` |
| `grafana.timescale.database.dbName`                   | Database where to store the data | `postgres` |
| `grafana.timescale.database.schema`                   | Schema to use (will be created) | `grafanadb` |
| `grafana.timescale.database.sslMode`                  | SSL mode for connection | `require` |
| `grafana.timescale.datasource.host` | Hostname (templated) of database, defaults to host deployed with this chart | `"{{ .Release.Name }}.{{ .Release.Namespace}}.svc.cluster.local` |
| `grafana.timescale.datasource.enabled` |  If false a TimescaleDB data source will not be provisioned | `true` |
| `grafana.timescale.datasource.user` | User to connect with | `grafana` |
| `grafana.timescale.datasource.pass` | Pass for user | `grafana` |
| `grafana.timescale.datasource.dbName` | Database storing the metrics (Should be same with `promscale.connection.dbName`) | `postgres` |
| `grafana.timescale.datasource.sslMode` | SSL mode for connection | `require` |
| `grafana.timescale.adminUser`                | Admin user to create the users and schemas with | `postgres` |
| `grafana.timescale.adminPassSecret`      | Name (templated) of secret containing password for admin user | `"{{ .Release.Name }}-timescaledb-passwords"` |

### TimescaleDB user for the Grafana Database

This chart is configured to deploy Grafana so that it uses a TimescaleDB/PostgreSQL instance for it's database.
This is controlled with the `grafana.timescale.database.enabled` value. If enabled it will run a Job that creates
a user (as specified with `grafana.timescale.database.user`) and a separate schema (`grafana.timescale.database.schema`).
This user is created as the owner of the schema, and will not have access to any other schemas/tables in the specified
database (`grafana.timescale.database.dbName`). In order for the user and schema to be created, the `grafana.timescale.adminUser`
must be set to a db user with the ability to do so (e.g. postgres), and `grafana.timescale.adminPassSecret` must be
the name of a secret that contains the password for this user.

### TimescaleDB user for a provisioned Data Source in Grafana

The chart is configured to provision a TimescaleDB data source. This is controlled with the `grafana.timescale.datasource.enabled`
If enabled it will run a Job that creates a user (as specified with `grafana.timescale.datasource.user`) and grant read-only
access to the promscale schemas. In order for the user and schema to be created, the `grafana.timescale.adminUser`
must be set to a db user with the ability to do so (e.g. postgres), and `grafana.timescale.adminPassSecret` must be
the name of a secret that contains the password for this user.

### Additional configuration for Grafana

The Grafana Community chart is used as a dependency for deploying Grafana. We specify a Secret that
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
[prometheus-helm-hub]: https://prometheus-community.github.io/helm-charts
[prometheus-remote-tune]: https://prometheus.io/docs/practices/remote_write/
[grafana-helm-hub]: https://grafana.github.io/helm-charts
