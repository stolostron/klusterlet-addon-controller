# IBM Confidential
# OCO Source Materials
# 5737-E67
# (C) Copyright IBM Corporation 2018 All Rights Reserved
# The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.

SHELL := /bin/bash

.EXPORT_ALL_VARIABLES:

PROJECT_DIR = $(shell 'pwd')
BUILD_DIR = $(PROJECT_DIR)/build
BIN_DIR = $(PROJECT_DIR)/bin
VENDOR_DIR = $(PROJECT_DIR)/vendor
I18N_DIR = $(PROJECT_DIR)/pkg/i18n

CGO_ENABLED=0
GO111MODULE := off
# GOFLAGS=-mod=vendor
GOPACKAGES=$(shell go list ./... | grep -v /vendor/ | grep -v /internal | grep -v /build | grep -v /test | grep -v /i18n/resources)
GOOS = $(shell go env GOOS)

DOCKER_FILE = build/Dockerfile
DOCKER_BUILD_PATH = .build-docker
DOCKER_IMAGE ?= icp-multicluster-endpoint-operator
## WARNING: OPERATOR IMAGE_DESCRIPTION VAR MUST NOT CONTAIN SPACES.
IMAGE_DESCRIPTION ?= IBM_Multicloud_Operator
DOCKER_REGISTRY ?= hyc-cloud-private-integration-docker-local.artifactory.swg-devops.com
DOCKER_NAMESPACE ?= ibmcom
DOCKER_BUILD_TAG ?= latest
DOCKER_TAG ?= $(shell whoami)
WORKING_CHANGES = $(shell git status --porcelain)
BUILD_DATE = $(shell date +%m/%d@%H:%M:%S)
VCS_REF = $(if $(WORKING_CHANGES),$(GIT_COMMIT)-$(BUILD_DATE),$(GIT_COMMIT))
GIT_REMOTE_URL = "git@github.ibm.com:IBMPrivateCloud/ibm-klusterlet-operator.git"
SWAGGER_API_DIR = "$(DOCKER_BUILD_PATH)/api"

ARCH ?= $(shell uname -m)
ARCH_TYPE = $(ARCH)

ifeq ($(ARCH), x86_64)
	ARCH_TYPE = amd64
endif

ifeq ($(GOOS), darwin)
	OPERATOR_SDK_DOWNLOAD_URL = https://github.com/operator-framework/operator-sdk/releases/download/v0.9.0/operator-sdk-v0.9.0-x86_64-apple-darwin
endif
ifeq ($(GOOS), linux)
	OPERATOR_SDK_DOWNLOAD_URL = https://github.com/operator-framework/operator-sdk/releases/download/v0.9.0/operator-sdk-v0.9.0-x86_64-linux-gnu
endif

BEFORE_SCRIPT := $(shell ./build/before-make-script.sh)

-include $(shell curl -fso .build-harness -H "Authorization: token ${GITHUB_TOKEN}" -H "Accept: application/vnd.github.v3.raw" "https://raw.github.ibm.com/ICP-DevOps/build-harness/master/templates/Makefile.build-harness"; echo .build-harness)

DOCKER_BUILD_OPTS = --build-arg VCS_REF=$(VCS_REF) --build-arg VCS_URL=$(GIT_REMOTE_URL) --build-arg IMAGE_NAME=$(DOCKER_IMAGE) --build-arg IMAGE_DESCRIPTION=$(IMAGE_DESCRIPTION) --build-arg ARCH_TYPE=$(ARCH_TYPE)

.PHONY: deps
## Download all project dependencies
deps: init
	go get -u github.com/golang/dep/cmd/dep
	go get -u github.com/rws-github/go-swagger/cmd/swagger
	dep ensure -v

.PHONY: check
## Runs a set of required checks
# check: %check: %go:check %go:copyright:check %i18n:check
check: %check: %go:copyright:check
	@echo "WARNING: GO API & i18n CHECKING IS NOT BEING RUN. NEED GIT USER ACCESS TO armada-opensource-lib. FIX IN ISSUE IBMPrivateCloud/roadmap#28411"

.PHONY: build
## Build all cmd binary files
build: %build:

.PHONY: image
image:: deps
	$(MAKE) operator:build

.PHONY: test
## Run all project tests
# test: %test: %i18n:resources %go:test
test: %test: 
	@echo "WARNING: TEST NOT BEING RUN. THERE ARE NO TESTS. LET'S ADD SOME TESTS, PLEASE. FIX IN ISSUE IBMPrivateCloud/roadmap#28411"

.PHONY: clean
## Clean build-harness and remove Go generated build and test files
clean:: %clean: %go:clean
	@[ "$(BUILD_HARNESS_PATH)" == '/' ] || \
	 [ "$(BUILD_HARNESS_PATH)" == '.' ] || \
	   rm -rf $(BUILD_HARNESS_PATH)

# ### SWAGGER ###########################

.PHONY: swagger
## Generate swagger documentation
swagger:
	@mkdir -p $(SWAGGER_API_DIR)
	@$(GOPATH)/bin/swagger generate spec -b ./pkg/apis/klusterlet -m -o $(SWAGGER_API_DIR)/swagger.json

.PHONY: swagger\:lint
## Run lint check againt swagger documentation
swagger\:lint:
	@echo "WARNING: API LINT IS NOT BEING RUN. FIX IN ISSUE IBMPrivateCloud/roadmap#28411"
	# #- If the npm install fails because of permissions, do not run the command with sudo, just run:
	# #- sudo chown -R $(whoami) ~/.npm
	# #- sudo chown -R $(whoami) /usr/local/lib/node_modules
	# @$(BUILD_DIR)/install-apilint.sh
	# -@apilint $(SWAGGER_API_DIR)/swagger.json 2>/dev/null | tee $(SWAGGER_API_DIR)/api-lint.log

.PHONY: swagger\:diff
## Run diff check again swagger documentation
swagger\:diff:
	@echo "WARNING: API DIFF IS NOT BEING RUN. FIX IN ISSUE IBMPrivateCloud/roadmap#28411"
	# @echo "Running api-diff ..."
	# @$(BUILD_DIR)/api-diff.sh $(API_API_DIR) 3.2.0


# ### OPERATOR SDK #######################
.PHONY: operator\:tools
operator\:tools: $(GOPATH)/bin/operator-sdk

$(GOPATH)/bin/operator-sdk:
	@curl -Lo $(GOPATH)/bin/operator-sdk $(OPERATOR_SDK_DOWNLOAD_URL) 
	@chmod +x $(GOPATH)/bin/operator-sdk

.PHONY: operator\:build
operator\:build: deps
	## WARNING: DOCKER_BUILD_OPTS MUST NOT CONTAIN ANY SPACES.
	operator-sdk build $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/$(DOCKER_IMAGE):$(DOCKER_TAG) --image-build-args "$(DOCKER_BUILD_OPTS)"

.PHONY: operator\:run
operator\:run:
	operator-sdk up local --namespace="" --operator-flags="--zap-encoder=console"

.PHONY: helpz
helpz:
ifndef build-harness
	$(eval MAKEFILE_LIST := Makefile build-harness/modules/go/Makefile)
endif
