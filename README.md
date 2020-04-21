# Timescale Observability

A Helm chart for deploying Prometheus configured to use TimescaleDB as compressed long-term store for time-series metrics through the Timescale-Prometheus Connector.

For a detailed description of the architecture, please see [our design doc][design-doc].

### Table of contents
* **[Quick start](#quick-start)**
  * **[Cleanup](#cleanup)**
    * [TimescaleDB PVCs and Backup](#timescaledb-pvcs-and-backup)
    * [TimescaleDB config service](#timescaledb-config-service)
* **[Configuring Helm Chart](#configuring-helm-chart)**
  * **[TimescaleDB related values](#timescaledb-related-values)**
    * [Additional configuration for TimescaleDB](#additional-configuration-for-timescaledb)
  * **[Timescale-Prometheus related values](#timescale-prometheus-related-values)**
    * [Additional configuration for Timescale-Prometheus Connector](#additional-configuration-for-timescale-prometheus-connector)
  * **[Prometheus related values](#prometheus-related-values)**
    * [Additional configuration for Prometheus](#additional-configuration-for-prometheus)
  * **[Grafana related values](#grafana-related-values)**
    * [Additional configuration for Grafana](#additional-configuration-for-grafana)
* **[Contributing](#contributing)**

# Quick start

The following command will install Prometheus, TimescaleDB, Timescale-Prometheus Connector, and Grafana
into your Kubernetes cluster:
```
helm repo add timescale https://charts.timescale.com/
helm repo update
helm install <release_name> timescale/timescale-observability
```

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

Sometimes one of the services created with the deployment is not deleted. The `<release_name>-config` service
may need to be manually deleted with 
```
kubectl delete svc <release_name>-config
```

# Configuring Helm Chart

To get a fully-documented configuration file for `timescale-observability`, please run:

```
helm show values timescale/timescale-observability > my_values.yml
```

You can then edit `my_values.yml` and deploy the release with the following command:

```
helm upgrade --install <release_name> --values my_values.yml timescale/timescale-observability
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

By default, the `timescale-observability` Helm chart sets up a single-instance of TimescaleDB; if you are
interested in a replicated setup for high-availability with automated backups, please see
[this github repo][timescaledb-helm-repo] for additional instructions.

You can set up the credentials, nodeSelector, volume sizes (default volumes created are 1GB for WAL and 2GB for storage).

## Timescale-Prometheus related values
| Parameter                                           | Description                                           | Default     |
|-----------------------------------------------------|-------------------------------------------------------|-------------|
| `timescale-prometheus.enabled`                      | If false Timescale-Prometheus Connector will not be created| `true` |
| `timescale-prometheus.image`                        | Docker image to use for the Connector                 | `timescale/timescale-prometheus:0.1.0-alpha.2` |
| `timescale-prometheus.connection.dbName`            | Database to store the metrics in                      | `postgres`  |
| `timescale-prometheus.service.loadBalancer.enabled` | Create a LB for the connector instead of a Cluster IP | `false`     |
| `timescale-prometheus.resources.requests.memory`    | Amount of memory for the Connector pod                | `2Gi`       |
| `timescale-prometheus.resources.requests.cpu`       | Number of vCPUs for the Connector pod                 | `1`         |
| `timescale-prometheus.remote.queue.max-shards`      | Max number of shards the prometheus server should create for the Timescale-Prometheus endpoint | `30` |

### Additional configuration for Timescale-Prometheus Connector

The connector is configured to connect to the TimescaleDB instance deployed with this chart.
But it can be configured to connect to any TimescaleDB host, and expose whichever port you like.
For more details about how to configure the Timescale-Prometheus connector please see the
[Helm chart directory][timescale-prometheus-helm] of the [Timescale-Prometheus][timescale-prometheus-repo] repo.

## Prometheus related values

| Parameter                                           | Description                                           | Default     |
|-----------------------------------------------------|-------------------------------------------------------|-------------|
| `prometheus.enabled`                                | If false, none of the Prometheus resources will be created | `true` |
| `prometheus.alertmanager.enabled`                   | If true will create the Prometheus Alert Manager       | `false`    |
| `prometheus.pushgateway.enabled`                    | If true will create the Prometheus Push Gateway        | `false`    |
| `prometheus.server.configMapOverrideName`           | The name of the ConfigMap that provides the Prometheus config. Resolves to `{{ .Release.Name }}-{{ .Values.prometheus.server.configMapOverrideName }}` | `prometheus-config` |

### Additional configuration for Prometheus

The stable/prometheus chart is used as a dependency for deploying Prometheus. We specify
a ConfigMap override where the Timescale-Prometheus Connector is already configured as a `remote_write`
and `remote_read` endpoint. We create a ConfigMap that is still compatible and respects all the configuration
properties for the prometheus chart, so no functionality is lost.

For all the properties that can be configured and more details on how to set up the Prometheus
deployment see the [Helm hub entry][prometheus-helm-hub].

For more information about the `remote_write` configuration that can be set with
`timescale-prometheus.remote.queue` visit the Prometheus [Remote Write Tuning][prometheus-remote-tune] guide.

## Grafana related values

| Parameter                                           | Description                                           | Default     |
|-----------------------------------------------------|-------------------------------------------------------|-------------|
| `grafana.enabled`                                   | If false, Grafana will not be created                 | `true`      |
| `grafana.sidecar.datasources.enabled`               | If false, no data sources will be provisioned         | `true`      |
| `grafana.sidecar.datasources.prometheus.enabled`    | If false, a Prometheus data source will not be provisioned | `true` |
| `grafana.sidecar.datasources.prometheus.urlTemplate` | Template parsed to the url of the Prometheus API. Defaults to Prometheus deployed with this chart  | `http://{{ .Release.Name }}-prometheus-service.{{ .Release.Namespace }}.svc.cluster.local` |
| `grafana.sidecar.datasources.timescaledb.enabled` | If false a TimescaleDB data source will not be provisioned | `true` |
| `grafana.sidecar.datasources.timescaledb.user` | User to connect to TimescaleDB | `postgres`
| `grafana.sidecar.datasources.timescaledb.urlTemplate` | Template that will be parsed to the host of TimescaleDB, defaults to DB deployed with this chart | `{{ .Release.Name }}.{{ .Release.Namespace }}.svc.cluster.local` |
| `grafana.sidecar.dashboards.enabled`                | If false, no dashboards will be provisioned by default | `true`     |
| `grafana.sidecar.dashboards.files`                  | Files with dashboard definitions (in JSON) to be provisioned | `['dashboards/k8s-cluster.json','dashboards/k8s-hardware.json']` |

### Additional configuration for Grafana

The stable/grafana chart is used as a dependency for deploying Grafana. We specify a Secret that 
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

For all the properties that can be configured and more details on how to set up the Grafana deployment see the [Helm hub entry][grafana-helm-hub]

# Contributing

We welcome contributions to the Timescale-Observability Helm charts, which is
licensed and released under the open-source Apache License, Version 2.  The
same [Contributor's
Agreement](//github.com/timescale/timescaledb/blob/master/CONTRIBUTING.md)
applies as in TimescaleDB; please sign the [Contributor License
Agreement](https://cla-assistant.io/timescale/timescale-observability) (CLA) if
you're a new contributor.


[design-doc]: https://tsdb.co/prom-design-doc
[timescaledb-helm-cleanup]: https://github.com/timescale/timescaledb-kubernetes/blob/master/charts/timescaledb-single/admin-guide.md#optional-delete-the-s3-backups
[timescaledb-helm-repo]: https://github.com/timescale/timescaledb-kubernetes/tree/master/charts/timescaledb-single
[timescale-prometheus-repo]: https://github.com/timescale/timescale-prometheus
[timescale-prometheus-helm]: https://github.com/timescale/timescale-prometheus/tree/master/helm-chart
[prometheus-helm-hub]: https://hub.helm.sh/charts/stable/prometheus
[prometheus-remote-tune]: https://prometheus.io/docs/practices/remote_write/
[grafana-helm-hub]: https://hub.helm.sh/charts/stable/grafana
