REGISTRY?=icr.io/ai-services-private
IMAGE=caddy
TAG?=v2.8.4
CREDS_ARG := $(if $(and $(REGISTRY_USER),$(REGISTRY_PASSWORD)),--creds="$(REGISTRY_USER):$(REGISTRY_PASSWORD)")

# Image references
IMAGE_TAG := $(REGISTRY)/$(IMAGE):$(TAG)

CONTAINER_BUILDER?=podman
build:
	$(CONTAINER_BUILDER) build -t $(IMAGE_TAG) .
.PHONY: build

push:
	$(CONTAINER_BUILDER) push $(CREDS_ARG) $(IMAGE_TAG)
.PHONY: push

image-tag:
	@echo $(TAG)
.PHONY: image-tag
