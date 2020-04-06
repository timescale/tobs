# Timescale Observability

A Helm chart for deploying Prometheus configured to use TimescaleDB as compressed long-term store for time-series metrics through the Timescale-Prometheus Connector.

For a detailed description of the architecture, please see [our design doc][design-doc].

# Quick start

The following command will install Prometheus, TimescaleDB, and Timescale-Prometheus Connector
into your Kubernetes cluster:
```
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

By default, the `timescale-observability` Helm chart sets up a single-instance of TimescaleDB; if you are
interested in a replicated setup for high-availability with automated backups, please see
[this github repo][timescaledb-helm-repo] for additional instructions.

For more details about how to configure the Timescale-Prometheus connector please see the [Helm chart directory][timescale-prometheus-helm] of the [Timescale-Prometheus][timescale-prometheus-repo] repo.

[design-doc]: https://www.timescale.com/404/
[timescaledb-helm-repo]: https://github.com/timescale/timescaledb-kubernetes/tree/master/charts/timescaledb-single
[timescale-prometheus-repo]: https://github.com/timescale/timescale-prometheus
[timescale-prometheus-helm]: https://github.com/timescale/timescale-prometheus/tree/master/helm-chart
