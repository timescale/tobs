# Upgrading tobs using helm (without tobs CLI)

The following steps are necessary if using helm without the tobs CLI. The tobs CLI will handle these upgrade tasks automatically for you.

Firstly upgrade the helm repo to pull the latest available tobs helm chart.  We always recommend upgrading to the [latest](https://github.com/timescale/tobs/releases/latest) tobs stack available. 
```
helm repo update
```
## Upgrading to 0.8.0

With tobs `0.8.0` release there are multiple steps which needs to be performed manually to upgrade the tobs helm chart.

In tobs `0.8.0` we upgraded the CRDs of Prometheus-Operator that are part of Kube-Prometheus helm chart `30.0.0`. You need to manually upgrade the CRDs by following the instructions below.

```
kubectl apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/v0.53.1/example/prometheus-operator-crd/monitoring.coreos.com_alertmanagerconfigs.yaml
kubectl apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/v0.53.1/example/prometheus-operator-crd/monitoring.coreos.com_alertmanagers.yaml
kubectl apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/v0.53.1/example/prometheus-operator-crd/monitoring.coreos.com_podmonitors.yaml
kubectl apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/v0.53.1/example/prometheus-operator-crd/monitoring.coreos.com_probes.yaml
kubectl apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/v0.53.1/example/prometheus-operator-crd/monitoring.coreos.com_prometheuses.yaml
kubectl apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/v0.53.1/example/prometheus-operator-crd/monitoring.coreos.com_servicemonitors.yaml
kubectl apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/v0.53.1/example/prometheus-operator-crd/monitoring.coreos.com_thanosrulers.yaml
kubectl apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/v0.53.1/example/prometheus-operator-crd/monitoring.coreos.com_prometheusrules.yaml
```

### Re-structure the Promscale values section

1. If you are using the default `values.yaml` from tobs helm chart, copy the `0.8.0` values.yaml so the default values are 
structured as expected, assign the database password by reading it from `<release-name>-credentials` secret with key `PATRONI_SUPERUSER_PASSWORD` 
decode the base64 encoded password and assign it to `promscale.connection.password` in the new `values.yaml` you can skip step 2.
2. If you are using the custom `values.yaml` for tobs installation make the below suggested changes, 
```
    promscale:
        # tracing field name has been changed to openTelemetry
        openTelemetry:
            enabled: <value>
        # as this is your custom values.yaml
        # do not forget to change Promscale image to 0.8.0 tag
        image: timescale/promscale:0.8.0    
        connection:
            # assign the db-password here, the password should be in <release-name>-credentials secret from previous installation
            # with key PATRONI_SUPERUSER_PASSWORD
            password: <value>
            # if you are using db-uri based auth assign the db-uri to below field
            uri: <>
        # change service section only if you enabling LoadBalancer type service for Promscale
        service:
            type: LoadBalancer  
```
4. If you want to enable tracing do not forget to enable `promscale.openTelemetry.enabled` to true and `openTelemetryOperator.enabled` to true.
5. If you are using Promscale HA with Prometheus HA change the Promscale HA arg from `--high-availability` to `--metrics.high-availability` in `promscale.extraArgs`. 
6. Drop `timescaledbExternal` section of `values.yaml` as the db-uri will be observed from `promscale.connection.db_uri` if configured any. 

### Re-structure openTelemetryOperator values section (only if you have enabled tracing)

1. Drop `jaegerPromscaleQuery` section in `openTelemetryOperator` as we have moved from Jaeger Promscale gRPC based plugin to integrating directly with upstream Jaeger query.
2. Add the existing default openTelemetry collector config in `values.yaml` at `openTelemetryOperator.collector.config` as mentioned [here](https://github.com/timescale/tobs/blob/0.8.0/chart/values.yaml#L432).

**Note**: If tracing is enabled upgrade the `cert-manager` to `v1.6.1` as the latest openTelemetryOperator expects the cert-manager of `v1.6.1` version.

### Delete Kube-State-Metrics as per Kube-Prometheus stack upgrade docs

1. With the upgrade the kube-state-metrics will be re-deployed. The existing deployment cannot be upgraded so delete it using `kubectl delete deployment/<tobs-release-name>-kube-state-metrics -n <namespace>`. 
For more reference on kube-state-metrics deletion follow Kube-Prometheus docs [here](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack#from-21x-to-22x).

### Delete tobs-grafana-job 

1. With the upgrade the tobs-grafana-job will be re-deployed. The existing job cannot be upgraded so delete it using `kubectl delete job/<tobs-release-name>-grafana-db`

### Upgrade tobs

```
helm upgrade <release_name> timescale/tobs
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
