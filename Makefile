
SHELL := /bin/bash

.EXPORT_ALL_VARIABLES:

GIT_COMMIT     = $(shell git rev-parse --short HEAD)
GIT_REMOTE_URL = $(shell git config --get remote.origin.url)
GITHUB_USER    := $(shell echo $(GITHUB_USER) | sed 's/@/%40/g')
GITHUB_TOKEN   ?=

PROJECT_DIR       = $(shell 'pwd')
BUILD_DIR         = $(PROJECT_DIR)/build
BIN_DIR           = $(PROJECT_DIR)/bin
VENDOR_DIR        = $(PROJECT_DIR)/vendor
I18N_DIR          = $(PROJECT_DIR)/pkg/i18n
DOCKER_BUILD_PATH = $(PROJECT_DIR)/.build-docker

ARCH       ?= $(shell uname -m)
ARCH_TYPE  = $(if $(patsubst x86_64,,$(ARCH)),$(ARCH),amd64)
BUILD_DATE = $(shell date +%m/%d@%H:%M:%S)
VCS_REF    = $(if $(shell git status --porcelain),$(GIT_COMMIT)-$(BUILD_DATE),$(GIT_COMMIT))

CGO_ENABLED = 0
GO111MODULE := off
GOOS        = $(shell go env GOOS)
GOARCH      = $(ARCH_TYPE)
# GOFLAGS=-mod=vendor
GOPACKAGES  = $(shell go list ./... | grep -v /vendor/ | grep -v /internal | grep -v /build | grep -v /test | grep -v /i18n/resources)

## WARNING: OPERATOR IMAGE_DESCRIPTION VAR MUST NOT CONTAIN SPACES.
IMAGE_DESCRIPTION ?= IBM_Multicloud_Operator
DOCKER_FILE       = $(BUILD_DIR)/Dockerfile
DOCKER_REGISTRY   ?= hyc-cloud-private-scratch-docker-local.artifactory.swg-devops.com
DOCKER_NAMESPACE  ?= ibmcom
DOCKER_IMAGE      ?= icp-multicluster-endpoint-operator
DOCKER_BUILD_TAG  ?= latest
DOCKER_TAG        ?= $(shell whoami)
DOCKER_BUILD_OPTS = --build-arg VCS_REF=$(VCS_REF) --build-arg VCS_URL=$(GIT_REMOTE_URL) --build-arg IMAGE_NAME=$(DOCKER_IMAGE) --build-arg IMAGE_DESCRIPTION=$(IMAGE_DESCRIPTION) --build-arg ARCH_TYPE=$(ARCH_TYPE)

BEFORE_SCRIPT := $(shell ./build/before-make-script.sh)

-include $(shell curl -fso .build-harness -H "Authorization: token ${GITHUB_TOKEN}" -H "Accept: application/vnd.github.v3.raw" "https://raw.github.ibm.com/ICP-DevOps/build-harness/master/templates/Makefile.build-harness"; echo .build-harness)


.PHONY: deps
## Download all project dependencies
deps: init
	go get -u github.com/golang/dep/cmd/dep
	go get -u github.com/rws-github/go-swagger/cmd/swagger
	dep ensure -v

.PHONY: check
## Runs a set of required checks
check: %check: %go:check %go:copyright:check
#	@echo "WARNING: i18n is not yet supported by `make check`."

.PHONY: image
## Builds operator binary inside of an image
image::
	@$(BUILD_DIR)/download-kubectl.sh
	$(MAKE) operator:build

.PHONY: clean
## Clean build-harness and remove Go generated build and test files
clean:: %clean: %go:clean
	@[ "$(BUILD_HARNESS_PATH)" == '/' ] || \
	 [ "$(BUILD_HARNESS_PATH)" == '.' ] || \
	   rm -rf $(BUILD_HARNESS_PATH)

.PHONY: helpz
helpz:
ifndef build-harness
	$(eval MAKEFILE_LIST := Makefile build-harness/modules/go/Makefile)
endif

### OPERATOR SDK #######################

.PHONY: operator\:tools
operator\:tools:
	./build/install-operator-sdk.sh

.PHONY: operator\:build
operator\:build: deps
	## WARNING: DOCKER_BUILD_OPTS MUST NOT CONTAIN ANY SPACES.
	$(info Building operator)
	$(info GOOS: $(GOOS))
	$(info GOARCH: $(GOARCH))
	$(info --IMAGE: $(DOCKER_IMAGE))
	$(info --TAG: $(DOCKER_BUILD_TAG))
	$(info --DOCKER_BUILD_OPTS: $(DOCKER_BUILD_OPTS))
	operator-sdk build $(DOCKER_IMAGE):$(DOCKER_BUILD_TAG) --image-build-args "$(DOCKER_BUILD_OPTS)"

.PHONY: operator\:run
operator\:run:
	operator-sdk up local --namespace="" --operator-flags="--zap-devel=true"

### HELPER UTILS #######################

.PHONY: utils\:crds\:install
utils\:crds\:install:
	kubectl apply -f deploy/crds/multicloud_v1beta1_endpoint_crd.yaml

.PHONY: utils\:crds\:uninstall
utils\:crds\:uninstall:
	kubectl delete -f deploy/crds/multicloud_v1beta1_endpoint_crd.yaml
