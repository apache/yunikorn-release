#
# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements.  See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership.  The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License.  You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

.PHONY: lint check_scripts license-check helm_lint
.PHONY: perf-tools tools clean distclean

# Check if this GO tools version used is at least the version of go specified in
# the go.mod file. The version in go.mod should be in sync with other repos.

# Go compiler selection
ifeq ($(GO),)
GO := go
endif

GO_VERSION := $(shell "$(GO)" version | awk '{print substr($$3, 3, 4)}')
MOD_VERSION := $(shell cat .go_version)

GM := $(word 1,$(subst ., ,$(GO_VERSION)))
MM := $(word 1,$(subst ., ,$(MOD_VERSION)))
FAIL := $(shell if [ $(GM) -lt $(MM) ]; then echo MAJOR; fi)
ifdef FAIL
$(error Build should be run with at least go $(MOD_VERSION) or later, found $(GO_VERSION))
endif
GM := $(word 2,$(subst ., ,$(GO_VERSION)))
MM := $(word 2,$(subst ., ,$(MOD_VERSION)))
FAIL := $(shell if [ $(GM) -lt $(MM) ]; then echo MINOR; fi)
ifdef FAIL
$(error Build should be run with at least go $(MOD_VERSION) or later, found $(GO_VERSION))
endif

# Make sure we are in the same directory as the Makefile
BASE_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
TOOLS_DIR := tools
BUILD_DIR := build

# Force Go modules even when checked out inside GOPATH
GO111MODULE := on
export GO111MODULE

REPO=github.com/apache/yunikorn-core/pkg
# when using the -race option CGO_ENABLED is set to 1 (automatically)
# it breaks cross compilation.
RACE=-race
# build commands on local os by default, uncomment for cross-compilation
#GOOS=darwin
#GOARCH=amd64

ifeq ($(HOST_ARCH),)
HOST_ARCH := $(shell uname -m)
endif

# Kernel (OS) Name
OS := $(shell uname -s | tr '[:upper:]' '[:lower:]')

# Allow architecture to be overwritten
ifeq ($(HOST_ARCH),)
HOST_ARCH := $(shell uname -m)
endif

# Build architecture settings:
# EXEC_ARCH defines the architecture of the executables that gets compiled
ifeq (x86_64, $(HOST_ARCH))
EXEC_ARCH := amd64
else ifeq (i386, $(HOST_ARCH))
EXEC_ARCH := 386
else ifneq (,$(filter $(HOST_ARCH), arm64 aarch64))
EXEC_ARCH := arm64
else ifeq (armv7l, $(HOST_ARCH))
EXEC_ARCH := arm
else
$(info Unknown architecture "${HOST_ARCH}" defaulting to: amd64)
EXEC_ARCH := amd64
endif

# shellcheck
SHELLCHECK_VERSION=v0.9.0
SHELLCHECK_PATH=${TOOLS_DIR}/shellcheck-$(SHELLCHECK_VERSION)
SHELLCHECK_BIN=${SHELLCHECK_PATH}/shellcheck
SHELLCHECK_ARCHIVE := shellcheck-$(SHELLCHECK_VERSION).$(OS).$(HOST_ARCH).tar.xz
ifeq (darwin, $(OS))
ifeq (arm64, $(HOST_ARCH))
SHELLCHECK_ARCHIVE := shellcheck-$(SHELLCHECK_VERSION).$(OS).x86_64.tar.xz
endif
else ifeq (linux, $(OS))
ifeq (armv7l, $(HOST_ARCH))
SHELLCHECK_ARCHIVE := shellcheck-$(SHELLCHECK_VERSION).$(OS).armv6hf.tar.xz
endif
endif

# golangci-lint
GOLANGCI_LINT_VERSION=2.10.1
GOLANGCI_LINT_PATH=$(TOOLS_DIR)/golangci-lint-v$(GOLANGCI_LINT_VERSION)
GOLANGCI_LINT_BIN=$(GOLANGCI_LINT_PATH)/golangci-lint
GOLANGCI_LINT_ARCHIVEBASE=golangci-lint-$(GOLANGCI_LINT_VERSION)-$(OS)-$(EXEC_ARCH)
GOLANGCI_LINT_ARCHIVE=$(GOLANGCI_LINT_ARCHIVEBASE).tar.gz

# helm
HELM_VERSION=v3.12.1
HELM_PATH=$(TOOLS_DIR)/helm-$(HELM_VERSION)
HELM_BIN=$(HELM_PATH)/helm
HELM_ARCHIVE=helm-$(HELM_VERSION)-$(OS)-$(EXEC_ARCH).tar.gz
HELM_ARCHIVE_BASE=$(OS)-$(EXEC_ARCH)

all:
	$(MAKE) -C $(dir $(BASE_DIR)) test_all

test_all: license-check check_scripts lint helm_lint

# Install tools
tools: $(SHELLCHECK_BIN) $(GOLANGCI_LINT_BIN) $(HELM_BIN)

# Install shellcheck
$(SHELLCHECK_BIN):
	@echo "installing shellcheck $(SHELLCHECK_VERSION)"
	@mkdir -p "$(SHELLCHECK_PATH)"
	@curl -sSfL "https://github.com/koalaman/shellcheck/releases/download/$(SHELLCHECK_VERSION)/$(SHELLCHECK_ARCHIVE)" \
		| tar -x -J --strip-components=1 -C "$(SHELLCHECK_PATH)" "shellcheck-$(SHELLCHECK_VERSION)/shellcheck"

# Install helm
$(HELM_BIN):
	@echo "installing helm $(HELM_VERSION)"
	@mkdir -p "$(HELM_PATH)"
	@curl -sSfL "https://get.helm.sh/$(HELM_ARCHIVE)" \
		| tar -x -z --strip-components=1 -C "$(HELM_PATH)" "$(HELM_ARCHIVE_BASE)/helm"

# Install golangci-lint
$(GOLANGCI_LINT_BIN):
	@echo "installing golangci-lint v$(GOLANGCI_LINT_VERSION)"
	@mkdir -p "$(GOLANGCI_LINT_PATH)"
	@curl -sSfL "https://github.com/golangci/golangci-lint/releases/download/v$(GOLANGCI_LINT_VERSION)/$(GOLANGCI_LINT_ARCHIVE)" \
		| tar -x -z --strip-components=1 -C "$(GOLANGCI_LINT_PATH)" "$(GOLANGCI_LINT_ARCHIVEBASE)/golangci-lint"

# Run lint against the previous commit for PR and branch build
# In dev setup look at all changes on top of master
lint: $(GOLANGCI_LINT_BIN)
	@echo "running golangci-lint"
	@"${GOLANGCI_LINT_BIN}" run

CHART_DIR := helm-charts/yunikorn
helm_lint: $(HELM_BIN)
	@echo "running helm lint"
	@"$(HELM_BIN)" lint "${CHART_DIR}" -f "${CHART_DIR}/values.yaml"

# Check scripts
ALLSCRIPTS := $(shell find . -not \( -path ./"${TOOLS_DIR}" -prune \) -not \( -path ./"${BUILD_DIR}" -prune \) -name '*.sh')
check_scripts: $(SHELLCHECK_BIN)
	@echo "running shellcheck"
	@"$(SHELLCHECK_BIN)" ${ALLSCRIPTS}

# This is a bit convoluted but using a recursive grep on linux fails to write anything when run
# from the Makefile. That caused the pull-request license check run from the github action to
# always pass. The syntax for find is slightly different too but that at least works in a similar
# way on both Mac and Linux. Excluding all .git* files from the checks.
LICENSE_CHECK_OUT := $(BUILD_DIR)/license-check.txt
license-check:
	@echo "checking license headers:"
ifeq (darwin,$(OS))
	$(shell mkdir -p "${BUILD_DIR}" && find -E . -not \( -path './.git*' -prune \) -not \( -path ./"${BUILD_DIR}" -prune \) -not \( -path ./"${TOOLS_DIR}" -prune \) -regex ".*\.(go|sh|md|yaml|yml|mod)" -exec grep -L "Licensed to the Apache Software Foundation" {} \; > "${LICENSE_CHECK_OUT}")
else
	$(shell mkdir -p "${BUILD_DIR}" && find . -not \( -path './.git*' -prune \) -not \( -path ./"${BUILD_DIR}" -prune \) -not \( -path ./"${TOOLS_DIR}" -prune \) -regex ".*\.\(go\|sh\|md\|yaml\|yml\|mod\)" -exec grep -L "Licensed to the Apache Software Foundation" {} \; > "${LICENSE_CHECK_OUT}")
endif
	@if [ -s "${LICENSE_CHECK_OUT}" ]; then \
		echo "following files are missing license header:" ; \
		cat "${LICENSE_CHECK_OUT}" ; \
		exit 1; \
	fi
	@echo "  all OK"

perf-tools:
	@echo "Running perf-tools"
	@cd perf-tools && make build
	@cd ../

# Remove generated build artifacts
clean:
	@echo "cleaning up caches and output"
	"$(GO)" clean -cache -testcache -r
	@echo "removing generated files"
	@rm -rf "${BUILD_DIR}"

# Remove all generated content
distclean: clean
	@echo "removing tools"
	@rm -rf "${TOOLS_DIR}"
