#!/bin/bash

: "${NAMESPACE:="default"}"
: "${DB_SECRET_NAME:="tobs-credentials"}"
: "${PROMSCALE_SECRET_NAME:="tobs-promscale-connection"}"

while ! kubectl get secret "${DB_SECRET_NAME}" --namespace "${NAMESPACE}"; do
    echo "Waiting for ${DB_SECRET_NAME} secret."
    sleep 1
done
PASS="$(kubectl get secret --namespace "${NAMESPACE}" "${DB_SECRET_NAME}" -o json | jq -r '.data["PATRONI_SUPERUSER_PASSWORD"]')"

kubectl get secret --namespace "${NAMESPACE}" "${PROMSCALE_SECRET_NAME}" -o json | jq --arg PASS "$PASS" '.data["PROMSCALE_DB_PASSWORD"]=$PASS' | kubectl apply -f -
