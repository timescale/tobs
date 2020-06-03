#!/bin/bash

if [ -n "$DEBUG" ]; then
	set -x
fi

# Default values.yaml file to provide to helm template command.
FILE_ARG=./chart/values.yaml

# Check if values file filepath was supplied.
if [ "$#" -eq  "1" ]; then
	if [ -f "$1" ]; then
		FILE_ARG=$1
	fi
fi

set -o errexit
set -o nounset
set -o pipefail

DIR=$(cd $(dirname "${BASH_SOURCE}") && pwd -P)

RELEASE_NAME=ts-observability
NAMESPACE=ts-observability

NAMESPACE_VAR="
apiVersion: v1
kind: Namespace
metadata:
  name: $NAMESPACE
  labels:
    app.kubernetes.io/name: $RELEASE_NAME
    app.kubernetes.io/instance: timescale-observability
"
helm dependency update ${DIR}/chart

OUTPUT_FILE="${DIR}/deploy/static/deploy.yaml"
cat << EOF | helm template $RELEASE_NAME ${DIR}/chart --namespace $NAMESPACE --values $FILE_ARG > ${OUTPUT_FILE}
EOF

echo "${NAMESPACE_VAR}
$(cat ${OUTPUT_FILE})" > ${OUTPUT_FILE}
