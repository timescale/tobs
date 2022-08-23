# Connect tobs to an external TimescaleDB

## Issue

In the past it wasn't straight forward to connect a tobs provisioned Promscale
instance to a non-tobs provisioned TimescaleDB instance.  This documentation
should provide best practices in making this work pretty seamless.

## Connecting using a Postgres URI

It's now possible to set a connection URI in the Promscale configuration in
tobs.  To do so you just need to edit your Helm `values.yaml` or set the value
in the CLI.  Here is a more complete example:
[example-values.yaml](../chart/ci/externaldb-values.yaml)

```yaml
timescaledb-single:
  enabled: false

promscale:
  connectionSecretName: ""
  connection:
    uri: "postgres://user@pass:host.svc.local:5432/database"

kube-prometheus-stack:
  grafana:
    envValueFrom: null
```

If you have Grafana enabled this will not setup the datasource connection for
you automatically.  If you wish to have that connection you will need to do so
manually for now. The following sql will need to be ran on the TimescaleDB
instance after Promscale has setup the connection and is running. Just replace
the `<grafana role>` with the name of the role you wish to give it.

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
    CREATE ROLE <grafana role> WITH LOGIN PASSWORD 'pass';
  EXCEPTION WHEN duplicate_object THEN
    RAISE NOTICE 'role <grafana role> already exists, skipping create';
  END
$$;
GRANT prom_reader TO <grafana role>;
```

For any other information for creating roles and users please look at the
documentation
[here](https://docs.timescale.com/promscale/latest/roles-and-permissions/#example-permissions)
