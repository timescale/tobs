# Upgrading tobs using helm (without tobs CLI)

The following steps are necessary if using helm without the tobs CLI. The tobs CLI will handle these upgrade tasks automatically for you.

Firstly upgrade the helm repo to pull the latest available tobs helm chart.  We always recommend upgrading to the [latest](https://github.com/timescale/tobs/releases/latest) tobs stack available. 
```
helm repo update
```

## Upgrading to 0.7.0:

Upgrade tobs:
```
helm upgrade <release_name> timescale/tobs
```

## Upgrading to 0.6.1:

In tobs `0.6.1` we upgraded the CRDs of Prometheus-Operator that are part of Kube-Prometheus helm chart. You need to manually upgrade the CRDs by following the instructions below. 

```
kubectl apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/v0.50.0/example/prometheus-operator-crd/monitoring.coreos.com_alertmanagerconfigs.yaml
kubectl apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/v0.50.0/example/prometheus-operator-crd/monitoring.coreos.com_alertmanagers.yaml
kubectl apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/v0.50.0/example/prometheus-operator-crd/monitoring.coreos.com_podmonitors.yaml
kubectl apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/v0.50.0/example/prometheus-operator-crd/monitoring.coreos.com_probes.yaml
kubectl apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/v0.50.0/example/prometheus-operator-crd/monitoring.coreos.com_prometheuses.yaml
kubectl apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/v0.50.0/example/prometheus-operator-crd/monitoring.coreos.com_servicemonitors.yaml
kubectl apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/v0.50.0/example/prometheus-operator-crd/monitoring.coreos.com_thanosrulers.yaml
kubectl apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/v0.50.0/example/prometheus-operator-crd/monitoring.coreos.com_prometheusrules.yaml
```

Upgrade tobs:
```
helm upgrade <release_name> timescale/tobs
```

## Upgrading to 0.6.x:

Upgrade tobs:
```
helm upgrade <release_name> timescale/tobs
```

## Upgrading to 0.5.x:

Upgrade tobs:
```
helm upgrade <release_name> timescale/tobs
```

## Upgrading to 0.4.x:

In tobs `0.4.x` we swapped our existing Prometheus and Grafana helm charts with Kube-Prometheus helm charts. Kube-Prometheus depends on Prometheus-Operator which uses the CRDs (Custom Resource Definitions) to upgrade tobs. You need to manually install the CRDs:

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

The tobs `0.4.x` installation uses the node-exporter daemonset & node-exporter service from the Kube-Prometheus stack. This requires manual deletion of these resources. The upgrade flow will recreate these resources. 

Delete `tobs-node-exporter` daemonset:

```
kubectl delete daemonset <RELEASE_NAME>-prometheus-node-exporter -n <NAMESPACE>
``` 

Delete `tobs-node-exporter` service:

```
kubectl delete svc <RELEASE_NAME>-prometheus-node-exporter -n <NAMESPACE>
```

To migrate data from an old Prometheus instance to a new one follow the steps below:

Scale down the existing Prometheus replicas to 0 so that all the in-memory data is stored in Prometheus persistent volume. 

```
kubectl scale --replicas=0 deploy/tobs-prometheus-server 
```
**Note**: Wait for the Prometheus pod to gracefully shut down.

Find the Persistent Volume (PV) name that is claimed by the Persistent Volume Claim (PVC):

```
kubectl get pvc/<RELEASE_NAME>-prometheus-server
```

Patch the PVC reference to null so that new PVC created for the Kube-Prometheus stack will mount to the PV owned by the previous Prometheus pod.

```
kubectl edit pv/<PERSISTENT_VOLUME>
```

Now update the PVC reference field to `null` i.e. `spec.claimRef: null` so that new PVC will mount to this PV. 

Create a new PVC and mount its volumeName to the PV released in the previous step:

```
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  labels:
    app: prometheus
    prometheus: tobs-kube-prometheus
    release: tobs
  name: prometheus-tobs-kube-prometheus-prometheus-db-prometheus-tobs-kube-prometheus-prometheus-0
  namespace: <NAMESPACE>
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 8Gi
  volumeMode: Filesystem
  volumeName: <PERSISTENT_VOLUME>
```

Create the PVC defined in the above code snippet:

```
kubectl create -f pvc-file-name.yaml
```

Change the permissions of the Prometheus data directory as the new Kube-Prometheus instance comes with security context by default.

```
apiVersion: batch/v1
kind: Job
metadata:
  creationTimestamp: "2021-05-16T16:37:21Z"
  labels:
    app: tobs-upgrade
    heritage: helm
    release: tobs
  name: tobs-prometheus-permission-change
  namespace: default
spec:
  template:
    metadata:
      labels:
        job-name: tobs-prometheus-permission-change
    spec:
      restartPolicy: OnFailure
      containers:
      - command:
        - chown
        - 1000:1000
        - -R
        - /data/
        image: alpine
        imagePullPolicy: IfNotPresent
        name: upgrade-tobs
        volumeMounts:
        - mountPath: /data
          name: prometheus
      volumes:
      - name: prometheus
        persistentVolumeClaim:
          claimName: prometheus-tobs-kube-prometheus-prometheus-db-prometheus-tobs-kube-prometheus-prometheus-0
```

Create the job from the code snippet defined above:

```
kubectl create -f job-file-name.yaml
```


Now upgrade tobs:
```
helm upgrade <release_name> timescale/tobs
```

## Upgrading to 0.3.x:

In tobs `0.3.x` TimescaleDB doesn't create the required secrets by default. During the upgrade you need to copy the existing timescaledb passwords to new secrets. This can be done by running this [script](https://github.com/timescale/timescaledb-kubernetes/blob/master/charts/timescaledb-single/upgrade-guide.md#migrate-the-secrets).

Delete the `grafana-db` job as the upgrade re-creates the same job for the upgraded tobs deployment

```
kubectl delete job/<RELEASE_NAME>-grafana-db -n <NAMESPACE>
``` 

Now upgrade tobs:
```
helm upgrade <release_name> timescale/tobs
```