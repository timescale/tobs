#!/bin/bash

set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

rm -rf tmp/promscale
mkdir -p tmp
git clone --depth 1 https://github.com/timescale/promscale.git tmp/promscale

cp -r tmp/promscale/docs/mixin/dashboards/*.json chart/dashboards/
