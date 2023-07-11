# Copyright Contributors to the Open Cluster Management project


SHELL := /bin/bash

export BUILD_DATE  = $(shell date +%m/%d@%H:%M:%S)

export CGO_ENABLED  = 1
export GOFLAGS ?= 
export GO111MODULE := on
export GOOS         = $(shell go env GOOS)
export GOPACKAGES   = $(shell go list ./... | grep -v /vendor | grep -v /build | grep -v /test)

export PROJECT_DIR            = $(shell 'pwd')
export BUILD_DIR              = $(PROJECT_DIR)/build
export COMPONENT_SCRIPTS_PATH = $(BUILD_DIR)

export COMPONENT_NAME ?= $(shell cat ./COMPONENT_NAME 2> /dev/null)
export COMPONENT_VERSION ?= $(shell cat ./COMPONENT_VERSION 2> /dev/null)

export DOCKER_FILE        = $(BUILD_DIR)/Dockerfile
export DOCKERFILE_COVERAGE = $(BUILD_DIR)/Dockerfile-coverage
export DOCKER_REGISTRY   ?= quay.io/stolostron
export DOCKER_IMAGE      ?= $(COMPONENT_NAME)
export DOCKER_IMAGE_COVERAGE_POSTFIX ?= -coverage
export DOCKER_IMAGE_COVERAGE      ?= $(DOCKER_IMAGE)$(DOCKER_IMAGE_COVERAGE_POSTFIX)
export DOCKER_TAG        ?= latest
export DOCKER_BUILDER    ?= docker

BEFORE_SCRIPT := $(shell build/before-make.sh)

# Only use git commands if it exists
ifdef GIT
GIT_COMMIT      = $(shell git rev-parse --short HEAD)
GIT_REMOTE_URL  = $(shell git config --get remote.origin.url)
VCS_REF     = $(if $(shell git status --porcelain),$(GIT_COMMIT)-$(BUILD_DATE),$(GIT_COMMIT))
endif

KUBECONFIG ?= ./.kubeconfig
KUBECTL?=kubectl

.PHONY: deps
## Download all project dependencies
deps:
	build/install-dependencies.sh

.PHONY: check
## Runs a set of required checks
check: lint

.PHONY: copyright-check
copyright-check:
	@build/copyright-check.sh


.PHONY: test
## Runs go unit tests
test:
	@build/run-unit-tests.sh

.PHONY: build
## Builds operator binary inside of an image
build:
	go build ./cmd/manager
	go test -covermode=atomic -coverpkg=github.com/stolostron/klusterlet-addon-controller/pkg/... -c -tags testrunmain ./cmd/manager -o manager-coverage

.PHONY: build-image
## Builds controller binary inside of an image
build-image:
	@$(DOCKER_BUILDER) build -t $(DOCKER_IMAGE) -f $(DOCKER_FILE) .
	echo "${DOCKER_REGISTRY}/${DOCKER_IMAGE}:$(DOCKER_TAG)"
	@$(DOCKER_BUILDER) tag $(DOCKER_IMAGE) ${DOCKER_REGISTRY}/${DOCKER_IMAGE}:$(DOCKER_TAG)


.PHONY: go-mod-check
go-mod-check:
	./build/go-mod-check.sh

.PHONY: clean
## Cleanremove Go generated build and test files
clean::
	@rm -rf $(BUILD_DIR)/_output

.PHONY: run
## Run the operator against the kubeconfig targeted cluster
run:
	operator-sdk run local --watch-namespace=""

.PHONE: request-destruct
request-destruct:
	build/bin/self-destruct.sh

.PHONY: lint-all
lint-all:
	@echo "Running linting tool ..."
	@golangci-lint run --timeout 5m --build-tags e2e,functional

.PHONY: lint
## Runs linter against go files
lint:
	build/run-lint-check.sh

### HELPER UTILS #######################

.PHONY: utils-crds-install
utils-crds-install:
	$(KUBECTL) apply -f deploy/dev-crds/agent.open-cluster-management.io_klusterletaddonconfigs_crd.yaml

.PHONY: utils-crds-uninstall
utils-crds-uninstall:
	$(KUBECTL) delete -f deploy/dev-crds/agent.open-cluster-management.io_klusterletaddonconfigs_crd.yaml

### FUNCTIONAL TESTS UTILS ###########

.PHONY: deploy
deploy:
	$(KUBECTL) apply -k deploy

.PHONY: functional-test
functional-test:
	ginkgo -v -tags functional -failFast --slowSpecThreshold=10 test/functional -- --v=1 --image-registry=${COMPONENT_DOCKER_REPO}

.PHONY: build-image-coverage
## Builds controller binary inside of an image
build-image-coverage: build-image
	$(DOCKER_BUILDER) build -f $(DOCKERFILE_COVERAGE) . -t $(DOCKER_IMAGE_COVERAGE) --build-arg DOCKER_BASE_IMAGE=$(DOCKER_IMAGE)

	# @$(DOCKER_BUILDER) build -t ${DOCKER_REGISTRY}/${DOCKER_IMAGE}-coverage -f $(DOCKERFILE_COVERAGE) .
	# @$(DOCKER_BUILDER) tag ${DOCKER_REGISTRY}/${DOCKER_IMAGE}-coverage ${DOCKER_REGISTRY}/${DOCKER_IMAGE}-coverage:$(DOCKER_TAG)

# download script for coverage entrypoint.
.PHONY: sync-coverage-entrypoint
sync-coverage-entrypoint:
	@echo downloading coverage entrypoint file
	@tmp_dir=$$(mktemp -d); \
	curl  --fail -H 'Accept: application/vnd.github.v4.raw' -L https://api.github.com/repos/open-cluster-management/build-harness-extensions/contents/modules/component/bin/component/coverage-entrypoint-func.sh > "$$tmp_dir/coverage-entrypoint-func.sh" \
	&& mv "$$tmp_dir/coverage-entrypoint-func.sh" build/bin/ && chmod +x build/bin/coverage-entrypoint-func.sh ;

.PHONY: build-coverage
## Builds operator binary inside of an image
build-coverage:
	build/build-coverage.sh ${COMPONENT_DOCKER_REPO}/${COMPONENT_NAME}:${COMPONENT_VERSION}${COMPONENT_TAG_EXTENSION}-coverage

# Ensure controller-gen
ensure-controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.5.0 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

# Generate crds
manifests: ensure-controller-gen
	$(CONTROLLER_GEN) "crd:crdVersions=v1" paths="./pkg/apis/agent/v1" output:crd:artifacts:config=deploy/
	mv deploy/agent.open-cluster-management.io_klusterletaddonconfigs.yaml deploy/agent.open-cluster-management.io_klusterletaddonconfigs_crd.yaml

# Generate deepcopy
generate: ensure-controller-gen
	$(CONTROLLER_GEN) "object" paths="./pkg/apis/agent/v1" output:dir="./pkg/apis/agent/v1"

# e2e test
.PHONY: prepare-e2e-cluster
prepare-e2e-cluster:
	echo $(KUBECONFIG)
	build/e2e/install-e2e-cluster.sh

.PHONY: build-e2e
build-e2e:
	go test -c ./test/e2e -mod=mod

.PHONY: test-e2e
test-e2e: build-e2e prepare-e2e-cluster deploy
	./e2e.test -test.v -ginkgo.v
