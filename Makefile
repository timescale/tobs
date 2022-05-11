MDOX_BIN=mdox
MDOX_VALIDATE_CONFIG?=.mdox.validate.yaml
MD_FILES_TO_FORMAT=$(shell find -type f -name '*.md')
.PHONY: docs
docs:
	@echo ">> formatting and local/remote links"
	$(MDOX_BIN) fmt --soft-wraps -l --links.validate.config-file=$(MDOX_VALIDATE_CONFIG) $(MD_FILES_TO_FORMAT)

.PHONY: check-docs
check-docs:
	@echo ">> checking formatting and local/remote links"
	$(MDOX_BIN) fmt --soft-wraps --check -l --links.validate.config-file=$(MDOX_VALIDATE_CONFIG) $(MD_FILES_TO_FORMAT)