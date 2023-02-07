> **Warning**
>
> Tobs has been discontinued and is deprecated.
>
> The code in this repository is no longer maintained.
>
> [Learn more](https://github.com/timescale/promscale/issues/1836).

[![Security Audit](https://github.com/timescale/tobs/actions/workflows/sec-audit.yml/badge.svg)](https://github.com/timescale/tobs/actions/workflows/sec-audit.yml)
[![Test Helm Charts](https://github.com/timescale/tobs/actions/workflows/tests.yml/badge.svg)](https://github.com/timescale/tobs/actions/workflows/tests.yml)
[![Version](https://img.shields.io/github/v/release/timescale/tobs)](https://github.com/timescale/tobs/releases)
[![Commit activity](https://img.shields.io/github/commit-activity/m/timescale/tobs)](https://github.com/timescale/tobs/pulse/monthly)
[![License](https://img.shields.io/github/license/timescale/tobs)](https://github.com/timescale/tobs/blob/main/LICENSE)
[![Slack](https://img.shields.io/badge/chat-join%20slack-brightgreen.svg)](https://timescaledb.slack.com/)

# tobs - The Observability Stack for Kubernetes

Tobs is a tool that aims to make it as easy as possible to install a full observability
stack into a Kubernetes cluster. Currently this stack includes:

<img src="docs/assets/tobs-arch.png" alt="Tobs Architecture Diagram" width="800"/>

* [Kube-Prometheus](https://github.com/prometheus-operator/kube-prometheus#kube-prometheus) the Kubernetes monitoring stack
  * [Prometheus](https://github.com/prometheus/prometheus) to collect metrics
  * [AlertManager](https://github.com/prometheus/alertmanager#alertmanager-) to fire the alerts
  * [Grafana](https://github.com/grafana/grafana) to visualize what's going on
  * [Node-Exporter](https://github.com/prometheus/node_exporter) to export metrics from the nodes
  * [Kube-State-Metrics](https://github.com/kubernetes/kube-state-metrics) to get metrics from kubernetes api-server
  * [Prometheus-Operator](https://github.com/prometheus-operator/prometheus-operator#prometheus-operator) to manage the life-cycle of Prometheus and AlertManager custom resource definitions (CRDs)
* [Promscale](https://github.com/timescale/promscale) ([design doc](https://tsdb.co/prom-design-doc)) to store metrics for the long-term and allow analysis with both PromQL and SQL.
* [TimescaleDB](https://github.com/timescale/timescaledb) for long term storage of metrics and provides ability to query metrics data using SQL.
  * [Postgres-Exporter](https://github.com/prometheus-community/postgres_exporter) to get metrics from PostgreSQL server
* [Opentelemetry-Operator](https://github.com/open-telemetry/opentelemetry-operator#opentelemetry-operator-for-kubernetes) to manage the lifecycle of OpenTelemetryCollector Custom Resource Definition (CRDs)

We plan to expand this stack over time and welcome contributions.

Tobs provides a helm chart to make deployment and operations easier. It can be used directly or as a sub-chart for other projects.

# Quick start

## Prerequisites

Using tobs to install full observability stack with openTelemetry support currently requires installation of cert-manager.
To do install it please follow [cert-manager documentation](https://cert-manager.io/docs/installation/).

*Note*: cert-manager is not required when using tobs with opentelemetry support disabled.

## Installing the helm chart

The following command will install Kube-Prometheus, OpenTelemetry Operator, TimescaleDB, and Promscale
into your Kubernetes cluster:

```
helm repo add timescale https://charts.timescale.com/
helm repo update
helm install --wait <release_name> timescale/tobs
```

*Note*: `--wait` flag is necessary for successfull installation as tobs helm chart can create opentelemetry Custom Resources only after opentelemetry-operator is up and running. This flag can be omitted when using tobs without opentelemetry support.

For detailed configuration and usage instructions, take a look at the [helm chart's README](/chart/README.md).

# Configuring the stack

All configuration for all components happens through the helm values.yaml file.
You can view the self-documenting [default values.yaml](chart/values.yaml) in the repo.
We also have additional documentation about individual configuration settings in our
[Helm chart docs](chart/README.md#configuring-helm-chart).

# Compatibility matrix

## Tobs vs. Kubernetes

| Tobs Version | Kubernetes Version |
|--------------|--------------------|
| 12.0.x       | v1.23 to v1.24     |
| 0.11.x       | v1.23 to v1.24     |
| 0.10.x       | v1.21 to v1.23     |
| 0.9.x        | v1.21 to v1.23     |
| 0.8.x        | v1.21 to v1.23     |
| 0.7.x        | v1.19 to v1.21     |

# Contributing

We welcome contributions to tobs, which is
licensed and released under the open-source Apache License, Version 2.  The
same [Contributor's
Agreement](https://github.com/timescale/timescaledb/blob/master/CONTRIBUTING.md)
applies as in TimescaleDB; please sign the [Contributor License
Agreement](https://cla-assistant.io/timescale/tobs) (CLA) if
you're a new contributor.
