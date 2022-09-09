OS ?= $(shell uname)
ARCH ?= $(shell uname -m)

GOOS ?= $(shell echo "$(OS)" | tr '[:upper:]' '[:lower:]')
GOARCH_x86_64 = amd64
GOARCH_aarch64 = arm64
GOARCH_arm64 = arm64
GOARCH ?= $(shell echo "$(GOARCH_$(ARCH))")

REVISION := dev.$(shell echo $(CIRCLE_SHA1) | head -c 8)

OUTPUT_DIR := ./
OUTPUT_BIN := sidecar-mutatingwebhook-init-container

.PHONY: build
build:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(OUTPUT_DIR)/$(OUTPUT_BIN)
	chmod +x $(OUTPUT_DIR)/$(OUTPUT_BIN)
	docker build -t docker.io/twdps/sidecar-mutatingwebhook-init-container:$(REVISION) .

.PHONY: push
push:
	docker push docker.io/twdps/sidecar-mutatingwebhook-init-container:$(REVISION)