KUBE_VERSION ?= 1.23
KIND_CONFIG ?= ./cli/tests/kind-$(KUBE_VERSION).yaml
CERT_MANAGER_VERSION ?= 1.6.1

KUBESCAPE_THRESHOLD=29

MDOX_BIN=mdox
MDOX_VALIDATE_CONFIG?=.mdox.validate.yaml
MD_FILES_TO_FORMAT=$(shell find -type f -name '*.md')

all: docs helm-install

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
	kubectl wait --for=condition=Ready pods --all --all-namespaces --timeout=300s

.PHONY: cert-manager
cert-manager: start-kind
	kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v$(CERT_MANAGER_VERSION)/cert-manager.yaml
	# Give enough time for a cluster to register new Pods
	sleep 7
	# Wait for pods to be up and running
	kubectl wait --timeout=120s --for=condition=ready pod -l app.kubernetes.io/instance=cert-manager -n cert-manager

.PHONY: load-images
load-images:  ## Load images into the local kubernetes kind cluster.
	./scripts/load-images.sh

.PHONY: helm-install
helm-install: cert-manager load-images  ## This is a phony target that is used to install the Tobs Helm chart.
	helm dep up chart/
	helm upgrade --install --wait --timeout 15m test chart/

.PHONY: check-datasources
	./scripts/check-datasources.sh

manifests.yaml:
	helm template --namespace test test chart/ > $@

.PHONY: kubescape
kubescape: manifests.yaml  ## Runs a security analysis on generated manifests - failing if risk score is above threshold percentage 'KUBESCAPE_THRESHOLD'
	kubescape scan --verbose framework -t $(KUBESCAPE_THRESHOLD) nsa manifests.yaml --exceptions 'kubescape-exceptions.json'

help: ## Displays help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-z0-9A-Z_-]+:.*?##/ { printf "  \033[36m%-13s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
