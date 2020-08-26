# tobs - The Observability Stack for Kubernetes

Tobs deploys an observability stack consisting of Prometheus, TimescaleDB, and Grafana to
monitor a Kubernetes cluster. It is packaged as a Helm chart that is controlled through a CLI tool.

For a detailed description of the architecture, please see [our design doc][design-doc].

# Quick start

__Dependencies__: [Helm](https://helm.sh/docs/intro/install/)

To download and install tobs, run the following in your terminal, then follow the on-screen instructions.

```bash
curl --proto '=https' --tlsv1.2 -sSLf  https://tsdb.co/install-tobs-sh |sh
```

Alternatively, you can download the CLI directly via [our releases page](/releases)

After setting up tobs run the following to install the tobs helm charts into your kubernetes cluster

```bash
tobs install
```

# Using the Helm charts without the CLI tool

Instructions on using the Helm charts without the CLI tool can be found [here](/chart/README.md).

# Contributing

We welcome contributions to tobs, which is
licensed and released under the open-source Apache License, Version 2.  The
same [Contributor's
Agreement](//github.com/timescale/timescaledb/blob/master/CONTRIBUTING.md)
applies as in TimescaleDB; please sign the [Contributor License
Agreement](https://cla-assistant.io/timescale/tobs) (CLA) if
you're a new contributor.


[design-doc]: https://tsdb.co/prom-design-doc
[timescaledb-helm-cleanup]: https://github.com/timescale/timescaledb-kubernetes/blob/master/charts/timescaledb-single/admin-guide.md#optional-delete-the-s3-backups
[timescaledb-helm-repo]: https://github.com/timescale/timescaledb-kubernetes/tree/master/charts/timescaledb-single
[timescale-prometheus-repo]: https://github.com/timescale/timescale-prometheus
[timescale-prometheus-helm]: https://github.com/timescale/timescale-prometheus/tree/master/helm-chart
[prometheus-helm-hub]: https://hub.helm.sh/charts/stable/prometheus
[prometheus-remote-tune]: https://prometheus.io/docs/practices/remote_write/
[grafana-helm-hub]: https://hub.helm.sh/charts/stable/grafana
