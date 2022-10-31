#! /usr/bin/env bash

set -euo pipefail

# Variables
psql_exporter="tmp/postgres_exporter"
psql_mixin="${psql_exporter}/postgres_mixin"
prom="tmp/promscale"
prom_mixin="${prom}/docs/mixin/dashboards"

# This mixin requires a few tools to be installed, we need to check for those
# and print out that they need to be installed if they are not.

if [ ! -x "$(command -v go)" ]; then
	echo "go is not installed, please install it"
	exit 1
fi

if [ ! -x "$(command -v mixtool)" ]; then
	echo "mixtool is not installed, please install it"
	echo "https://github.com/prometheus-community/postgres_exporter/tree/master/postgres_mixin"
	exit 1
fi

# TODO(nhudson) think about adding in jq formatting to make the dashboard
# code look nicer.
#if [ ! -x "$(command -v jq)" ]; then
#	echo "jq is not installed, please install it"
#	exit 1
#fi

git_clone() {
  # Checkout mixins for postgres-exporter and promscale
  cd "$(git rev-parse --show-toplevel)"

  # Remove local tmp directory if exists
  if [ -d "tmp/" ]; then
    rm -fr "tmp/"
  fi

  mkdir -p tmp
  git clone -q --depth 1 https://github.com/prometheus-community/postgres_exporter.git ${psql_exporter}
  git clone -q --depth 1 https://github.com/timescale/promscale.git ${prom}

	if [ -d "${psql_exporter}" ]; then
		echo "postgres-exporter mixin cloned at ${psql_exporter}"
	else
		echo "postgres-exporter mixin clone failed"
		exit 1
	fi

	if [ -d "${prom}" ]; then
		echo "promscale mixin cloned at ${prom}"
	else
		echo "promscale mixin clone failed"
		exit 1
	fi

}


build_psql_exporter() {

  # To build the postgres-exporter mixin alerts and dashboard we need to run
  # the following commands:
  cd ${psql_mixin}
	make build
	if [ $? -ne 0 ]; then
		echo "postgres-exporter mixin build failed"
		exit 1
	fi
	cd -

	# This seems like the most straightforward way to replace the datasource in the
	# generated dashboard for now. We can revisit this if we find a better way to
	# replace the generated datasource.
	for file in ${psql_mixin}/dashboards_out/*.json
	do
		cp -r $file chart/dashboards/$(basename $file)
		sed -i.orig 's/\"datasource\": .*$/\"datasource\": {\n    \"type\": \"prometheus\",\n    \"uid\": \"dc08d25c8f267b054f12002f334e6d3d32a853e4\"\n }, /g' chart/dashboards/$(basename $file)
		find chart/dashboards/ -name '*.orig' -exec rm -f {} \;
	done

	# Set and copy over PrometheusRule alert configuration
	if [ ! -f ${psql_mixin}/alerts.yaml ]; then
		echo "make build failed, alerts.yaml is not found"
		exit 1
	else
		cp ${psql_mixin}/alerts.yaml chart/alerts/postgres-exporter-alerts.yaml
	fi

}

copy_promscale_mixin() {

	# Start copying the promscale mixin alerts and dashboard
	cp -r ${prom_mixin}/*.json chart/dashboards/

	# FIXME(paulfantom): those changes should be incorporated into promscale mixin
	# dashboard UID is a sha256 hash of ".title"
	# datasource UID is a sha256 hash of ".name"
	# "__inputs" field needs to be removed
	# replace all `${DS_TIMESCALEDB}` with timescaledb datasource UID
	# replace all `${DS_PROMSCALE_JAEGER}` with promscale tracing datasource UID
	find chart/dashboards/ \( -type d -name '*.json' -prune \) -o -type f -print0 | xargs -0 sed -i.orig 's/${DS_TIMESCALEDB}/c4729dfb8ceeaa0372ef27403a3932695eee995d/g'
	find chart/dashboards/ -name '*.orig' -exec rm -f {} \;
	find chart/dashboards/ \( -type d -name '*.json' -prune \) -o -type f -print0 | xargs -0 sed -i.orig 's/${DS_PROMSCALE_JAEGER}/f78291126102e0f2e841734d1e90250257543042/g'
	find chart/dashboards/ -name '*.orig' -exec rm -f {} \;

}

git_clone
build_psql_exporter
copy_promscale_mixin

echo ""
echo "Copy of alerts and dashboards is complete."
echo "If you added any new dashboards please make sure you add them in the values.yaml file as well."
