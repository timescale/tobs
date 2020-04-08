# Timescale Observability

A Helm chart for deploying Prometheus configured to use TimescaleDB as compressed long-term store for time-series metrics through the Timescale-Prometheus Connector.

For a detailed description of the architecture, please see [our design doc][design-doc].

# Quick start

The following command will install Prometheus, TimescaleDB, and Timescale-Prometheus Connector
into your Kubernetes cluster:
```
helm repo add timescale https://charts.timescale.com/
helm repo update
helm install <release_name> timescale/timescale-observability
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

The chart defines the following parameters in the `values.yaml` file:

| Parameter                                           | Description                                           | Default     |
|-----------------------------------------------------|-------------------------------------------------------|-------------|
| `timescaledb-single.enabled`                        | If false TimescaleDB will not be created              | `true`      |
| `timescaledb-single.loadBalancer.enabled`           | Create a LB for the DB instead of a ClusterIP         | `false`     |
| `timescaledb-single.replicaCount`                   | Number of pods for DB, set to 3 for HA                | `1`         |
| `timescale-prometheus.enabled`                      | If false Timescale-Prometheus Connector will not be created | `true` |
| `timescale-prometheus.service.loadBalancer.enabled` | Create a LB for the connector instead of a Cluster IP | `false`     |
| `timescale-prometheus.resources.requests.memory`    | Amount of memory for the Connector pod                | `2Gi`       |
| `timescale-prometheus.resources.requests.cpu`       | Number of vCPUs for the Connector pod                 | `1`         |
| `timescale-prometheus.remote.queue.max-shards`      | Max number of shards the prometheus server should create for the Timescale-Prometheus endpoint | `30` |
| `prometheus.enabled`                                | If false, none of the Prometheus resources will be created | `true` |
| `prometheus.alertmanager.enabled`                   | If true will create the Prometheus Alert Manager       | `false`    |
| `prometheus.pushgateway.enabled`                    | If true will create the Prometheus Push Gateway        | `false`    |
| `prometheus.server.configMapOverrideName`           | The name of the ConfigMap that provides the Prometheus config. Resolves to `{{ .Release.Name }}-{{ .Values.prometheus.server.configMapOverrideName }}` | `prometheus-config` |

The properties described in the table above are only those that this chart overrides for each of the sub-charts it depends on.
You can additionally change any of the configurable properties of each sub-chart.

## Additional configuration for TimescaleDB

By default, the `timescale-observability` Helm chart sets up a single-instance of TimescaleDB; if you are
interested in a replicated setup for high-availability with automated backups, please see
[this github repo][timescaledb-helm-repo] for additional instructions.

You can set up the credentials, nodeSelector, volume sizes (default volumes created are 1GB for WAL and 2GB for storage).

## Additional configuration for Timescale-Prometheus Connector

The connector is configured to connect to the TimescaleDB instance deployed with this chart.
But it can be configured to connect to any TimescaleDB host, and expose whichever port you like.
For more details about how to configure the Timescale-Prometheus connector please see the
[Helm chart directory][timescale-prometheus-helm] of the [Timescale-Prometheus][timescale-prometheus-repo] repo.

## Additional configuration for Prometheus

The stable/prometheus chart is used as a dependency for deploying Prometheus. We specify
a ConfigMap override where the Timescale-Prometheus Connector is already configured as a `remote_write`
and `remote_read` endpoint. We create a ConfigMap that is still compatible and respects all the configuration
properties for the prometheus chart, so no functionality is lost.

For all the properties that can be configured and more details on how to set up the Prometheus
deployment see the [Helm hub entry][prometheus-helm-hub].

For more information about the `remote_write` configuration that can be set with
`timescale-prometheus.remote.queue` visit the Prometheus [Remote Write Tuning][prometheus-remote-tune] guide.


## Contributing

We welcome contributions to the Timescale-Observability Helm charts, which is
licensed and released under the open-source Apache License, Version 2.  The
same [Contributor's
Agreement](//github.com/timescale/timescaledb/blob/master/CONTRIBUTING.md)
applies as in TimescaleDB; please sign the [Contributor License
Agreement](https://cla-assistant.io/timescale/timescale-observability) (CLA) if
you're a new contributor.


[design-doc]: https://www.timescale.com/404/
[timescaledb-helm-repo]: https://github.com/timescale/timescaledb-kubernetes/tree/master/charts/timescaledb-single
[timescale-prometheus-repo]: https://github.com/timescale/timescale-prometheus
[timescale-prometheus-helm]: https://github.com/timescale/timescale-prometheus/tree/master/helm-chart
[prometheus-helm-hub]: https://hub.helm.sh/charts/stable/prometheus
[prometheus-remote-tune]: https://prometheus.io/docs/practices/remote_write/
