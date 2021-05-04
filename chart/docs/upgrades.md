# Upgrading tobs using helm (without tobs CLI)

Firstly upgrade the helm repo to pull the latest available tobs helm chart.  We always recommend to upgrading latest tobs stack available. 
```
helm repo update
```

## Upgrading to 0.4.x:

In tobs `0.4.x` we swapped our existing Prometheus and Grafana helm charts with Kube-Prometheus helm charts. The Kube-Prometheus depends on Prometheus-Operator which uses the CRDs (Custom Resource Definitations) to upgrade the tobs. You need to manually install the CRDs by running:

```
kubectl apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/v0.47.0/example/prometheus-operator-crd/monitoring.coreos.com_alertmanagerconfigs.yaml
kubectl apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/v0.47.0/example/prometheus-operator-crd/monitoring.coreos.com_alertmanagers.yaml
kubectl apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/v0.47.0/example/prometheus-operator-crd/monitoring.coreos.com_podmonitors.yaml
kubectl apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/v0.47.0/example/prometheus-operator-crd/monitoring.coreos.com_probes.yaml
kubectl apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/v0.47.0/example/prometheus-operator-crd/monitoring.coreos.com_prometheuses.yaml
kubectl apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/v0.47.0/example/prometheus-operator-crd/monitoring.coreos.com_servicemonitors.yaml
kubectl apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/v0.47.0/example/prometheus-operator-crd/monitoring.coreos.com_thanosrulers.yaml
kubectl apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/v0.47.0/example/prometheus-operator-crd/monitoring.coreos.com_prometheusrules.yaml
```

Now upgrade the tobs by running:
```
helm upgrade <release_name> timescale/tobs
```

## Upgrading to 0.3.x:

In tobs `0.3.x` TimescaleDB doesn't create the secrets by default. During the upgrade you need to copy the existing timescaledb passwords to new secrets. This can be done by running this [script](https://github.com/timescale/timescaledb-kubernetes/blob/master/charts/timescaledb-single/upgrade-guide.md#migrate-the-secrets).

Now upgrade the tobs by running:
```
helm upgrade <release_name> timescale/tobs
```