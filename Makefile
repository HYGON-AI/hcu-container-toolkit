#
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2026 Hygon Information Technology Co., Ltd.
#

DOCKER   ?= docker
MKDIR    ?= mkdir
DIST_DIR ?= $(CURDIR)/dist

include $(CURDIR)/versions.mk

MODULE := hcu-container-toolkit

# By default run all native docker-based targets
docker-native:
include $(CURDIR)/docker/docker.mk

BUILDIMAGE_TAG ?= golang$(GOLANG_VERSION)
BUILDIMAGE ?= $(IMAGE_NAME)-build:$(BUILDIMAGE_TAG)

CMDS := $(patsubst ./cmd/%/,%,$(sort $(dir $(wildcard ./cmd/*/))))
CMD_TARGETS := $(patsubst %,cmd-%, $(CMDS))

CHECK_TARGETS := lint
MAKE_TARGETS := binaries build check fmt test examples cmds coverage generate licenses vendor check-vendor $(CHECK_TARGETS)

TARGETS := $(MAKE_TARGETS) $(CMD_TARGETS)

DOCKER_TARGETS := $(patsubst %,docker-%, $(TARGETS))
.PHONY: $(TARGETS) $(DOCKER_TARGETS)

ifeq ($(VERSION),)
CLI_VERSION = $(LIB_VERSION)$(if $(LIB_TAG),-$(LIB_TAG))
else
CLI_VERSION = $(VERSION)
endif
CLI_VERSION_PACKAGE = hcu-container-toolkit/internal/info

binaries: cmds
ifneq ($(PREFIX),)
cmd-%: COMMAND_BUILD_OPTIONS = -o $(PREFIX)/$(*)
endif
cmds: $(CMD_TARGETS)

ifneq ($(shell uname),Darwin)
EXTLDFLAGS = -Wl,--export-dynamic -Wl,--unresolved-symbols=ignore-in-object-files
else
EXTLDFLAGS = -Wl,-undefined,dynamic_lookup
endif
$(CMD_TARGETS): cmd-%:
	go build -ldflags "-s -w '-extldflags=$(EXTLDFLAGS)' -X $(CLI_VERSION_PACKAGE).gitCommit=$(GIT_COMMIT) -X $(CLI_VERSION_PACKAGE).version=$(CLI_VERSION)" $(COMMAND_BUILD_OPTIONS) $(MODULE)/cmd/$(*)

build:
	go build ./...

all: check test build binary
check: $(CHECK_TARGETS)

# Apply go fmt to the codebase
fmt:
	go list -f '{{.Dir}}' $(MODULE)/... \
		| xargs gofmt -s -l -w

# Apply goimports -local container-toolkit to the codebase
goimports:
	go list -f {{.Dir}} $(MODULE)/... \
		| xargs goimports -local $(MODULE) -w

lint:
	golangci-lint run ./...

vendor:
	go mod tidy
	go mod vendor
	go mod verify

check-vendor: vendor
	git diff --quiet HEAD -- go.mod go.sum vendor

licenses:
	go-licenses csv $(MODULE)/...

COVERAGE_FILE := coverage.out
test: build cmds
	go test -coverprofile=$(COVERAGE_FILE) $(MODULE)/...

coverage: test
	cat $(COVERAGE_FILE) | grep -v "_mock.go" > $(COVERAGE_FILE).no-mocks
	go tool cover -func=$(COVERAGE_FILE).no-mocks

generate:
	go generate $(MODULE)/...

