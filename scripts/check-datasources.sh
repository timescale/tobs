#!/bin/bash

# TODO(paulfantom): consider using jsonnet for modifications from this script

set -euo pipefail

function query() {
    local uid="$1"
    local query="$2"
    local format="$3"

    local body=$(cat <<-EOM
{
    "queries":[
        {
            "datasource":{
                "uid":"$uid"
            },
            "refId":"A",
            "format":"$format",
            "expr":"$query"
        }
    ],
    "from":"now-5m",
    "to":"now"
}
EOM
)
    curl -H "Content-Type: application/json" -X POST -d "$body" "http://${GRAFANA_USER}:${GRAFANA_PASS}@localhost:3000/api/ds/query" 2>/dev/null | jq '.results.A'
}


# Get helm release name
RELEASE="$(helm list -o json | jq -r '.[0].name')"
NAMESPACE="$(helm list -o json | jq -r '.[0].namespace')"

# Get grafana credentials
GRAFANA_USER="admin"
GRAFANA_PASS="$(kubectl get secret -n "${NAMESPACE}" "${RELEASE}-grafana" -o json | jq -r '.data["admin-password"]' | base64 -d)"

# Cleanup port-forward on exit
trap 'kill $(jobs -p)' EXIT

# Port-forward to grafana SVC
kubectl -n "${NAMESPACE}" port-forward svc/test-grafana 3000:80 &
sleep 5

SQL_QUERY="SELECT * FROM pg_extension WHERE extname = 'timescaledb_toolkit';"
RESULT_SQL=$(query "c4729dfb8ceeaa0372ef27403a3932695eee995d" "$SQL_QUERY" "table")
if [ "$(jq 'has("error")' <<< ${RESULT_SQL})" == "true" ]; then
    echo "GRAFANA SQL DATASOURCE CANNOT QUERY DATA DUE TO:"
    jq '.error' <<< ${RESULT_SQL}
    exit 1
fi

RESULT_PRM=$(query "dc08d25c8f267b054f12002f334e6d3d32a853e4" "ALERTS" "time_series")
if [ "$(jq 'has("error")' <<< ${RESULT_PRM})" == "true" ]; then
    echo "GRAFANA PROMQL DATASOURCE CANNOT QUERY DATA DUE TO:"
    jq '.error' <<< ${RESULT_PRM}
    exit 1
fi

echo "All queries passed"