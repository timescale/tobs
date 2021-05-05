# Prometheus High-Availability

**Note**: This is unnecessary if using the tobs CLI, To enable Prometheus high-availability with tobs CLI use `--enable-prometheus-ha`. 

The following steps will explain how to enable Prometheus high-availability with Promscale when using tobs helm chart (without tobs CLI). 

Update the tobs `values.yaml` with below HA configuration.

Increase the TimescaleDB connection pool i.e.

```
timescaledb-single:
  patroni:
    bootstrap:
      dcs:
        postgresql:
          parameters:
            max_connections: 400
```

Update the Promscale configuration to enable HA mode and increase the replicas to 3:

```
promscale:
  replicaCount: 3
  args:
    - --high-availability
```

Update Prometheus configuration to send prometheus pod name with `__replica__` and prometheus cluster name as `cluster` labelSets in the form of external labels and run Prometheus as 3 replicas for HA. 

```
kube-prometheus-stack:
  prometheus:
    prometheusSpec:
      replicaExternalLabelName: "__replica__"
      prometheusExternalLabelName: "cluster"
      replicas: 3
```