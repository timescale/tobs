#!/bin/bash

set -eu

: "${NAMESPACE:="default"}"
: "${PROMQL_QUERY_URL:="http://tobs-promscale.default.svc:9201/api/v1/query"}"
: "${FEATURE_KUBE_PROMETHEUS:=0}"
: "${FEATURE_TIMESCALEDB:=0}"
: "${FEATURE_PROMSCALE:=0}"

kubePrometheusTests=$(cat <<-EOF
{
  "expression": "alertmanager_build_info{namespace=\"$NAMESPACE\"}",
  "expected": true
},{
  "expression": "node_exporter_build_info{namespace=\"$NAMESPACE\"}",
  "expected": true
},{
  "expression": "prometheus_build_info{namespace=\"$NAMESPACE\"}",
  "expected": true
},{
  "expression": "prometheus_operator_build_info{namespace=\"$NAMESPACE\"}",
  "expected": true
}
EOF
)
promscaleTests=$(cat <<-EOF
{
  "expression": "promscale_build_info{namespace=\"$NAMESPACE\"}",
  "expected": true
}
EOF
)
timescaleTests=$(cat <<-EOF
{
  "expression": "postgres_exporter_build_info{namespace=\"$NAMESPACE\"}",
  "expected": true
}
EOF
)
genericTests=$(cat <<-EOF
{
  "expression": "up{namespace=\"$NAMESPACE\",pod!~".*-grafana-test"}==0",
  "expected": false
}
EOF
)

testset="$genericTests"
if [ "$FEATURE_KUBE_PROMETHEUS" -eq 1 ]; then
  testset="$kubePrometheusTests,$testset"
fi
if [ "$FEATURE_PROMSCALE" -eq 1 ]; then
  testset="$promscaleTests,$testset"
fi
if [ "$FEATURE_TIMESCALEDB" -eq 1 ]; then
  testset="$timescaleTests,$testset"
fi

testset="[ $testset ]"

function query() {
  local expr="$1"
  curl -XPOST -G -H "Content-Type: application/x-www-form-urlencoded"  --data-urlencode "query=${expr}" "${PROMQL_QUERY_URL}" 2>/dev/null | jq '.data.result | length > 0'
}

function singletest() {
  local expr="$1"
  local expected="$2"
  local result

  local attempt=0
  local max_attempts=9
  local timeout=1

  echo "Testing ${expr}"

  while [[ "$attempt" -lt "$max_attempts" ]]; do
    result=$(query "${expr}")
    if [[ "${result}" == "${expected}" ]]; then
      echo "PASSED"
      return
    fi
    attempt=$(( attempt + 1 ))
    timeout=$(( timeout * 2 ))
    echo "RETRYING ${attempt}/${max_attempts}"
    sleep "${timeout}"
  done

  echo "FAILED"
  exit 1
}

for t in $(echo "${testset}" | jq -c '.[]'); do
  expr=$(echo "${t}" | jq -r '.expression')
  expected=$(echo "${t}" | jq -r '.expected')
  singletest "${expr}" "${expected}"
done
