# Upgrading tobs

Firstly upgrade the helm repo to pull the latest available tobs helm chart. We always recommend upgrading to the [latest](https://github.com/timescale/tobs/releases/latest) tobs stack available.

```shell
helm repo update
```

## Upgrading from 19.x to 20.x

Version 20 of tobs is specifying `clusterName` for timescaledb resources. This is required to make seamless connection
string propagation work out of the box with version 0.27+ of timescaledb helm chart. Since tobs is now specifying
`clusterName` in values file, we took this opportunity to also change the default `clusterName` to `{{ .Release.Name }}-tsdb`.
This allows easier operations on the installed cluster as the objects are clearly associated with particular component.

Sadly this is a breaking change for users who are using tobs with version 0.26 or lower of timescaledb helm chart.
If you are using tobs with version 0.26 or lower of timescaledb helm chart and you don't want to manually migrate
your timescaledb resources, you can specify the following option in your `values.yaml` file.

```yaml
timescaledb-single:
  clusterName: "{{ .Release.Name }}"
```

## Upgrading from 18.x to 19.x

There is a breaking change in [kube-prometheus-stack](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack#from-41x-to-42x).
This will add a new configuration option `global.imageRegistry` along with adding
a `registry` configuration to all `image` objects.

```yaml
    image:
      registry: quay.io
      repository: prometheus/alertmanager
      tag: v0.24.0
```

## Upgrading from 17.x to 18.x

To get the best performance out of promscale we recommend to query it directly. Since tobs is already shipping grafana datasource configured this way, there is no need to configure `remote_read` option in prometheus. This is a breaking change for people using `remote_read` option. If you need to use `remote_read` option, you can still add it back by putting the following code snippet into your `values.yaml` file.

```yaml
kube-prometheus-stack:
  prometheus:
    prometheusSpec:
      remoteRead:
        - url: "http://{{ .Release.Name }}-promscale.{{ .Release.Namespace }}.svc:9201/read"
          readRecent: false
```

## Upgrading from 16.x to 17.x

With `17.0.0` we decided to diverge from gathering metrics data only from
namespace in which tobs is deployed and extend it to all namespaces. To
accomplish this we changed default kube-prometheus-stack selectors to gather
all prometheus-operator resources that are not labeled with `tobs/excluded`
(label value doesn't matter). If you have any other prometheus-operator
resources in your cluster that you don't want to be scraped by tobs, you need
to label them with `tobs/excluded` label.

Additionally, to prevent data duplication, we are disabling by default
ability to scrape endpoints using prometheus label annotations. If you wish
to continue using this option, you need to explicitly set the following
option:

```yaml
kube-prometheus-stack:
  prometheus:
    prometheusSpec:
      additionalScrapeConfigsSecret:
        enabled: true
```

In `17.0.0` we are also updating timescaledb-single chart to version `0.20.0`, which by default uses `ClusterIP` instead of `LoadBalancer` service. This change removes opttion removes field of `timescaledb-single.service.loadBalancerIP`.

## Upgrading from 15.x to 16.x

With `16.0.0` we removed `grafana-db-sec.yaml` generated Secret as it's no
longer needed to use with Grafana. If you wish to retain it, please make a
backup.

If you wish to keep using the `GF_DATABASE_*` env variables you will need to
create a new Secret and reference it in your Grafana configuration.  Here is
a simple example.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: grafana-db-secret
  namespace: default
type: Opaque
data:
  GF_DATABASE_HOST: host.svc.local:5432
  GF_DATABASE_NAME: postgres
  GF_DATABASE_USER: user
  GF_DATABASE_PASSWORD: pass
  GF_DATABASE_SSL_MODE:
  GF_DATABASE_TYPE: postgres
```

```yaml
kube-prometheus-stack:
  grafana:
    envFromSecret: "grafana-db-secret"
```

## Upgrading from 14.x to 15.x

Be aware that the upgrade of prometheus-node-exporter to `4.x.x` inside
kube-prometheus-stack changes to use the [Kubernetes Recommended Labels](https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/).
Therefore you may have to delete the DaemonSet before you upgrade.

Please see the notes from [kube-prometheus-stack](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack#from-39x-to-40x)
for more information

## Upgrading from 13.x to 14.x

### promscale image configuration change

Starting with tobs `14.0.0` the configuration of the Promscale image has changed.
If you are not overriding the Promscale version you can ignore this.  If you
are explicitly overriding the version you will need to follow the new image/tag
format

```shell
promscale:
  enabled: true
  image:
    repository: timescale/promscale
    tag: 0.13.0
    pullPolicy: IfNotPresent
```

## Upgrading from 12.0.x to 13.0.0

### kube-prometheus-stack name override removed

Starting with tobs `13.0.0` Helm chart the
`kube-prometheus-stack.fullNameOverride` option is removed in default
`values.yaml`. If you are upgrading it is suggested that you add it back.
Not adding it back will result in a Helm upgrade failure.  It will also delete
and redeploy the entire kube-prometheus-stack Helm chart.  This may delete any
non-ephemeral data that is stored in Prometheus, Alertmanager or Grafana.

The error upon upgrade is somewhat trivial and is due to the removal and
reinstallation of prometheus-operator.  Running it a second time, will result
in a successful upgrade.

```shell
helm upgrade --wait --timeout 15m <helm-release-name> timescale/tobs --version
13.0.0"

Error: UPGRADE FAILED: failed to create resource: Internal error occurred:
failed calling webhook "prometheusrulemutate.monitoring.coreos.com": failed to
call webhook: Post "https://tobs-kube-prometheus-operator.default.svc:443/admission-prometheusrules/validate?timeout=10s":
x509: certificate is valid for <helm-release-name>-kube-prom-operator,
<helm-release-name>-kube-prom-operator.default.svc, not tobs-kube-prometheus-operator.default.svc
```

If you wish to keep all current settings when running the upgrade please be
sure to add back the `kube-prometheus-stack.fullNameOverride` option in your
`values.yaml` or add it to your upgrade command

```shell
helm upgrade --wait --timeout 15m <helm-release-name> timescale/tobs --version
13.0.0 --set kube-prometheus-stack.fullNameOverride="tobs-kube-prometheus"
```

## Upgrading from 0.11.x to 12.x

### Promscale labels immutable

Due to a change in deployment labels in Promscale you will need to delete the Promscale deployment and reinstall when updating tobs.

```shell
Helm upgrade failed: cannot patch "tobs-promscale" with kind Deployment: Deployment.apps "tobs-promscale" is invalid: spec.selector: Invalid value: v1.LabelSelector{MatchLabels:map[st │
│ ing[]string{"app":"tobs-promscale", "release":"tobs"}, MatchExpressions:[]v1.LabelSelectorRequirement(nil)}: field is immutable
```

### tobs CLI deprecated

Starting with 12.0.0 the tobs CLI is now deprecated.  Going forward you must
install and upgrade using the Helm chart.

### SQL Datasource credential handling improvements

Starting with tobs `12.0.0` we are deprecating old way of setting up SQL datasource in grafana with kubernetes `Job` object and we are moving this to database initialization script. This in turn has a few consequences:

1) there is no longer a need to set timescaledb admin credentials in helm (options `kube-prometheus-stack.grafana.timescale.adminPassSecret` and `kube-prometheus-stack.grafana.timescale.adminUser`)
2) For new installations password will be created automatically, so there is no need to store it in helm values
3) If you are using external DB, you now need to create a user that will be used by grafana to access data from promscale and set proper values in:

```yaml
kube-prometheus-stack:
  grafana:
    timescale:
      user: "<<USERNAME>>"
      pass: "<<PASSWORD>>"
```

Adding user to the database can be done by executing a following SQL script:

```sql
\set ON_ERROR_STOP on
DO $$
  BEGIN
    CREATE ROLE prom_reader;
  EXCEPTION WHEN duplicate_object THEN
    RAISE NOTICE 'role prom_reader already exists, skipping create';
  END
$$;
DO $$
  BEGIN
    CREATE ROLE <<USERNAME>> WITH LOGIN PASSWORD '<<PASSWORD>>';
  EXCEPTION WHEN duplicate_object THEN
    RAISE NOTICE 'role <<USER>> already exists, skipping create';
  END
$$;
GRANT prom_reader TO <<USERNAME>>;
```

### Open-telemetry configuration change in values.yaml

Starting with tobs `12.0.0` the configuration of Open-telemetry has changed
from `opentelemetryOperator` to `opentelemetry-operator`. If you are using the
default values in `values.yaml` nothing should be needed. If you are
customizing values please make sure you have updated name.

## Upgrading to 0.11.0

Starting with tobs `0.11.0` we are tackling mostly reliability improvements. One of such improvements is switching grafana database back to dedicated sqlite3 instead of sharing TimescaleDB between grafana and promscale. Sadly this change requires manual intervention from end-users. If you wish to temporarily still use TimescaleDB as a grafana backend, you need to change following value:

```yaml
kube-prometheus-stack:
  grafana:
    timescale:
      database:
        enabled: true
```

Bear in mind that next tobs release will not support TimescaleDB as a grafana backend and you will need to migrate either to sqlite3 or to separate grafana instance.

## Upgrading to 0.10.0

With tobs `0.10.0` release we are starting a process of redesigning tobs. Most notable changes that may require user interaction are listed below.

### Open-telemetry by default and cert-manager requirement

This release enables opentelemetry support by default and as such it also requires cert-manager to be preinstalled. Please follow <https://cert-manager.io/docs/installation/> to get more information on how to install cert-manager. If you cannot use cert-manager, you still can use tobs but with opentelemetry support disabled. We are working to remove this limitation and allow installing opentelemetry-operator without cert-manager [(issue#198](https://github.com/timescale/tobs/issues/198)).

### TimescaleDB secrets management

Starting with tobs `0.10.0` we moved timescaledb secrets (certificates and credentials) management into helm. As such `tobs` cli no longer offers abilities to set those secrets. Side effect of this change is that you are no longer required to provide any secret in helm values or on tobs cli. TimescaleDB helm chart with generate new credentials on first run (and only on first run!) and kubernetes Job will copy it to promscale.

### Removal of jaeger-query

Jaeger ui and query endpoints are removed in this tobs release. This is done because grafana already offers similar UI
while promscale `0.11.0` has an integrated jaeger query endpoint. As such jaeger qeury is no longer needed and helm values located in``openTelemetry.jaeger` have to be removed to continue with installation.

## Upgrading to 0.8.0

With tobs `0.8.0` release there are multiple steps which needs to be performed manually to upgrade the tobs helm chart.

In tobs `0.8.0` we upgraded the CRDs of Prometheus-Operator that are part of Kube-Prometheus helm chart `30.0.0`. You need to manually upgrade the CRDs by following the instructions below.

```shell
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

```yaml
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

```shell
helm upgrade <release_name> timescale/tobs
```

## Upgrading to 0.7.0

Upgrade tobs:

```shell
helm upgrade <release_name> timescale/tobs
```

## Upgrading to 0.6.1

In tobs `0.6.1` we upgraded the CRDs of Prometheus-Operator that are part of Kube-Prometheus helm chart. You need to manually upgrade the CRDs by following the instructions below.

```shell
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

```shell
helm upgrade <release_name> timescale/tobs
```

## Upgrading to 0.6.x

Upgrade tobs:

```shell
helm upgrade <release_name> timescale/tobs
```

## Upgrading to 0.5.x

Upgrade tobs:

```shell
helm upgrade <release_name> timescale/tobs
```

## Upgrading to 0.4.x

In tobs `0.4.x` we swapped our existing Prometheus and Grafana helm charts with Kube-Prometheus helm charts. Kube-Prometheus depends on Prometheus-Operator which uses the CRDs (Custom Resource Definitions) to upgrade tobs. You need to manually install the CRDs:

```shell
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

```shell
kubectl delete daemonset <RELEASE_NAME>-prometheus-node-exporter -n <NAMESPACE>
```

Delete `tobs-node-exporter` service:

```shell
kubectl delete svc <RELEASE_NAME>-prometheus-node-exporter -n <NAMESPACE>
```

To migrate data from an old Prometheus instance to a new one follow the steps below:

Scale down the existing Prometheus replicas to 0 so that all the in-memory data is stored in Prometheus persistent volume.

```shell
kubectl scale --replicas=0 deploy/tobs-prometheus-server
```

**Note**: Wait for the Prometheus pod to gracefully shut down.

Find the Persistent Volume (PV) name that is claimed by the Persistent Volume Claim (PVC):

```shell
kubectl get pvc/<RELEASE_NAME>-prometheus-server
```

Patch the PVC reference to null so that new PVC created for the Kube-Prometheus stack will mount to the PV owned by the previous Prometheus pod.

```shell
kubectl edit pv/<PERSISTENT_VOLUME>
```

Now update the PVC reference field to `null` i.e. `spec.claimRef: null` so that new PVC will mount to this PV.

Create a new PVC and mount its volumeName to the PV released in the previous step:

```yaml
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

```shell
kubectl create -f pvc-file-name.yaml
```

Change the permissions of the Prometheus data directory as the new Kube-Prometheus instance comes with security context by default.

```yaml
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

```shell
kubectl create -f job-file-name.yaml
```

Now upgrade tobs:

```shell
helm upgrade <release_name> timescale/tobs
```

## Upgrading to 0.3.x

In tobs `0.3.x` TimescaleDB doesn't create the required secrets by default. During the upgrade you need to copy the existing timescaledb passwords to new secrets. This can be done by running this [script](https://github.com/timescale/timescaledb-kubernetes/blob/master/charts/timescaledb-single/upgrade-guide.md#migrate-the-secrets).

Delete the `grafana-db` job as the upgrade re-creates the same job for the upgraded tobs deployment

```shell
kubectl delete job/<RELEASE_NAME>-grafana-db -n <NAMESPACE>
```

Now upgrade tobs:

```shell
helm upgrade <release_name> timescale/tobs
```
