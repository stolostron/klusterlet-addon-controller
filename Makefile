
SHELL := /bin/bash

.EXPORT_ALL_VARIABLES:

GIT_COMMIT      = $(shell git rev-parse --short HEAD)
GIT_REMOTE_URL  = $(shell git config --get remote.origin.url)
GITHUB_USER    := $(shell echo $(GITHUB_USER) | sed 's/@/%40/g')
GITHUB_TOKEN   ?=

PROJECT_DIR       = $(shell 'pwd')
BUILD_DIR         = $(PROJECT_DIR)/build
BIN_DIR           = $(PROJECT_DIR)/bin
VENDOR_DIR        = $(PROJECT_DIR)/vendor
I18N_DIR          = $(PROJECT_DIR)/pkg/i18n
DOCKER_BUILD_PATH = $(PROJECT_DIR)/.build-docker

ARCH       ?= $(shell uname -m)
ARCH_TYPE   = $(if $(patsubst x86_64,,$(ARCH)),$(ARCH),amd64)
BUILD_DATE  = $(shell date +%m/%d@%H:%M:%S)
VCS_REF     = $(if $(shell git status --porcelain),$(GIT_COMMIT)-$(BUILD_DATE),$(GIT_COMMIT))

CGO_ENABLED  = 0
GO111MODULE := on
GOOS         = $(shell go env GOOS)
GOARCH       = $(ARCH_TYPE)
GOPACKAGES   = $(shell go list ./... | grep -v /vendor/ | grep -v /internal | grep -v /build | grep -v /test | grep -v /i18n/resources)

## WARNING: IMAGE_DESCRIPTION & DOCKER_BUILD_OPTS MUST NOT CONTAIN ANY SPACES.
IMAGE_DESCRIPTION ?= Endpoint_Operator
DOCKER_FILE        = $(BUILD_DIR)/Dockerfile
DOCKER_REGISTRY   ?= quay.io
DOCKER_NAMESPACE  ?= open-cluster-management
DOCKER_IMAGE      ?= $(COMPONENT_NAME)
DOCKER_BUILD_TAG  ?= latest
DOCKER_TAG        ?= $(shell whoami)
DOCKER_BUILD_OPTS  = --build-arg "VCS_REF=$(VCS_REF)" \
	--build-arg "VCS_URL=$(GIT_REMOTE_URL)" \
	--build-arg "IMAGE_NAME=$(DOCKER_IMAGE)" \
	--build-arg "IMAGE_DESCRIPTION=$(IMAGE_DESCRIPTION)" \
	--build-arg "IMAGE_VERSION=$(SEMVERSION)" \
	--build-arg "ARCH_TYPE=$(ARCH_TYPE)"

# Use project's own component scripts
COMPONENT_SCRIPTS_PATH = ${BUILD_DIR}

BEFORE_SCRIPT := $(shell ./build/before-make-script.sh)

-include $(shell curl -s -H 'Authorization: token ${GITHUB_TOKEN}' -H 'Accept: application/vnd.github.v4.raw' -L https://api.github.com/repos/open-cluster-management/build-harness-extensions/contents/templates/Makefile.build-harness-bootstrap -o .build-harness-bootstrap; echo .build-harness-bootstrap)

.PHONY: no-op
no-op:

.PHONY: deps
## Download all project dependencies
deps: init component/init
	cd $(shell mktemp -d) && go get -u github.com/open-cluster-management/go-ossc/ossc

.PHONY: check
## Runs a set of required checks
check: lint ossccheck

.PHONY: ossccheck
ossccheck:
	$(GOPATH)/bin/ossc --check

.PHONY: ossc
ossc:
	ossc

.PHONY: lint
## Runs linter against go files
lint:
	golangci-lint run

.PHONY: test
## Runs go unit tests
test: component/test/unit

.PHONY: build
## Builds operator binary inside of an image
build: component/build

.PHONY: clean
## Clean build-harness and remove Go generated build and test files
clean::
	@rm -rf $(BUILD_DIR)/_output
	@[ "$(BUILD_HARNESS_PATH)" == '/' ] || \
	 [ "$(BUILD_HARNESS_PATH)" == '.' ] || \
	   rm -rf $(BUILD_HARNESS_PATH)

.PHONY: run
## Run the operator against the kubeconfig targeted cluster
run:
	operator-sdk up local --namespace="" --operator-flags="--zap-devel=true"

.PHONY: helpz
helpz:
ifndef build-harness
	$(eval MAKEFILE_LIST := Makefile build-harness/modules/go/Makefile)
endif

### HELPER UTILS #######################

.PHONY: utils\:crds\:install
utils\:crds\:install:
	kubectl apply -f deploy/crds/multicloud_v1beta1_endpoint_crd.yaml

.PHONY: utils\:crds\:uninstall
utils\:crds\:uninstall:
	kubectl delete -f deploy/crds/multicloud_v1beta1_endpoint_crd.yaml
