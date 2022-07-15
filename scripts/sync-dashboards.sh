#!/bin/bash

# TODO(paulfantom): consider using jsonnet for modifications from this script

set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

rm -rf tmp/promscale
mkdir -p tmp
git clone --depth 1 https://github.com/timescale/promscale.git tmp/promscale

cp -r tmp/promscale/docs/mixin/dashboards/*.json chart/dashboards/

# FIXME(paulfantom): those changes should be incorporated into promscale mixin
# dashboard UID is a sha256 hash of ".title"
# datasource UID is a sha256 hash of ".name"
# "__inputs" field needs to be removed
# replace all `${DS_TIMESCALEDB}` with timescaledb datasource UID
# replace all `${DS_PROMSCALE_JAEGER}` with promscale tracing datasource UID
# replace all `${DS_PROMETHEUS}` with promscale prometheus datasource UID
find chart/dashboards/ \( -type d -name '*.json' -prune \) -o -type f -print0 | xargs -0 sed -i.orig 's/${DS_TIMESCALEDB}/c4729dfb8ceeaa0372ef27403a3932695eee995d/g'
find chart/dashboards/ -name '*.orig' -exec rm -f {} \;
find chart/dashboards/ \( -type d -name '*.json' -prune \) -o -type f -print0 | xargs -0 sed -i.orig 's/${DS_PROMSCALE_JAEGER}/f78291126102e0f2e841734d1e90250257543042/g'
find chart/dashboards/ -name '*.orig' -exec rm -f {} \;
find chart/dashboards/ \( -type d -name '*.json' -prune \) -o -type f -print0 | xargs -0 sed -i.orig 's/${DS_PROMETHEUS}/dc08d25c8f267b054f12002f334e6d3d32a853e4/g'
find chart/dashboards/ -name '*.orig' -exec rm -f {} \;
