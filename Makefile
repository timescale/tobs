KUBE_VERSION ?= 1.24
KIND_CONFIG ?= ./testdata/kind-$(KUBE_VERSION).yaml
CERT_MANAGER_VERSION ?= v1.11.0

KUBESCAPE_THRESHOLD=31

MDOX_BIN=mdox
MDOX_VALIDATE_CONFIG?=.mdox.validate.yaml
MD_FILES_TO_FORMAT=$(shell find -type f -name '*.md')

all: docs helm-install

.PHONY: clean
clean:  ## Remove artifacts from previous installations
	rm -rf chart/charts
	rm chart/Chart.lock

.PHONY: docs
docs:  ## This is a phony target that is used to force the docs to be generated.
	@echo ">> formatting and local/remote links"
	$(MDOX_BIN) fmt --soft-wraps -l --links.validate.config-file=$(MDOX_VALIDATE_CONFIG) $(MD_FILES_TO_FORMAT)

.PHONY: check-docs
check-docs:
	@echo ">> checking formatting and local/remote links"
	$(MDOX_BIN) fmt --soft-wraps --check -l --links.validate.config-file=$(MDOX_VALIDATE_CONFIG) $(MD_FILES_TO_FORMAT)

.PHONY: delete-kind
delete-kind:  ## This is a phony target that is used to delete the local kubernetes kind cluster.
	kind delete cluster && sleep 10

.PHONY: start-kind
start-kind: delete-kind  ## This is a phony target that is used to create a local kubernetes kind cluster.
	kind create cluster --config $(KIND_CONFIG)
	sleep 10  # Wait for additional objects to be present
	kubectl wait pods --for=condition=Ready --timeout=300s --all --all-namespaces

.PHONY: cert-manager
cert-manager:
	kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/$(CERT_MANAGER_VERSION)/cert-manager.yaml
	# Give enough time for a cluster to register new Pods
	sleep 7
	# Wait for pods to be up and running
	kubectl wait --timeout=120s --for=condition=ready pod -l app.kubernetes.io/instance=cert-manager -n cert-manager

.PHONY: load-images
load-images:  ## Load images into the local kubernetes kind cluster.
	./scripts/load-images.sh

.PHONY: helm-install
helm-install: start-kind cert-manager load-images  ## This is a phony target that is used to install the Tobs Helm chart.
	helm dep up chart/
	helm install --wait --timeout 15m test chart/ --values chart/values.yaml --values chart/ci/default-values.yaml

.PHONY: helm-upgrade
helm-upgrade: cert-manager
	helm dep up chart/
	helm upgrade --wait --timeout 15m test chart/

.PHONY: lint
lint:  ## Lint helm chart using ct (chart-testing).
	ct lint --config ct.yaml

.PHONY: shellcheck
shellcheck: ## Lint shell scripts using locally installed shellcheck.
	for f in $$(find chart/scripts/ -name "*.sh" -type f); do \
		shellcheck $$f  ;\
	done
	for f in $$(find scripts/ -name "*.sh" -type f) $$(find chart/scripts/ -name "*.sh" -type f); do \
		shellcheck --severity=error $$f  ;\
	done


.PHONY: timescaledb
timescaledb: ## This is a phony target that is used to install the timescaledb-single chart.
	kubectl create ns timescaledb
	-helm repo add timescaledb 'https://charts.timescale.com'
	helm repo update timescaledb
	helm install test --wait --timeout 15m \
		timescaledb/timescaledb-single \
		--namespace=timescaledb \
		--set replicaCount=1 \
		--set secrets.credentials.PATRONI_SUPERUSER_PASSWORD="temporarypassword" \
		--set secrets.credentials.PATRONI_admin_PASSWORD="temporarypassword" \
		--set patroni.log.level=INFO

.PHONY: e2e
e2e:  ## Run e2e installation tests using ct (chart-testing).
	ct install --config ct.yaml

manifests.yaml:
	helm template --namespace test test chart/ > $@

.PHONY: kubescape
kubescape: manifests.yaml  ## Runs a security analysis on generated manifests - failing if risk score is above threshold percentage 'KUBESCAPE_THRESHOLD'
	kubescape scan --verbose framework -t $(KUBESCAPE_THRESHOLD) nsa manifests.yaml --exceptions 'kubescape-exceptions.json'

help: ## Displays help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-z0-9A-Z_-]+:.*?##/ { printf "  \033[36m%-13s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

.PHONY: sync-mixins
sync-mixins:  ## Syncs mixins from Promscale and Postgres-Exporter
	./scripts/sync-mixins.sh
