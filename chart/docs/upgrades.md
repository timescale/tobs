# Upgrading tobs

Firstly upgrade the tobs CLI to the latest by running the [install-cli.sh](../../install-cli.sh). Now the tobs CLI is upgraded to the latest. We always recommend to upgrading latest tobs stack available. 

## Upgrading to 0.4.x:

In tobs `0.4.x` we swapped our existing Prometheus and Grafana helm charts with Kube-Prometheus helm charts. The Kube-Prometheus depends on Prometheus-Operator which uses the CRDs (Custom Resource Definitations) to upgrade the tobs. You need to manually install the CRDs by running the steps:

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
tobs upgrade
```