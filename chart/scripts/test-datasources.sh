#!/bin/bash

# set -euo pipefail
set -o pipefail

: "${RELEASE:="tobs"}"
: "${NAMESPACE:="default"}"
: "${GRAFANA_USER:="admin"}"

# use curl instead of kubectl to access k8s api. This way we don't need to use container image with kubectl in it.
TOKEN="$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)"
K8S_API_URI="https://${KUBERNETES_SERVICE_HOST}:${KUBERNETES_SERVICE_PORT_HTTPS}/api/v1/namespaces/${NAMESPACE}/secrets/${RELEASE}-grafana"
GRAFANA_PASS="$(
    curl -s \
        --header "Authorization: Bearer ${TOKEN}" \
        --cacert /var/run/secrets/kubernetes.io/serviceaccount/ca.crt \
        "$K8S_API_URI" \
    | jq -r '.data["admin-password"]' \
    | base64 -d
)"

GRAFANA_QUERY_URL="http://${GRAFANA_USER}:${GRAFANA_PASS}@${RELEASE}-grafana.${NAMESPACE}.svc:80/api/ds/query"

function query() {
    local uid="$1"
    local query="$2"
    local format="$3"
    local body

    body=$(cat <<-EOM
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

    curl -H "Content-Type: application/json" -X POST -d "$body" "${GRAFANA_QUERY_URL}" 2>/dev/null | jq '.results.A'
}

SQL_QUERY="SELECT * FROM pg_extension WHERE extname = 'timescaledb_toolkit';"
RESULT_SQL=$(query "c4729dfb8ceeaa0372ef27403a3932695eee995d" "$SQL_QUERY" "table")
if [ "$(jq 'has("error")' <<< "${RESULT_SQL}")" == "true" ]; then
    echo "GRAFANA SQL DATASOURCE CANNOT QUERY DATA DUE TO:"
    jq '.error' <<< "${RESULT_SQL}"
    exit 1
fi

RESULT_PRM=$(query "dc08d25c8f267b054f12002f334e6d3d32a853e4" "ALERTS" "time_series")
if [ "$(jq 'has("error")' <<< "${RESULT_PRM}")" == "true" ]; then
    echo "GRAFANA PROMQL DATASOURCE CANNOT QUERY DATA DUE TO:"
    jq '.error' <<< "${RESULT_PRM}"
    exit 1
fi

echo "All queries passed"
