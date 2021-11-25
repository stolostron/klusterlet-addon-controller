# Copyright Contributors to the Open Cluster Management project


SHELL := /bin/bash

export BUILD_DATE  = $(shell date +%m/%d@%H:%M:%S)

export CGO_ENABLED  = 0
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
export DOCKER_REGISTRY   ?= quay.io/open-cluster-management
export DOCKER_IMAGE      ?= $(COMPONENT_NAME)
export DOCKER_IMAGE_COVERAGE_POSTFIX ?= -coverage
export DOCKER_IMAGE_COVERAGE      ?= $(DOCKER_IMAGE)$(DOCKER_IMAGE_COVERAGE_POSTFIX)
export DOCKER_TAG        ?= latest
export DOCKER_BUILDER    ?= docker

export BINDATA_TEMP_DIR := $(shell mktemp -d)

BEFORE_SCRIPT := $(shell build/before-make.sh)

# Only use git commands if it exists
ifdef GIT
GIT_COMMIT      = $(shell git rev-parse --short HEAD)
GIT_REMOTE_URL  = $(shell git config --get remote.origin.url)
VCS_REF     = $(if $(shell git status --porcelain),$(GIT_COMMIT)-$(BUILD_DATE),$(GIT_COMMIT))
endif

.PHONY: deps
## Download all project dependencies
deps: build/install-dependencies.sh

.PHONY: check
## Runs a set of required checks
check: lint go-bindata-check copyright-check

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
	go build -o build/_output/manager -mod=mod ./cmd/manager

.PHONY: build-image
## Builds controller binary inside of an image
build-image: 
	@$(DOCKER_BUILDER) build -t $(DOCKER_IMAGE) -f $(DOCKER_FILE) . 
	@$(DOCKER_BUILDER) tag $(DOCKER_IMAGE) ${DOCKER_REGISTRY}/${DOCKER_IMAGE}:$(DOCKER_TAG)

.PHONY: build-e2e
build-e2e:
	$(SELF) component/build COMPONENT_TAG_EXTENSION=-e2e COMPONENT_BUILD_COMMAND=$(PWD)/build/build-e2e.sh 
	
.PHONY: go-bindata
go-bindata:
	go-bindata -nometadata -pkg bindata -o pkg/bindata/bindata_generated.go -prefix deploy/ deploy/resources/ deploy/crds/ deploy/crds-v1/ deploy/crds-kube1.11/ deploy/resources/...

.PHONY: gobindata-check
go-bindata-check:
	cd $(mktemp -d) && GO111MODULE=off go get -u github.com/go-bindata/go-bindata/...
	@go-bindata --version
	@go-bindata -nometadata -pkg bindata -o $(BINDATA_TEMP_DIR)/bindata_generated.go -prefix deploy/ deploy/resources/ deploy/crds/ deploy/crds-v1/ deploy/crds-kube1.11/ deploy/resources/...; \
	diff $(BINDATA_TEMP_DIR)/bindata_generated.go pkg/bindata/bindata_generated.go > go-bindata.diff; \
	if [ $$? != 0 ]; then \
	  echo "Run 'make go-bindata' to regenerate the bindata_generated.go"; \
	  cat go-bindata.diff; \
	  exit 1; \
	fi
	rm go-bindata.diff
	@echo "##### go-bindata-check #### Success"

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
	kubectl apply -f deploy/dev-crds/agent.open-cluster-management.io_klusterletaddonconfigs_crd.yaml

.PHONY: utils-crds-uninstall
utils-crds-uninstall:
	kubectl delete -f deploy/dev-crds/agent.open-cluster-management.io_klusterletaddonconfigs_crd.yaml

### FUNCTIONAL TESTS UTILS ###########
.PHONY: deploy
deploy:
	kubectl apply -k overlays/community

.PHONY: functional-test
functional-test: 
	ginkgo -v -tags functional -failFast --slowSpecThreshold=10 test/functional -- --v=1 --image-registry=${COMPONENT_DOCKER_REPO}

.PHONY: build-image-coverage
## Builds controller binary inside of an image
build-image-coverage: build-image
	$(DOCKER_BUILDER) build -f $(DOCKERFILE_COVERAGE) . -t $(DOCKER_IMAGE_COVERAGE) --build-arg DOCKER_BASE_IMAGE=$(DOCKER_IMAGE)

	# @$(DOCKER_BUILDER) build -t ${DOCKER_REGISTRY}/${DOCKER_IMAGE}-coverage -f $(DOCKERFILE_COVERAGE) . 
	# @$(DOCKER_BUILDER) tag ${DOCKER_REGISTRY}/${DOCKER_IMAGE}-coverage ${DOCKER_REGISTRY}/${DOCKER_IMAGE}-coverage:$(DOCKER_TAG)

.PHONY: functional-test-full
functional-test-full: build-image-coverage
	build/run-functional-tests.sh $(DOCKER_IMAGE_COVERAGE)

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
	
