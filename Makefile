# Default registry for image updates
REGISTRY?=icr.io/ai-services-cicd
# Base branch used for git diff in auto-detection mode
BASE_BRANCH?=origin/main

# Bump TAG in Makefiles for components with changed source files
# Usage:
#   make bump-makefile-tags                         # Auto-detect changed components
#   make bump-makefile-tags BASE_BRANCH=origin/dev  # Use custom base branch
#   make bump-makefile-tags COMPONENT=services/chatbot  # Bump a specific component
.PHONY: bump-makefile-tags
bump-makefile-tags:
	@echo "Bumping Makefile TAGs..."
	@if [ -n "$(COMPONENT)" ]; then \
		python3 hack/bump_makefile_tag.py --component $(COMPONENT); \
	else \
		python3 hack/bump_makefile_tag.py --base-branch $(BASE_BRANCH); \
	fi

# Update image tags in values.yaml files
# Usage:
#   make bump-values-yaml                    # Auto-detect modified Makefiles
#   make bump-values-yaml SYNC=true          # Sync ALL Makefiles tags with values.yaml files
#   make bump-values-yaml REGISTRY=custom    # Use custom registry
.PHONY: bump-values-yaml
bump-values-yaml:
	@echo "Updating image tags in values.yaml files..."
	@if [ "$(SYNC)" = "true" ]; then \
		python3 hack/update_image_tags.py --sync --registry $(REGISTRY); \
	else \
		python3 hack/update_image_tags.py --registry $(REGISTRY); \
	fi

# Bump TAGs and then sync image tags in values.yaml files
# Usage:
#   make bump-up-image-version                        # Auto-detect changed components + sync values.yaml
#   make bump-up-image-version REGISTRY=custom        # Use custom registry
.PHONY: bump-up-image-version
bump-up-image-version: bump-makefile-tags bump-values-yaml

# Help target
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  bump-makefile-tags             - Bump TAG in Makefiles for components with changed source files"
	@echo "  bump-values-yaml               - Update image tags in values.yaml files"
	@echo "  bump-up-image-version          - Bump TAGs then update image tags in values.yaml (bump-makefile-tags + bump-values-yaml)"
	@echo ""
	@echo "Variables:"
	@echo "  REGISTRY=<url>        - Container registry URL (default: icr.io/ai-services-cicd)"
	@echo "  BASE_BRANCH=<ref>     - Base branch for git diff (default: origin/main)"
	@echo "  SYNC=true             - Sync ALL Makefiles instead of auto-detecting (bump-values-yaml only)"
	@echo "  COMPONENT=<path>      - Bump a specific component only, e.g. services/chatbot (bump-makefile-tags only)"
	@echo ""
	@echo "Examples:"
	@echo "  make bump-makefile-tags                              # Auto-detect and bump changed components"
	@echo "  make bump-makefile-tags COMPONENT=services/chatbot   # Bump a specific component"
	@echo "  make bump-values-yaml                                # Auto-detect modified Makefiles"
	@echo "  make bump-values-yaml SYNC=true                      # Sync ALL Makefiles tags with values.yaml files"
	@echo "  make bump-values-yaml REGISTRY=custom-registry       # Use custom registry"
	@echo "  make bump-up-image-version                           # Bump TAGs + update values.yaml (full flow)"
	@echo "  make bump-up-image-version REGISTRY=icr.io/custom    # Full flow with custom registry"

# Made with Bob
