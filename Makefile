# Default registry for image updates
REGISTRY?=icr.io/ai-services-cicd
# Base branch used for git diff in auto-detection mode
BASE_BRANCH?=origin/main

# Bump TAG in Makefiles for components with changed source files
# Usage:
#   make bump-tags                         # Auto-detect changed components
#   make bump-tags BASE_BRANCH=origin/dev  # Use custom base branch
#   make bump-tags COMPONENT=services/chatbot  # Bump a specific component
.PHONY: bump-tags
bump-tags:
	@echo "Bumping Makefile TAGs..."
	@if [ -n "$(COMPONENT)" ]; then \
		python3 hack/bump_makefile_tag.py --component $(COMPONENT); \
	else \
		python3 hack/bump_makefile_tag.py --base-branch $(BASE_BRANCH); \
	fi

# Update image tags in values.yaml files
# Usage:
#   make update-tags                    # Auto-detect modified Makefiles
#   make update-tags SYNC=true          # Sync ALL Makefiles tags with values.yaml files
#   make update-tags REGISTRY=custom    # Use custom registry
.PHONY: update-tags
update-tags:
	@echo "Updating image tags in values.yaml files..."
	@if [ "$(SYNC)" = "true" ]; then \
		python3 hack/update_image_tags.py --sync --registry $(REGISTRY); \
	else \
		python3 hack/update_image_tags.py --registry $(REGISTRY); \
	fi

# Bump TAGs and then sync image tags in values.yaml files
# Usage:
#   make release                        # Auto-detect changed components + sync values.yaml
#   make release REGISTRY=custom        # Use custom registry
.PHONY: release
release: bump-tags update-tags

# Help target
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  bump-tags             - Bump TAG in Makefiles for components with changed source files"
	@echo "  update-tags           - Update image tags in values.yaml files"
	@echo "  release               - Bump TAGs then update image tags in values.yaml (bump-tags + update-tags)"
	@echo ""
	@echo "Variables:"
	@echo "  REGISTRY=<url>        - Container registry URL (default: icr.io/ai-services-cicd)"
	@echo "  BASE_BRANCH=<ref>     - Base branch for git diff (default: origin/main)"
	@echo "  SYNC=true             - Sync ALL Makefiles instead of auto-detecting (update-tags only)"
	@echo "  COMPONENT=<path>      - Bump a specific component only, e.g. services/chatbot (bump-tags only)"
	@echo ""
	@echo "Examples:"
	@echo "  make bump-tags                              # Auto-detect and bump changed components"
	@echo "  make bump-tags COMPONENT=services/chatbot   # Bump a specific component"
	@echo "  make update-tags                            # Auto-detect modified Makefiles"
	@echo "  make update-tags SYNC=true                  # Sync ALL Makefiles tags with values.yaml files"
	@echo "  make update-tags REGISTRY=custom-registry   # Use custom registry"
	@echo "  make release                                # Bump TAGs + update values.yaml (full flow)"
	@echo "  make release REGISTRY=icr.io/custom         # Full flow with custom registry"

# Made with Bob