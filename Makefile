KUBE_VERSION ?= 1.23
KIND_CONFIG ?= ./cli/tests/kind-$(KUBE_VERSION).yaml
CERT_MANAGER_VERSION ?= 1.6.1

MDOX_BIN=mdox
MDOX_VALIDATE_CONFIG?=.mdox.validate.yaml
MD_FILES_TO_FORMAT=$(shell find -type f -name '*.md')

all: docs helm-install

.PHONY: docs
docs:
	@echo ">> formatting and local/remote links"
	$(MDOX_BIN) fmt --soft-wraps -l --links.validate.config-file=$(MDOX_VALIDATE_CONFIG) $(MD_FILES_TO_FORMAT)

.PHONY: check-docs
check-docs:
	@echo ">> checking formatting and local/remote links"
	$(MDOX_BIN) fmt --soft-wraps --check -l --links.validate.config-file=$(MDOX_VALIDATE_CONFIG) $(MD_FILES_TO_FORMAT)

.PHONY: delete-kind
delete-kind:
	kind delete cluster && sleep 10

.PHONY: start-kind
start-kind: delete-kind
	kind create cluster --config $(KIND_CONFIG)
	kubectl wait --for=condition=Ready pods --all --all-namespaces --timeout=300s

.PHONY: cert-manager
cert-manager: start-kind
	kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v$(CERT_MANAGER_VERSION)/cert-manager.yaml
	# Give enough time for a cluster to register new Pods
	sleep 7
	# Wait for pods to be up and running
	kubectl wait --for=condition=ready pod -l app.kubernetes.io/instance=cert-manager -n cert-manager

.PHONY: load-images
load-images:
	./scripts/load-images.sh

.PHONY: helm-install
helm-install: cert-manager load-images
	helm dep up chart/
	helm upgrade --install --wait --timeout 15m test chart/

.PHONY: check-datasources
	./scripts/check-datasources.sh