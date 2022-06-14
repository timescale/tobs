#!/bin/bash

# This script is meant to:
# 0) do nothing if running in GitHub Actions
# 1) discover all images used in tobs helm chart and its dependencies
# 2) download images locally if required
# 3) load images into kind cluster

if [ "$CI" == "true" ]; then
	echo "Running in CI. Skipping auto-loading images to get more realistic startup times."
	exit 0
fi

# Set this after CI detection to prevent "unbound variable" error
set -euo pipefail

cd $(git rev-parse --show-toplevel)

for img in $(helm template chart/ | grep 'image:' | cut -d':' -f2- | tr -d '"' | sort -u); do
	(
		# remove leading and trailing spaces
		img="${img#"${img%%[![:space:]]*}"}"
		echo "Checking for local existence of $img"
		if [[ "$(docker images -q $img 2>/dev/null)" == "" ]]; then
			docker pull $img
		fi
		kind load docker-image "$img"
	) &
done
wait
